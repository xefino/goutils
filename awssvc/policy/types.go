package policy

import (
	"fmt"

	"github.com/xefino/goutils/strings"
)

// Policy describes a policy document that can be used to configure permissions in IAM
type Policy struct {
	Version    string       `json:"Version"`
	ID         string       `json:"Id"`
	Statements []*Statement `json:"Statement"`
}

// Statement describes a set of permissions that define what resources and users should have access
// to the resources described therein
type Statement struct {
	ID            string     `json:"Sid"`
	Effect        Effect     `json:"Effect"`
	PrincipalArns Principals `json:"Principal"`
	ActionArns    Actions    `json:"Action"`
	ResourceArns  Resources  `json:"Resource"`
}

// Principals describes a list of principals associated with a policy statement
type Principals []string

// MarhsalJSON converts a Principals collection to JSON
func (p Principals) MarshalJSON() ([]byte, error) {

	// First, get the inner string from the list of principals
	var inner string
	if len(p) > 1 {
		inner = marshal(p...)
	} else if len(p) == 1 {
		inner = strings.Quote(p[0], "\"")
	} else {
		return nil, fmt.Errorf("Principal must contain at least one element")
	}

	// Next, create the principal block and return it
	return []byte(fmt.Sprintf("{\"AWS\": %s}", inner)), nil
}

// Actions describes a list of actions that may or may not be taken by principals with regard to the
// resources described in a policy statement
type Actions []Action

// MarshalJSON converts an Actions collection to JSON
func (a Actions) MarshalJSON() ([]byte, error) {

	// First, get the inner string from the list of actions
	var inner string
	if len(a) > 1 {
		inner = marshal(a...)
	} else if len(a) == 1 {
		inner = strings.Quote(a[0], "\"")
	} else {
		return nil, fmt.Errorf("Action must contain at least one element")
	}

	// Next, create the action block and return it
	return []byte(inner), nil
}

// Resources describes a list of resources effected by the policy statement
type Resources []string

// MarshalJSON converts a Resources collection to JSON
func (r Resources) MarshalJSON() ([]byte, error) {

	// First, get the inner string from the list of actions
	var inner string
	if len(r) > 1 {
		inner = marshal(r...)
	} else if len(r) == 1 {
		inner = strings.Quote(r[0], "\"")
	} else {
		return nil, fmt.Errorf("Resource must contain at least one element")
	}

	// Next, create the action block and return it
	return []byte(inner), nil
}

// Helper function that converts a list of items to a JSON-string
func marshal[S ~string](items ...S) string {
	return "[" + strings.ModifyAndJoin(func(item string) string { return strings.Quote(item, "\"") }, ",", items...) + "]"
}
