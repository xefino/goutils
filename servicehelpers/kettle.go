package servicehelpers

import (
	"context"
	"fmt"
	"time"

	kettle2 "github.com/flowerinthenight/kettle/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/xefino/goutils/concurrency"
	"github.com/xefino/goutils/utils"
)

// Keyer contains the data necessary to ensure that a given Kettle operation with the suffix runs at
// the desired frequency
type Keyer struct {
	Suffix    string
	Frequency time.Duration
	Operation KettleFunction
}

// NewKeyer creates a new Keyer from the suffix, frequency and kettle function
func NewKeyer(suffix string, freq time.Duration, operation KettleFunction) *Keyer {
	return &Keyer{
		Suffix:    suffix,
		Frequency: freq,
		Operation: operation,
	}
}

// MasterFunction defines the type of function that will act as the
// master when Kettle elects this instance of a service as the master
type KettleFunction func(context.Context) error

// KettleContext contains data that should be made available to the
// master function when it is instantiated from Kettle
type KettleContext struct {
	cancel      context.CancelFunc
	done        chan error
	Name        string
	Environment string
	Instance    *kettle2.Kettle
	Logger      *utils.Logger
}

// NewKettleContext creates a new Kettle context that can be used to run
// a master function across multiple services
func NewKettleContext(serviceName string, environment string, expirationSeconds int,
	logger *utils.Logger) (*KettleContext, error) {

	// Attempt to create a new Kettle with the service name; if this fails then return an error
	k, err := kettle2.New(
		kettle2.WithName(serviceName),
		kettle2.WithNodeName(serviceName),
		kettle2.WithTickTime(int64(expirationSeconds)),
		kettle2.WithVerbose(true))
	if err != nil {
		return nil, logger.Error(err, "Failed to create new Kettle instance")
	}

	// Returns a new Kettle context with the service name, environment and Kettle instance
	return &KettleContext{
		Name:        serviceName,
		Environment: environment,
		Instance:    k,
		Logger:      logger,
	}, nil
}

// Run begins a master-slave operation where the master is automatically run
// by Kettle and the slave is operated manually
func (kCtx *KettleContext) Run(ctx context.Context, slave KettleFunction,
	keyers ...*Keyer) error {

	// First, create a new context with a cancellation function and set
	// it on the context
	ctx, cancel := context.WithCancel(ctx)
	kCtx.cancel = cancel
	kCtx.done = make(chan error, 1)

	// Next, create a new start-input with this object as the context and
	// a function that wraps the master we received
	input := kettle2.StartInput{
		MasterCtx: kCtx,
		Master: func(raw interface{}) error {
			return kCtx.doMaster(ctx, keyers...)
		},
	}

	// Now, start the Kettle instance with our input and done channel
	// This can only fail if the start input is nil and since the input is
	// created in this function, Start cannot fail
	_ = kCtx.Instance.Start(ctx, &input, kCtx.done)

	// Before we run the slave function, ensure that the last thing to
	// be done here is to tell Kettle that the master should be stopped
	defer func() {
		kCtx.cancel()
		<-kCtx.done
	}()

	// Finally, run our slave function; return any error that occurs
	return slave(ctx)
}

// Helper function that performs the master function(s) if necessary
func (kCtx *KettleContext) doMaster(ctx context.Context, keyers ...*Keyer) error {

	// Attempt to get a Redis connection from the pool; ensure that this is
	// closed when the function exits regardless of its outcome
	pool := kCtx.Instance.Pool()

	// Iterate over all our keyers and check if we should run the associated
	// operation for each
	return concurrency.ForAllAsync(ctx, len(keyers), true,
		func(innerCtx context.Context, index int, cancel context.CancelFunc) error {
			return kCtx.doMasterInner(innerCtx, pool, keyers[index])
		})
}

// Helper function that actually does the master function
func (kCtx *KettleContext) doMasterInner(ctx context.Context, pool *redis.Pool, keyer *Keyer) error {

	// First, get a connection from the connection pool; ensure that this connection
	// is closed when we're done with this function
	conn := pool.Get()
	defer conn.Close()

	// Next, attempt to get the value associated with the key from Redis; we don't
	// actually care about the value here. We only care about its existence
	key := fmt.Sprintf("%s_%s", kCtx.Name, keyer.Suffix)
	_, err := redis.Int(conn.Do("GET", key))

	// Now, check if the Redis request returned an error; if it did then we'll have to
	// check the error type. Otherwise, we have nothing else to do for this iteration
	var shouldRun bool
	if err != nil {

		// Check if the error is a Redis NIL error. If it isn't then we have some sort of
		// failure so wrap the error in a more detailed message and return it
		if err != redis.ErrNil {
			return kCtx.Logger.Error(err, "Failed to get %s from Redis", key)
		}

		// We discovered that no entry associated with our key exists so set it in Redis
		// with the desired expiration; if this fails return an error. Also, set the flag
		// indicating that the operation should run
		shouldRun = true
		if _, err := conn.Do("SETEX", key, int(keyer.Frequency.Seconds()), 0); err != nil {
			return kCtx.Logger.Error(err, "Failed to set timer for %s in Redis", key)
		}
	}

	// Finally, if we want to run the operation then do so here
	if shouldRun {
		return keyer.Operation(ctx)
	}

	return nil
}
