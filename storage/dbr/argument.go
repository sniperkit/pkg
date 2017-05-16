// Copyright 2015-2017, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbr

import (
	"database/sql/driver"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/corestoreio/errors"
)

// https://www.adampalmer.me/iodigitalsec/2013/08/18/mysql_real_escape_string-wont-magically-solve-your-sql-injection-problems/

// Comparison functions and operators describe all available possibilities.
const (
	Null           Op = 'n'          // IS NULL
	NotNull        Op = 'N'          // IS NOT NULL
	In             Op = '∈'          // IN ?
	NotIn          Op = '∉'          // NOT IN ?
	Between        Op = 'b'          // BETWEEN ? AND ?
	NotBetween     Op = 'B'          // NOT BETWEEN ? AND ?
	Like           Op = 'l'          // LIKE ?
	NotLike        Op = 'L'          // NOT LIKE ?
	Greatest       Op = '≫'          // GREATEST(?,?,?)
	Least          Op = '≪'          // LEAST(?,?,?)
	Equal          Op = '='          // = ?
	NotEqual       Op = '≠'          // != ?
	Exists         Op = '∃'          // EXISTS(subquery)
	NotExists      Op = '∄'          // NOT EXISTS(subquery)
	Less           Op = '<'          // <
	Greater        Op = '>'          // >
	LessOrEqual    Op = '≤'          // <=
	GreaterOrEqual Op = '≥'          // >=
	Regexp         Op = 'r'          // REGEXP ?
	NotRegexp      Op = 'R'          // NOT REGEXP ?
	Xor            Op = '⊻'          // XOR ?
	SpaceShip      Op = '\U0001f680' // a <=> b is equivalent to a = b OR (a IS NULL AND b IS NULL)
)

// Op the Operator, defines comparison and operator functions used in any
// conditions. The upper case letter always negates.
// https://dev.mysql.com/doc/refman/5.7/en/comparison-operators.html
// https://mariadb.com/kb/en/mariadb/comparison-operators/
type Op rune

// isNotIn returns true if the operator is not one of the four types in the
// code.
func (o Op) isNotIn() bool {
	switch o {
	case In, NotIn, Greatest, Least:
		return false
	}
	return true
}

// String transforms the rune into a string.
func (o Op) String() string {
	if o < 1 {
		o = Equal
	}
	return string(o)
}

// With allows to use any argument with an operator.
func (o Op) With(arg Argument) Argument {
	return arg.Operator(o)
}

// Str uses string values for comparison.
func (o Op) Str(values ...string) Argument {
	if len(values) == 0 {
		return argPlaceHolder(o)
	}
	return &argStrings{data: values, op: o}
}

// NullString uses nullable string values for comparison.
func (o Op) NullString(values ...NullString) Argument {
	if len(values) == 1 {
		values[0].op = o
		return values[0]
	}
	return argNullStrings{data: values, op: o}
}

// Float64 uses float64 values for comparison.
func (o Op) Float64(values ...float64) Argument {
	return &argFloat64s{data: values, op: o}
}

// NullFloat64 uses nullable float64 values for comparison.
func (o Op) NullFloat64(values ...NullFloat64) Argument {
	if len(values) == 1 {
		values[0].op = o
		return values[0]
	}
	return argNullFloat64s{data: values, op: o}
}

// Int64 uses int64 values for comparison.
func (o Op) Int64(values ...int64) Argument {
	if len(values) == 0 {
		return argPlaceHolder(o)
	}
	return &argInt64s{data: values, op: o}
}

// NullInt64 uses nullable int64 values for comparison.
func (o Op) NullInt64(values ...NullInt64) Argument {
	if len(values) == 1 {
		values[0].op = o
		return values[0]
	}
	return argNullInt64s{data: values, op: o}
}

// Int uses int values for comparison.
func (o Op) Int(values ...int) Argument {
	return &argInts{data: values, op: o}
}

