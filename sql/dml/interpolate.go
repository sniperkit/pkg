/*
Sniperkit-Bot
- Status: analyzed
*/

// Copyright 2015-present, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package dml

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/corestoreio/errors"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/storage/null"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/util/bufferpool"
)

const (
	placeHolderRune = '?'
	placeHolderStr  = `?`
)

var placeHolderByte = []byte(placeHolderStr)

// ExpandPlaceHolders takes a SQL string and repeats the question marks with the provided
// arguments. If the amount of arguments does not match the number of questions
// marks, a Mismatch error gets returned. The arguments are getting converted to
// an interface slice to easy passing into the db.Query/db.Exec/etc functions at
// an argument.
//		ExpandPlaceHolders("SELECT * FROM table WHERE id IN (?) AND status IN (?)", Int(myIntSlice...), String(myStrSlice...))
// Gets converted to:
//		SELECT * FROM table WHERE id IN (?,?) AND status IN (?,?,?)
// The questions marks are of course depending on the values in the Arg*
// functions. This function should be generally used when dealing with prepared
// statements.
func ExpandPlaceHolders(sql string, args *Artisan) (string, error) {
	var a arguments
	if args != nil {
		a = args.arguments
	}

	phCount := strings.Count(sql, placeHolderStr)
	if want := len(a); phCount != want || want == 0 {
		return "", errors.Mismatch.Newf("[dml] ExpandPlaceHolders: Number of %s:%d do not match the number of repetitions: %d", placeHolderStr, phCount, want)
	}

	var buf strings.Builder
	buf.Grow(len(sql) * 3 / 2) // *1.5
	err := expandPlaceHolders(&buf, []byte(sql), a)
	return buf.String(), errors.WithStack(err)
}

// expandPlaceHolders multiplies the place holder with the arguments internal len.
func expandPlaceHolders(buf writer, sql []byte, args arguments) error {
	i := 0
	pos := 0
	for pos < len(sql) {
		r, w := utf8.DecodeRune(sql[pos:])
		pos += w

		switch r {
		case placeHolderRune:
			if i < len(args) {
				reps := args[i].len()
				if reps > 1 {
					buf.WriteByte('(')
				}
				for r := 0; r < reps; r++ {
					buf.WriteByte(placeHolderRune)
					if r < reps-1 {
						buf.WriteByte(',')
					}
				}
				if reps > 1 {
					buf.WriteByte(')')
				}
			}
			i++
		default:
			buf.Write(sql[pos-w : pos])
		}
	}
	return nil
}

// ip handles the interpolation of the SQL string and uses an internal argument
// pool for optimal slice usage.
type ip struct {
	queryCache string
	args       arguments
	// ärgErr represents an argument error caused in any of the other functions.
	// A stack has been attached to the error to identify properly the source.
	ärgErr error // Sorry Germans for that terrible pun #notSorry
}

// Interpolate takes a SQL byte slice with placeholders and a list of arguments
// to replace them with. It returns a blank string or an error if the number of
// placeholders does not match the number of arguments. Implements the ExpandPlaceHolders
// function.
func Interpolate(sql string) *ip {
	return &ip{
		queryCache: sql,
		args:       make(arguments, 0, defaultArgumentsCapacity),
	}
}

// String implements fmt.Stringer and prints errors into the string which will
// maybe generate invalid SQL code.
func (in *ip) String() string {
	str, _, err := in.ToSQL()
	if err != nil {
		return err.Error()
	}
	return str
}

// MustString returns the fully interpolated SQL string or panics on error.
func (in *ip) MustString() string {
	str, _, err := in.ToSQL()
	if err != nil {
		panic(err)
	}
	return str
}

// ToSQL implements dml.QueryBuilder. The interface slice is always nil.
func (in *ip) ToSQL() (_ string, alwaysNil []interface{}, _ error) {
	if in.ärgErr != nil {
		return "", nil, in.ärgErr
	}
	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	if err := writeInterpolate(buf, in.queryCache, in.args); err != nil {
		return "", nil, errors.WithStack(err)
	}
	return buf.String(), nil, nil
}

// Reset resets the internal argument cache for reuse. Avoids lots of
// allocations.
func (in *ip) Reset() *ip                 { in.args = in.args[:0]; return in }
func (in *ip) Null() *ip                  { in.args = in.args.add(nil); return in }
func (in *ip) Unsafe(arg interface{}) *ip { in.args = in.args.add(arg); return in }
func (in *ip) Int(i int) *ip              { in.args = in.args.add(i); return in }
func (in *ip) Ints(i ...int) *ip          { in.args = in.args.add(i); return in }
func (in *ip) Int64(i int64) *ip          { in.args = in.args.add(i); return in }
func (in *ip) Int64s(i ...int64) *ip      { in.args = in.args.add(i); return in }
func (in *ip) Uint64(i uint64) *ip        { in.args = in.args.add(i); return in }
func (in *ip) Uint64s(i ...uint64) *ip    { in.args = in.args.add(i); return in }
func (in *ip) Float64(f float64) *ip      { in.args = in.args.add(f); return in }
func (in *ip) Float64s(f ...float64) *ip  { in.args = in.args.add(f); return in }
func (in *ip) Str(s string) *ip           { in.args = in.args.add(s); return in }
func (in *ip) Strs(s ...string) *ip       { in.args = in.args.add(s); return in }
func (in *ip) Bool(b bool) *ip            { in.args = in.args.add(b); return in }
func (in *ip) Bools(b ...bool) *ip        { in.args = in.args.add(b); return in }

