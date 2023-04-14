package orm

import (
	"fmt"

	"github.com/xefino/goutils/strings"
)

// ProcedureCall contains the functionality necessary to describe calls to SQL stored procedures
type ProcedureCall struct {
	procedure string
	arguments []any
}

// NewProcedureCall creates a new procedure call from a name and a list of arguments
func NewProcedureCall(name string, args ...any) *ProcedureCall {
	return &ProcedureCall{
		procedure: name,
		arguments: args,
	}
}

// Source returns the source of the ProcedureCall, i.e. its procedure name
func (call *ProcedureCall) Source() string {
	return call.procedure
}

// String converts a ProcedureCall to its string equivalent
func (call *ProcedureCall) String() string {
	return fmt.Sprintf("CALL %s(%s)", call.procedure, strings.Delimit("?", ", ", len(call.arguments)))
}

// Arguments returns the arguments that should be injected into the ProcedureCall
func (call *ProcedureCall) Arguments() []any {
	return call.arguments
}