// Bool uses bool values for comparison.
func (o Op) Bool(values ...bool) Argument {
	return &argBools{data: values, op: o}
}

// NullBool uses nullable bool values for comparison.
func (o Op) NullBool(value NullBool) Argument {
	value.op = o
	return value
}

// Time uses time.Time values for comparison.
func (o Op) Time(values ...time.Time) Argument {
	return &argTimes{data: values, op: o}
}

// NullTime uses nullable time values for comparison.
func (o Op) NullTime(values ...NullTime) Argument {
	if len(values) == 1 {
		values[0].op = o
		return values[0]
	}
	return argNullTimes{data: values, op: o}
}

// Null is always a NULL.
func (o Op) Null() Argument {
	return argNull(o)
}

// Bytes uses a byte slice for comparison. Providing a nil argument returns a
// NULL type. Detects between valid UTF-8 strings and binary data. Later gets
// hex encoded.
func (o Op) Bytes(p ...[]byte) Argument {
	return argBytes{data: p, op: o}
}

// Value uses driver.Valuers for comparison.
func (o Op) Value(values ...driver.Valuer) Argument {
	return &argValue{data: values, op: o}
}

const (
	sqlStrNull     = "NULL"
	sqlStar        = "*"
	stmtTypeUpdate = 'u'
	stmtTypeDelete = 'd'
	stmtTypeInsert = 'i'
	stmtTypeSelect = 's'
)

// ArgumentAssembler assembles arguments for a SQL INSERT or UPDATE statement.
// The `stmtType` variable is either `i` for INSERT or `u` for UPDATE. Future
// uses for `s` (SELECT) might occur. Any new arguments must be append to
// variable `args` and then returned. Variable `columns` contains the name of
// the requested columns. E.g. if the first requested column names `id` then the
// first appended argument must be an integer. Variable `condition` contains the
// names and/or expressions used in the WHERE or ON clause, if applicable for
// the SQL statement type.
type ArgumentAssembler interface {
	AssembleArguments(stmtType rune, args Arguments, columns, condition []string) (Arguments, error)
}

// Argument transforms a value or values into an interface slice or encodes them
// into their textual representation to be used directly in a SQL query. This
// interface slice gets used in the database query functions as an argument. The
// underlying primitive type in the interface must be one of driver.Value
// allowed types.
type Argument interface {
	// Operator sets a comparison or logical operator. Please see the constants
	// Op* for the different flags. An underscore in the argument list of
	// a type indicates that no operator is yet supported.
	Operator(Op) Argument
	// toIFace appends the value or values to interface slice and returns it.
	toIFace([]interface{}) []interface{}
	// writeTo writes the value correctly escaped to the queryWriter. It must
	// avoid SQL injections.
	writeTo(w queryWriter, position int) error
	// len returns the length of the available values. If the IN clause has been
	// activated then len returns 1.
	len() int
	operator() Op
}

// Arguments representing multiple arguments.
type Arguments []Argument

