package orm

// Constant comparison operations
const (
	Equals             = "="
	NotEqual           = "<>"
	GreaterThan        = ">"
	GreaterThanEqualTo = ">="
	LessThan           = "<"
	LessThanEqualTo    = "<="
	Like               = "LIKE"
)

// WhereClause describes the functionality that should exist in any where clause term
type WhereClause[T any] interface {
	ModifyQuery(*Query[T])
}

// ConstantQueryTerm contains the functionality allowing the user to compare a database field to a concrete value
type ConstantQueryTerm[T any] struct {
	Name     string
	Operator string
	Value    string
}

// NewConstantQueryTerm creates a new constant query term from the field name, the comparison operation
// and a value to which the field should be compared
func NewConstantQueryTerm[T any](name string, op string, value string) *ConstantQueryTerm[T] {
	return &ConstantQueryTerm[T]{
		Name:     name,
		Operator: op,
		Value:    value,
	}
}

// ModifyQuery modifies the query to include this query term
func (term *ConstantQueryTerm[T]) ModifyQuery(query *Query[T]) {
	query.filter.WriteString(term.Name)
	query.filter.WriteByte(' ')
	query.filter.WriteString(term.Operator)
	query.filter.WriteByte(' ')
	query.filter.WriteString(term.Value)
}

// InjectedQueryTerm creates a new query term that injects an argument into the query's arguments list
// and a placeholder ? into the query string itself. This is intended for variables that need to be
// added to SQL queries, thereby avoiding the possibility of SQL injection
type InjectedQueryTerm[T any] struct {
	Name     string
	Operator string
	Value    any
}

// NewInjectedQueryTerm creates a new injected query term from a field name, a comparison operation,
// and a value to which the field should be compared, to be injected later
func NewInjectedQueryTerm[T any](name string, op string, value any) *InjectedQueryTerm[T] {
	return &InjectedQueryTerm[T]{
		Name:     name,
		Operator: op,
		Value:    value,
	}
}

// ModifyQuery modifies the query to include this query term
func (term *InjectedQueryTerm[T]) ModifyQuery(query *Query[T]) {
	query.filter.WriteString(term.Name)
	query.filter.WriteByte(' ')
	query.filter.WriteString(term.Operator)
	query.filter.WriteString(" ?")
	query.arguments = append(query.arguments, term.Value)
}

// MultiQueryTerm creates a new query term that allows multiple query terms to be joined together inside
// a set of parentheses. This is intended to allow for alternating sets of AND/OR logic (i.e. A AND (B OR C))
type MultiQueryTerm[T any] struct {
	Operator string
	Inner    []WhereClause[T]
}

// NewMultiQueryTerm creates a new multi-query term from a connecting operator and a list of inner terms
func NewMultiQueryTerm[T any](op string, terms ...WhereClause[T]) *MultiQueryTerm[T] {
	return &MultiQueryTerm[T]{
		Operator: op,
		Inner:    terms,
	}
}

// ModifyQuery modifies the query to include this query term
func (term *MultiQueryTerm[T]) ModifyQuery(query *Query[T]) {

	// First, if we have no inner terms then we have no work to do so return here
	if len(term.Inner) == 0 {
		return
	}

	// Next, add the opening parenthesis and first where clause. Also, ensure that the
	// closing parenthesis is added when the function exits
	query.filter.WriteRune('(')
	term.Inner[0].ModifyQuery(query)
	defer query.filter.WriteRune(')')

	// Now, if we only have one where clause then return here
	if len(term.Inner) == 1 {
		return
	}

	// Finally, iterate over all the remaining where clauses and add each to the query
	for _, clause := range term.Inner[1:] {
		query.filter.WriteByte(' ')
		query.filter.WriteString(term.Operator)
		query.filter.WriteByte(' ')
		clause.ModifyQuery(query)
	}

	return
}

// FunctionCallQueryTerm creates a new function call query term that allows the user to inject an SQL function
// call into the WHERE clause of an SQL query
type FunctionCallQueryTerm[T any] struct {
	Call      string
	Arguments []any
}

// NewFunctionCallQueryTerm creates a new function call query term from a function name and arguments
func NewFunctionCallQueryTerm[T any](call string, args ...any) *FunctionCallQueryTerm[T] {
	return &FunctionCallQueryTerm[T]{
		Call:      call,
		Arguments: args,
	}
}

// ModifyQuery modifies the query to include this query term
func (term *FunctionCallQueryTerm[T]) ModifyQuery(query *Query[T]) {
	query.filter.WriteString(term.Call)
	query.arguments = append(query.arguments, term.Arguments...)
}