// Bytes uses a byte slice for comparison. Providing a nil value returns a
// NULL type. Detects between valid UTF-8 strings and binary data. Later gets
// hex encoded.
func (in *ip) Bytes(p []byte) *ip                  { in.args = in.args.add(p); return in }
func (in *ip) BytesSlice(p ...[]byte) *ip          { in.args = in.args.add(p); return in }
func (in *ip) Time(t time.Time) *ip                { in.args = in.args.add(t); return in }
func (in *ip) Times(t ...time.Time) *ip            { in.args = in.args.add(t); return in }
func (in *ip) NullString(nv null.String) *ip       { in.args = in.args.add(nv); return in }
func (in *ip) NullStrings(nv ...null.String) *ip   { in.args = in.args.add(nv); return in }
func (in *ip) NullFloat64(nv null.Float64) *ip     { in.args = in.args.add(nv); return in }
func (in *ip) NullFloat64s(nv ...null.Float64) *ip { in.args = in.args.add(nv); return in }
func (in *ip) NullInt64(nv null.Int64) *ip         { in.args = in.args.add(nv); return in }
func (in *ip) NullInt64s(nv ...null.Int64) *ip     { in.args = in.args.add(nv); return in }
func (in *ip) NullBool(nv null.Bool) *ip           { in.args = in.args.add(nv); return in }
func (in *ip) NullBools(nv ...null.Bool) *ip       { in.args = in.args.add(nv); return in }
func (in *ip) NullTime(nv null.Time) *ip           { in.args = in.args.add(nv); return in }
func (in *ip) NullTimes(nv ...null.Time) *ip       { in.args = in.args.add(nv); return in }

// DriverValues adds each Valuer as its own argument.
func (in *ip) DriverValues(dvs ...driver.Valuer) *ip {
	if in.ärgErr != nil {
		return in
	}
	in.args, in.ärgErr = driverValues(in.args, dvs...)
	return in
}

// DriverValue packs the types of subsequent Valuer into its own slice. All
// Valuer arguments must have the same type. E.g. int,int,int as Valuer argument
// will be transformed into []int.
func (in *ip) DriverValue(dvs ...driver.Valuer) *ip {
	if in.ärgErr != nil {
		return in
	}
	in.args, in.ärgErr = driverValue(in.args, dvs...)
	return in
}

// Arguments sets the internal arguments slice and nothing else. It overwrites
// previously set arguments.
func (in *ip) Arguments(args *Artisan) *ip {
	if args != nil {
		in.args = args.arguments
	}
	return in
}

// Named uses the NamedArg for string replacement. Replaces the names with place
// holder character.
func (in *ip) Named(nArgs ...sql.NamedArg) *ip {
	// Slices in NamedArg.Value won't be supported.
	// for now this unoptimized version with a stupid string replacement and
	// converting between bytes and string.

	for _, na := range nArgs {
		in.queryCache = strings.Replace(in.queryCache, na.Name, "?", -1) // TODO optimize
		// BUG: depending on the position of the named argument the append function below creates the bug for not
		// positioning correctly the arguments.
		if in.ärgErr != nil {
			return in
		}
		in.args, in.ärgErr = iFaceToArgs(in.args, na.Value)
	}
	return in
}

// writeInterpolate merges `args` into `sql` and writes the result into `buf`. `sql`
// stays unchanged.
func writeInterpolate(buf *bytes.Buffer, sql string, args arguments) error {
	// TODO support :name identifier and the name field in argument

	phCount, argCount := strings.Count(sql, placeHolderStr), len(args)
	if argCount > 0 && phCount != argCount {
		return errors.Mismatch.Newf("[dml] Number of place holders (%d) vs number of arguments (%d) do not match.", phCount, argCount)
	}

	var phCounter int
	pos := 0
	for pos < len(sql) {
		r, w := utf8.DecodeRuneInString(sql[pos:])
		pos += w

		switch {
		case r == placeHolderRune && argCount > 0:
			if phCounter < argCount { // protect for index out of bounds
				if err := args[phCounter].writeTo(buf, 0); err != nil {
					return errors.WithStack(err)
				}
			}
			phCounter++
		case r == '`', r == '\'', r == '"':
			p := strings.IndexRune(sql[pos:], r)
			if r == '"' {
				r = '\''
			}
			buf.WriteRune(r)
			if p > -1 {
				buf.WriteString(sql[pos : pos+p])
				buf.WriteRune(r)
			}
			pos += p + 1
		case r == '[':
			w = strings.IndexRune(sql[pos:], ']')
			col := sql[pos : pos+w]
			dialect.EscapeIdent(buf, col)
			pos += w + 1 // size of ']'
		default:
			buf.WriteString(sql[pos-w : pos])
		}
	}

	return nil
}