func writeOperator(w queryWriter, op Op, hasArg bool) (addArg bool) {
	// hasArg argument only used in cases where we have in the parent caller
	// function a sub-select. sub-selects do not need a place holder.
	switch op {
	case Null:
		w.WriteString(" IS NULL")
	case NotNull:
		w.WriteString(" IS NOT NULL")
	case In:
		w.WriteString(" IN ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	case NotIn:
		w.WriteString(" NOT IN ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	case Like:
		w.WriteString(" LIKE ?")
		addArg = true
	case NotLike:
		w.WriteString(" NOT LIKE ?")
		addArg = true
	case Regexp:
		w.WriteString(" REGEXP ?")
		addArg = true
	case NotRegexp:
		w.WriteString(" NOT REGEXP ?")
		addArg = true
	case Between:
		w.WriteString(" BETWEEN ? AND ?")
		addArg = true
	case NotBetween:
		w.WriteString(" NOT BETWEEN ? AND ?")
		addArg = true
	case Greatest:
		w.WriteString(" GREATEST (?)")
		addArg = true
	case Least:
		w.WriteString(" LEAST (?)")
		addArg = true
	case Xor:
		w.WriteString(" XOR ?")
		addArg = true
	case Exists:
		w.WriteString(" EXISTS ")
		addArg = true
	case NotExists:
		w.WriteString(" NOT EXISTS ")
		addArg = true
	case Equal:
		w.WriteString(" = ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	case NotEqual:
		w.WriteString(" != ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	case Less:
		w.WriteString(" < ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	case Greater:
		w.WriteString(" > ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	case LessOrEqual:
		w.WriteString(" <= ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	case GreaterOrEqual:
		w.WriteString(" >= ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	case SpaceShip:
		w.WriteString(" <=> ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	default:
		w.WriteString(" = ")
		if hasArg {
			w.WriteByte('?')
			addArg = true
		}
	}
	return
}

// len calculates the total length of all values
func (as Arguments) len() (l int) {
	for _, a := range as {
		l += a.len()
	}
	return
}

// Interfaces converts the underlying concrete types into an interface slice.
// Each entry in the interface is guaranteed to be one of the following values:
// []byte, bool, float64, int64, string or time.Time. Use driver.IsValue() for a
// check.
func (as Arguments) Interfaces() []interface{} {
	if len(as) == 0 {
		return nil
	}
	ret := make([]interface{}, 0, len(as))
	for _, a := range as {
		ret = a.toIFace(ret)
	}
	return ret
}

// DriverValues converts the []interfaces to driver.Value.
func (as Arguments) DriverValues() []driver.Value {
	if len(as) == 0 {
		return nil
	}
	// TODO Optimize this function to reduce allocations. Maybe change the internal interface function.
	iFaces := as.Interfaces()
	dv := make([]driver.Value, len(iFaces))

	for i, r := range iFaces {
		dv[i] = driver.Value(r)
	}
	return dv
}

type argValue struct {
	op   Op
	data []driver.Valuer
}

func (a *argValue) toIFace(args []interface{}) []interface{} {
	for _, v := range a.data {
		args = append(args, v)
	}
	return args
}

func writeDriverValuer(w queryWriter, value driver.Valuer) error {
	if value == nil {
		_, err := w.WriteString("NULL")
		return err
	}
	val, err := value.Value()
	if err != nil {
		return errors.Wrapf(err, "[dbr] argValue.WriteTo: %#v", value)
	}
	switch t := val.(type) {
	case nil:
		_, err = w.WriteString("NULL")
	case int:
		_, err = w.WriteString(strconv.Itoa(t))
	case int64:
		_, err = w.WriteString(strconv.FormatInt(t, 10))
	case float64:
		_, err = w.WriteString(strconv.FormatFloat(t, 'f', -1, 64))
	case string:
		dialect.EscapeString(w, t)
	case bool:
		dialect.EscapeBool(w, t)
	case time.Time:
		dialect.EscapeTime(w, t)
	case []byte:
		dialect.EscapeBinary(w, t)
	default:
		return errors.NewNotSupportedf("[dbr] argValue.WriteTo Type not yet supported: %#v", value)
	}
	return err
}

func (a *argValue) writeTo(w queryWriter, pos int) error {
	if a.op.isNotIn() {
		return writeDriverValuer(w, a.data[pos])
	}

	l := len(a.data) - 1
	w.WriteByte('(')
	for i, value := range a.data {
		if err := writeDriverValuer(w, value); err != nil {
			return err
		}
		if i < l {
			w.WriteByte(',')
		}
	}
	w.WriteByte(')')
	return nil
}

func (a *argValue) len() int {
	if a.op.isNotIn() {
		return len(a.data)
	}
	return 1
}

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a *argValue) Operator(op Op) Argument {
	a.op = op
	return a
}

func (a *argValue) operator() Op { return a.op }

// ArgValue allows to use any type which implements driver.Valuer interface.
func ArgValue(args ...driver.Valuer) Argument {
	return &argValue{
		data: args,
	}
}

type argTimes struct {
	op   Op
	data []time.Time
}

func (a *argTimes) toIFace(args []interface{}) []interface{} {
	for _, v := range a.data {
		args = append(args, v)
	}
	return args
}

func (a *argTimes) writeTo(w queryWriter, pos int) error {
	if a.op.isNotIn() {
		dialect.EscapeTime(w, a.data[pos])
		return nil
	}
	l := len(a.data) - 1
	w.WriteByte('(')
	for i, v := range a.data {
		dialect.EscapeTime(w, v)
		if i < l {
			w.WriteByte(',')
		}
	}
	w.WriteByte(')')
	return nil
}

func (a *argTimes) len() int {
	if a.op.isNotIn() {
		return len(a.data)
	}
	return 1
}

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a *argTimes) Operator(op Op) Argument {
	a.op = op
	return a
}

func (a *argTimes) operator() Op { return a.op }

// ArgTime adds a time.Time or a slice of times to the argument list.
// Providing no arguments returns a NULL type.
func ArgTime(args ...time.Time) Argument {
	return &argTimes{data: args}
}

type argBytes struct {
	op   Op
	data [][]byte
}

func (a argBytes) toIFace(args []interface{}) []interface{} {
	return append(args, a.data)
}

func (a argBytes) writeTo(w queryWriter, pos int) (err error) {
	if a.op.isNotIn() {
		if !utf8.Valid(a.data[pos]) {
			dialect.EscapeBinary(w, a.data[pos])
		} else {
			dialect.EscapeString(w, string(a.data[pos]))
		}
		return nil
	}
	l := len(a.data) - 1
	w.WriteByte('(')
	for i, v := range a.data {
		if !utf8.Valid(v) {
			dialect.EscapeBinary(w, v)
		} else {
			dialect.EscapeString(w, string(v))
		}
		if i < l {
			w.WriteByte(',')
		}
	}
	w.WriteByte(')')
	return nil
}

func (a argBytes) len() int {
	if a.op.isNotIn() {
		return len(a.data)
	}
	return 1
}

// Op not supported
func (a argBytes) Operator(op Op) Argument { a.op = op; return a }
func (a argBytes) operator() Op            { return a.op }

// ArgBytes adds a byte slice to the argument list. Providing a nil argument
// returns a NULL type. Detects between valid UTF-8 strings and binary data. Later
// gets hex encoded.
func ArgBytes(p ...[]byte) Argument {
	if p == nil {
		return ArgNull()
	}
	return argBytes{data: p}
}

type argNull rune

func (i argNull) toIFace(args []interface{}) []interface{} {
	return append(args, nil)
}

func (i argNull) writeTo(w queryWriter, _ int) (err error) {
	if Op(i).isNotIn() {
		_, err = w.WriteString("NULL")
	} else {
		w.WriteByte('(')
		_, err = w.WriteString("NULL")
		w.WriteByte(')')
	}
	return err
}

func (i argNull) len() int { return 1 }

// Op not supported
func (i argNull) Operator(op Op) Argument { return argNull(op) }
func (i argNull) operator() Op {
	if i > 0 {
		return Op(i)
	}
	return Null
}

// ArgNull treats the argument as a SQL `IS NULL` or `NULL`.
// IN clause not supported.
func ArgNull() Argument {
	return argNull(0)
}

// argString implements interface Argument but does not allocate.
type argString string

func (a argString) toIFace(args []interface{}) []interface{} {
	return append(args, string(a))
}

func (a argString) writeTo(w queryWriter, _ int) error {
	if !utf8.ValidString(string(a)) {
		return errors.NewNotValidf("[dbr] Argument.WriteTo: String is not UTF-8: %q", a)
	}
	dialect.EscapeString(w, string(a))
	return nil
}

func (a argString) len() int { return 1 }

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a argString) Operator(op Op) Argument {
	return &argStrings{
		data: []string{string(a)},
		op:   op,
	}
}
func (a argString) operator() Op { return 0 }

type argStrings struct {
	data []string
	op   Op
}

func (a *argStrings) toIFace(args []interface{}) []interface{} {
	for _, v := range a.data {
		args = append(args, v)
	}
	return args
}

func (a *argStrings) writeTo(w queryWriter, pos int) error {
	if a.op.isNotIn() {
		if !utf8.ValidString(a.data[pos]) {
			return errors.NewNotValidf("[dbr] Argument.WriteTo: String is not UTF-8: %q", a.data[pos])
		}
		dialect.EscapeString(w, a.data[pos])
		return nil
	}
	l := len(a.data) - 1
	w.WriteByte('(')
	for i, v := range a.data {
		if !utf8.ValidString(v) {
			return errors.NewNotValidf("[dbr] Argument.WriteTo: String is not UTF-8: %q", v)
		}
		dialect.EscapeString(w, v)
		if i < l {
			w.WriteByte(',')
		}
	}
	w.WriteByte(')')
	return nil
}

func (a *argStrings) len() int {
	if a.op.isNotIn() {
		return len(a.data)
	}
	return 1
}

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a *argStrings) Operator(op Op) Argument {
	a.op = op
	return a
}
func (a *argStrings) operator() Op { return a.op }

// ArgString adds a string or a slice of strings to the argument list.
// Providing no arguments returns a NULL type.
// All arguments mut be a valid utf-8 string.
func ArgString(args ...string) Argument {
	if len(args) == 1 {
		return argString(args[0])
	}
	return &argStrings{data: args}
}

type argBool bool

func (a argBool) toIFace(args []interface{}) []interface{} {
	return append(args, a == true)
}

func (a argBool) writeTo(w queryWriter, _ int) error {
	dialect.EscapeBool(w, a == true)
	return nil
}
func (a argBool) len() int { return 1 }

// Op not supported
func (a argBool) Operator(_ Op) Argument { return a }
func (a argBool) operator() Op           { return 0 }

type argBools struct {
	op   Op
	data []bool
}

func (a *argBools) toIFace(args []interface{}) []interface{} {
	for _, v := range a.data {
		args = append(args, v == true)
	}
	return args
}

func (a *argBools) writeTo(w queryWriter, pos int) error {
	if a.op.isNotIn() {
		dialect.EscapeBool(w, a.data[pos])
		return nil
	}
	l := len(a.data) - 1
	w.WriteByte('(')
	for i, v := range a.data {
		dialect.EscapeBool(w, v == true)
		if i < l {
			w.WriteByte(',')
		}
	}
	w.WriteByte(')')
	return nil
}

func (a *argBools) len() int {
	if a.op.isNotIn() {
		return len(a.data)
	}
	return 1
}

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a *argBools) Operator(op Op) Argument {
	a.op = op
	return a
}
func (a *argBools) operator() Op { return a.op }

// ArgBool adds a string or a slice of bools to the argument list.
// Providing no arguments returns a NULL type.
func ArgBool(args ...bool) Argument {
	if len(args) == 1 {
		return argBool(args[0])
	}
	return &argBools{data: args}
}

// argInt implements interface Argument but does not allocate.
type argInt int

func (a argInt) toIFace(args []interface{}) []interface{} {
	return append(args, int64(a))
}

func (a argInt) writeTo(w queryWriter, _ int) error {
	_, err := w.WriteString(strconv.FormatInt(int64(a), 10))
	return err
}
func (a argInt) len() int { return 1 }

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a argInt) Operator(op Op) Argument {
	return &argInts{
		op:   op,
		data: []int{int(a)},
	}
}
func (a argInt) operator() Op { return 0 }

type argInts struct {
	op   Op
	data []int
}

func (a *argInts) toIFace(args []interface{}) []interface{} {
	for _, v := range a.data {
		args = append(args, int64(v))
	}
	return args
}

func (a *argInts) writeTo(w queryWriter, pos int) error {
	if a.op.isNotIn() {
		_, err := w.WriteString(strconv.Itoa(a.data[pos]))
		return err
	}
	l := len(a.data) - 1
	w.WriteByte('(')
	for i, v := range a.data {
		w.WriteString(strconv.Itoa(v))
		if i < l {
			w.WriteByte(',')
		}
	}
	w.WriteByte(')')
	return nil
}

func (a *argInts) len() int {
	if a.op.isNotIn() {
		return len(a.data)
	}
	return 1
}

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a *argInts) Operator(op Op) Argument {
	a.op = op
	return a
}

func (a *argInts) operator() Op { return a.op }

// ArgInt adds an integer or a slice of integers to the argument list.
// Providing no arguments returns a NULL type.
func ArgInt(args ...int) Argument {
	if len(args) == 1 {
		return argInt(args[0])
	}
	return &argInts{data: args}
}

// argInt64 implements interface Argument but does not allocate.
type argInt64 int64

func (a argInt64) toIFace(args []interface{}) []interface{} {
	return append(args, int64(a))
}

func (a argInt64) writeTo(w queryWriter, _ int) error {
	_, err := w.WriteString(strconv.FormatInt(int64(a), 10))
	return err
}
func (a argInt64) len() int { return 1 }

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a argInt64) Operator(op Op) Argument {
	return &argInt64s{
		op:   op,
		data: []int64{int64(a)},
	}
}
func (a argInt64) operator() Op { return 0 }

type argInt64s struct {
	op   Op
	data []int64
}

func (a *argInt64s) toIFace(args []interface{}) []interface{} {
	for _, v := range a.data {
		args = append(args, v)
	}
	return args
}

func (a *argInt64s) writeTo(w queryWriter, pos int) error {
	if a.op.isNotIn() {
		_, err := w.WriteString(strconv.FormatInt(a.data[pos], 10))
		return err
	}
	l := len(a.data) - 1
	w.WriteByte('(')
	for i, v := range a.data {
		w.WriteString(strconv.FormatInt(v, 10))
		if i < l {
			w.WriteByte(',')
		}
	}
	w.WriteByte(')')
	return nil
}

func (a *argInt64s) len() int {
	if a.op.isNotIn() {
		return len(a.data)
	}
	return 1
}

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a *argInt64s) Operator(op Op) Argument {
	a.op = op
	return a
}

