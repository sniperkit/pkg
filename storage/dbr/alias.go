package dbr

import "github.com/corestoreio/errors"

type alias struct {
	// Select used in cases where a sub-select is required.
	Select *Select
	// Expression can be any kind of SQL expression or a valid identifier.
	Expression string
	// Alias must be a valid identifier allowed for alias usage.
	Alias string
}

// MakeAlias creates a new alias expression
func MakeAlias(as ...string) alias {
	a := alias{
		Expression: as[0],
	}
	if len(as) > 1 {
		a.Alias = as[1]
	}
	return a
}

func (t alias) String() string {
	return Quoter.ExprAlias(t.Expression, t.Alias)
}

func (t alias) QuoteAs() string {
	return Quoter.QuoteAs(t.Expression, t.Alias)
}

// FquoteAs writes the quoted table and its maybe alias into w.
func (t alias) FquoteAs(w queryWriter) (Arguments, error) {
	if t.Select != nil {
		w.WriteRune('(')
		args, err := t.Select.toSQL(w)
		w.WriteRune(')')
		w.WriteString(" AS ")
		Quoter.quote(w, t.Alias)
		return args, errors.Wrap(err, "[dbr] FquoteAs.SubSelect")
	}
	Quoter.FquoteAs(w, t.Expression, t.Alias)
	return nil, nil
}

// TODO(CyS) if we need to distinguish between table name and the column or even need
// a sub select in the column list, then we can implement type aliases and replace
// all []string with type aliases. This costs some allocs but for modifying queries
// in dispatched events, it's getting easier ...
//type aliases []alias
//
//func makeAliasesFromStrings(columns ...string) aliases {
//
//}