// writeInterpolateByte same as writeInterpolate. Maybe package unsafe can do
// here some magic to avoid duplicate code, but for now we stick with a copy of
// the above original function writeInterpolateByte.
func writeInterpolateBytes(buf *bytes.Buffer, sql []byte, args arguments) error {

	phCount, argCount := bytes.Count(sql, placeHolderByte), len(args)
	if argCount > 0 && phCount != argCount {
		return errors.Mismatch.Newf("[dml] Number of place holders (%d) vs number of arguments (%d) do not match.", phCount, argCount)
	}

	var phCounter int
	pos := 0
	for pos < len(sql) {
		r, w := utf8.DecodeRune(sql[pos:])
		pos += w

		switch {
		case r == placeHolderRune && argCount > 0:
			if phCounter < argCount { // protect for index out of bounds
				if err := args[phCounter].writeTo(buf, 0); err != nil {
					return errors.WithStack(err)
				}
			}
			phCounter++
		case r == '`', r == '\'', r == '"':
			p := bytes.IndexRune(sql[pos:], r)
			if r == '"' {
				r = '\''
			}
			buf.WriteRune(r)
			if p > -1 {
				buf.Write(sql[pos : pos+p])
				buf.WriteRune(r)
			}
			pos += p + 1
		case r == '[':
			w = bytes.IndexByte(sql[pos:], ']')
			col := sql[pos : pos+w]
			dialect.EscapeIdent(buf, string(col))
			pos += w + 1 // size of ']'
		default:
			buf.Write(sql[pos-w : pos])
		}
	}

	return nil
}

// extractReplaceNamedArgs extracts all occurrences of a pattern `:[^\s]+` and
// replaces them with a ? placeholder. It does not remove duplicates because
// those are needed for the amount of arguments to get. The extracted strings
// get appended to qualifiedColumns argument.
func extractReplaceNamedArgs(sql []byte, qualifiedColumns []string) (_ []byte, _ []string, found bool) {
	if bytes.IndexByte(sql, namedArgStartByte) == -1 {
		return sql, qualifiedColumns, found
	}
	lSQL := len(sql)
	foundColon := false
	buf := bufferpool.Get()
	newSQL := bytes.NewBuffer(make([]byte, 0, lSQL))
	quoteStart := false
	pos := 0
	for pos < len(sql) {
		r, w := utf8.DecodeRune(sql[pos:])
		pos += w

		switch {
		case r == '\'' && !quoteStart:
			quoteStart = true
			newSQL.WriteRune(r)
		case quoteStart:
			// do nothing
			newSQL.WriteRune(r)
		case r == '\'' && quoteStart:
			quoteStart = false
			newSQL.WriteRune(r)
		case r == namedArgStartByte:
			foundColon = true
			buf.WriteByte(namedArgStartByte)
			newSQL.WriteByte(placeHolderRune)
		case isNotNamedArgSeperator(r) && foundColon: // character class can be changed to allow more, like emojis
			foundColon = false
			if s := buf.String(); buf.Len() > 1 {
				qualifiedColumns = append(qualifiedColumns, s)
				found = true
			}
			buf.Reset()
			newSQL.WriteRune(r)
		case foundColon:
			buf.WriteRune(r)
		default:
			newSQL.WriteRune(r)
		}
	}
	bufferpool.Put(buf)
	return newSQL.Bytes(), qualifiedColumns, found
}

func isNamedArg(placeHolder string) (ret bool) {
	lColon := len(namedArgStartStr)
	if len(placeHolder) >= lColon && placeHolder[0:lColon] == namedArgStartStr {
		// Remove the first colon : if there is one
		placeHolder = placeHolder[lColon:]
	}
	for _, r := range placeHolder {
		if isNotNamedArgSeperator(r) {
			return false
		}
	}
	return true
}

func isNotNamedArgSeperator(r rune) bool {
	return !unicode.IsLetter(r) && !isEmoji(r) && !unicode.IsDigit(r) && r != '.'
}

// isEmoji represents one of the most important functions in this project.
func isEmoji(r rune) bool {
	return r >= 0x1F600 && r <= 0x1F64F || // Emoticons
		r >= 0x1F680 && r <= 0x1F6FF || // Transport and Map
		r >= 0x2600 && r <= 0x26FF || // Misc symbols
		r >= 0x2700 && r <= 0x27BF || // Dingbats
		r >= 0xFE00 && r <= 0xFE0F || // Variation Selectors
		r >= 65024 && r <= 65039 || // Variation selector
		r >= 0x1F900 && r <= 0x1F9FF || // Supplemental Symbols and Pictographs
		r >= 8400 && r <= 8447 || // Combining Diacritical Marks for Symbols
		r >= 0x1F300 && r <= 0x1F5FF // Misc Symbols and Pictographs
}