func (a *argInt64s) operator() Op { return a.op }

// ArgInt64 adds an integer or a slice of integers to the argument list.
// Providing no arguments returns a NULL type.
func ArgInt64(args ...int64) Argument {
	if len(args) == 1 {
		return argInt64(args[0])
	}
	return &argInt64s{data: args}
}

type argFloat64 float64

func (a argFloat64) toIFace(args []interface{}) []interface{} {
	return append(args, float64(a))
}

func (a argFloat64) writeTo(w queryWriter, _ int) error {
	_, err := w.WriteString(strconv.FormatFloat(float64(a), 'f', -1, 64))
	return err
}
func (a argFloat64) len() int { return 1 }

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a argFloat64) Operator(op Op) Argument {
	return &argFloat64s{
		op:   op,
		data: []float64{float64(a)},
	}
}
func (a argFloat64) operator() Op { return 0 }

type argFloat64s struct {
	op   Op
	data []float64
}

func (a *argFloat64s) toIFace(args []interface{}) []interface{} {
	for _, v := range a.data {
		args = append(args, v)
	}
	return args
}

func (a *argFloat64s) writeTo(w queryWriter, pos int) error {
	if a.op.isNotIn() {
		_, err := w.WriteString(strconv.FormatFloat(a.data[pos], 'f', -1, 64))
		return err
	}
	l := len(a.data) - 1
	w.WriteByte('(')
	for i, v := range a.data {
		w.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		if i < l {
			w.WriteByte(',')
		}
	}
	w.WriteByte(')')
	return nil
}

