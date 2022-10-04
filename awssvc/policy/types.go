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
	} else {
		inner = strings.Quote(p[0], "\"")
	}

	// Next, create the principal block and return it
	return []byte(fmt.Sprintf("{\"AWS\": %s}", inner)), nil
}

// Actions describes a list of actions that may or may not be taken by principals with regard to the
// resources described in a policy statement
type Actions []string

// MarshalJSON converts an Actions collection to JSON
func (a Actions) MarshalJSON() ([]byte, error) {

	// First, get the inner string from the list of actions
	var inner string
	if len(a) > 1 {
		inner = marshal(a...)
	} else {
		inner = strings.Quote(a[0], "\"")
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
	} else {
		inner = strings.Quote(r[0], "\"")
	}

	// Next, create the action block and return it
	return []byte(inner), nil
}

func marshal(items ...string) string {
	return "[" + strings.ModifyAndJoin(func(item string) string { return strings.Quote(item, "\"") }, ",", items...) + "]"
}