func (a *argFloat64s) len() int {
	if a.op.isNotIn() {
		return len(a.data)
	}
	return 1
}

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (a *argFloat64s) Operator(op Op) Argument {
	a.op = op
	return a
}

func (a *argFloat64s) operator() Op { return a.op }

// ArgFloat64 adds a float64 or a slice of floats to the argument list.
// Providing no arguments returns a NULL type.
func ArgFloat64(args ...float64) Argument {
	if len(args) == 1 {
		return argFloat64(args[0])
	}
	return &argFloat64s{data: args}
}

type expr struct {
	SQL string
	Arguments
	op Op
}

// ArgExpr at a SQL fragment with placeholders, and a slice of args to replace them
// with. Mostly used in UPDATE statements.
func ArgExpr(sql string, args ...Argument) Argument {
	return &expr{SQL: sql, Arguments: args}
}

func (e *expr) toIFace(args []interface{}) []interface{} {
	for _, a := range e.Arguments {
		args = a.toIFace(args)
	}
	return args
}

func (e *expr) writeTo(w queryWriter, _ int) error {
	w.WriteString(e.SQL)
	return nil
}
func (e *expr) len() int { return 1 }

// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
// the constants Op*.
func (e *expr) Operator(op Op) Argument {
	e.op = op
	return e
}

func (e *expr) operator() Op { return e.op }

type argPlaceHolder rune

func (i argPlaceHolder) toIFace(args []interface{}) []interface{} {
	return args //append(args, nil)
}

func (i argPlaceHolder) writeTo(w queryWriter, _ int) (err error) {
	_, err = w.WriteString("?/*SHOULD NOT GET WRITTEN*/")
	return err
}

func (i argPlaceHolder) len() int {
	return cahensConstant
}

// Op not supported
func (i argPlaceHolder) Operator(op Op) Argument { return argPlaceHolder(op) }
func (i argPlaceHolder) operator() Op {
	return Op(i)
}

// for type subQuery see function SubSelect

//type argSubSelect struct {
//	// buf contains the cached SQL string
//	buf *bytes.Buffer
//	// args contains the arguments after calling ToSQL
//	args Arguments
//	s    *Select
//	op Op
//}

// I don't know anymore where I would have needed this ... but once the idea
// and a real world use case pops up, I'm gonna implement it. Until then use the function
// SubSelect(rawStatementOrColumnName string, operator rune, s *Select) ConditionArg
//// ArgSubSelect
//// The written sub-select gets wrapped in parenthesis: (SELECT ...)
//func ArgSubSelect(s *Select) Argument {
//	return &argSubSelect{s: s}
//}
//
//func (e *argSubSelect) toIFace(args []interface{}) []interface{} {
//
//	if e.buf == nil {
//		e.buf = new(bytes.Buffer)
//		var err error
//		e.args, err = e.s.toSQL(e.buf) // can be optimized later
//		if err != nil {
//			args = append(args, err) // not that optimal :-(
//		} else {
//			for _, a := range e.args {
//				a.toIFace(args)
//			}
//		}
//		return
//	}
//	for _, a := range e.args {
//		a.toIFace(args)
//	}
//}
//
//func (e *argSubSelect) writeTo(w queryWriter, _ int) (err error) {
//	if e.buf == nil {
//		e.buf = new(bytes.Buffer)
//		e.buf.WriteByte('(')
//		e.args, err = e.s.toSQL(e.buf)
//		if err != nil {
//			return errors.Wrap(err, "[dbr] argSubSelect.writeTo")
//		}
//		e.buf.WriteByte(')')
//	}
//	_, err = w.WriteString(e.buf.String())
//	return err
//}
//
//func (e *argSubSelect) len() int { return 1 }
//
//// Op sets the SQL operator (IN, =, LIKE, BETWEEN, ...). Please refer to
//// the constants Op*.
//func (e *argSubSelect) Op(op Op) Argument {
//	e.op = op
//	return e
//}
//func (e *argSubSelect) operator() Op { return e.op }
