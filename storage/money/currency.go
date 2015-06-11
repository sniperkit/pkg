// Copyright 2015, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package money

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"database/sql"

	"bytes"

	"github.com/corestoreio/csfw/i18n"
	"github.com/corestoreio/csfw/utils/log"
)

var (
	ErrOverflow = errors.New("Integer Overflow")

	guard     int64   = 10000
	guardi            = int(guard)
	guardf    float64 = float64(guard)
	dp        int64   = 10000
	dpi               = int(dp)
	dpf       float64 = float64(dp)
	swedish           = Interval000
	formatter i18n.CurrencyFormatter

	RoundTo = .5
	//	RoundTo  = .5 + (1 / Guardf)
	RoundToN = RoundTo * -1
)

// Interval* constants http://en.wikipedia.org/wiki/Swedish_rounding
const (
	// Interval000 no swedish rounding (default)
	Interval000 Interval = iota
	// Interval005 rounding with 0.05 intervals
	Interval005
	// Interval010 rounding with 0.10 intervals
	Interval010
	// Interval015 same as Interval010 except that 5 will be rounded down.
	// 0.45 => 0.40 or 0.46 => 0.50
	// Special case for New Zealand (a must visit!), it is up to the business
	// to decide if they will round 5¢ intervals up or down. The majority of
	// retailers follow government advice and round it down. Use then Interval015.
	// otherwise use Interval010.
	Interval015
	// Interval025 rounding with 0.25 intervals
	Interval025
	// Interval050 rounding with 0.50 intervals
	Interval050
	// Interval100 rounding with 1.00 intervals
	Interval100
	interval999
)

// FormatJSON to define the output format. Can be combined in a binary style.
// Adding more than option generates a JSON array and not a number
const (
	// JSONFormatNumber generates only the raw number, e.g. for use in JS
	JSONFormatNumber = 1 << iota
	// JSONFormatSign generates only the currency sign (short)
	JSONFormatSign
	// JSONFormatLocale generates the locale specific formatted currency string
	JSONFormatLocale
	JSONFormatDefault = JSONFormatNumber // think about default formatting
	jsonFormatMax
)

type (
	// Interval defines the type for the Swedish rounding.
	Interval uint8

	// Currency represents a money aka currency type to avoid rounding errors with floats.
	// Takes also care of http://en.wikipedia.org/wiki/Swedish_rounding
	Currency struct {
		// m money in Guard/DP
		m int64
		// Formatter to allow language specific output formats @todo
		Formatter i18n.CurrencyFormatter

		// Valid if false the internal value is NULL
		Valid bool
		// Interval defines how the swedish rounding can be applied.
		Interval Interval

		// JSONFormat output format when marshaling
		JSONFormat int

		guard  int64
		guardf float64
		dp     int64
		dpf    float64
		// bufC print buffer for number generation incl. locale settings ... or a sync.Pool ?
		bufC []byte
		// bufJ buffer when creating JSON
		bufJ []byte
	}

	// OptionFunc used to apply options to the Currency struct
	OptionFunc func(*Currency) OptionFunc
)

func init() {
	formatter = i18n.DefaultCurrency
}

// DefaultFormatter sets the package wide default locale specific currency formatter
func DefaultFormatter(cf i18n.CurrencyFormatter) {
	formatter = cf
}

// DefaultSwedish sets the global and New() defaults swedish rounding
// http://en.wikipedia.org/wiki/Swedish_rounding
// Errors will be logged.
func DefaultSwedish(i Interval) {
	if i < interval999 {
		swedish = i
	} else {
		log.Error("money=SetSwedishRounding", "err", errors.New("Interval out of scope"), "interval", i)
	}
}

// DefaultGuard sets the global default guard. A fixed-length guard for precision arithmetic.
// Returns the successful applied value.
func DefaultGuard(g int64) int64 {
	if g == 0 {
		g = 1
	}
	guard = g
	guardf = float64(g)
	return guard
}

// DefaultPrecision sets the global default decimal precision.
// 2 decimal places => 10^2; 3 decimal places => 10^3; x decimal places => 10^x
// Returns the successful applied value.
func DefaultPrecision(p int64) int64 {
	l := int64(math.Log(float64(p)))
	if p == 0 || (p != 0 && (l%2) != 0) {
		p = dp
	}
	dp = p
	dpf = float64(p)
	return dp
}

// Swedish sets the Swedish rounding
// http://en.wikipedia.org/wiki/Swedish_rounding
// Errors will be logged
func Swedish(i Interval) OptionFunc {
	if i >= interval999 {
		log.Error("Currency=SetSwedishRounding", "err", errors.New("Interval out of scope. Resetting."), "interval", i)
		i = Interval000
	}
	return func(c *Currency) OptionFunc {
		previous := c.Interval
		c.Interval = i
		return Swedish(previous)
	}
}

// SetGuard sets the guard
func Guard(g int) OptionFunc {
	if g == 0 {
		g = 1
	}
	return func(c *Currency) OptionFunc {
		previous := int(c.guard)
		c.guard = int64(g)
		c.guardf = float64(g)
		return Guard(previous)
	}
}

// Precision sets the precision.
// 2 decimal places => 10^2; 3 decimal places => 10^3; x decimal places => 10^x
// If not a decimal power then falls back to the default value.
func Precision(p int) OptionFunc {
	p64 := int64(p)
	l := int64(math.Log(float64(p64)))
	if p64 != 0 && (l%2) != 0 {
		p64 = dp
	}
	if p64 == 0 { // check for division by zero
		p64 = 1
	}
	return func(c *Currency) OptionFunc {
		previous := int(c.dp)
		c.dp = p64
		c.dpf = float64(p64)
		return Precision(previous)
	}
}

// Formatter sets the locale specific formatter. Allows to switch quickly
// between different locales.
func Formatter(f i18n.CurrencyFormatter) OptionFunc {
	return func(c *Currency) OptionFunc {
		previous := c.Formatter
		c.Formatter = f
		return Formatter(previous)
	}
}

// JSONFormat optional option helper.
func JSONFormat(f int) OptionFunc {
	return func(c *Currency) OptionFunc {
		previous := c.JSONFormat
		c.JSONFormat = f
		return JSONFormat(previous)
	}
}

// New creates a new empty Currency struct with package default values of
// Guard and decimal precision.
func New(opts ...OptionFunc) Currency {
	c := Currency{
		guard:      guard,
		guardf:     guardf,
		dp:         dp,
		dpf:        dpf,
		Formatter:  formatter,
		JSONFormat: JSONFormatDefault,
	}
	c.Option(opts...)
	return c
}

// Options besides New() also Option() can apply options to the current
// struct. It returns the last set option. More info about the returned function:
// http://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
func (c *Currency) Option(opts ...OptionFunc) (previous OptionFunc) {
	for _, o := range opts {
		if o != nil {
			previous = o(c)
		}
	}
	return previous
}

// Abs Returns the absolute value of Currency
func (c Currency) Abs() Currency {
	if c.m < 0 {
		return c.Neg()
	}
	return c
}

// Getf gets the float64 value of money (see Raw() for int64)
func (c Currency) Getf() float64 {
	return float64(c.m) / c.dpf
}

// Geti gets value of money truncating after decimal precision (see Raw() for no truncation).
// Rounds always down
func (c Currency) Geti() int64 {
	return c.m / c.dp
}

// Dec returns the decimals
func (c Currency) Dec() int64 {
	return c.Abs().Raw() % c.dp
}

// Raw returns in int64 the value of Currency (also see Gett(), See Get() for float64)
func (c Currency) Raw() int64 {
	return c.m
}

// Set sets the raw Currency field m
func (c Currency) Set(i int64) Currency {
	c.m = i
	c.Valid = true
	return c
}

// Setf sets a float64 into a Currency type for precision calculations
func (c Currency) Setf(f float64) Currency {
	fDPf := f * c.dpf
	r := int64(f * c.dpf)
	c.Valid = true
	return c.Set(rnd(r, fDPf-float64(r)))
}

// Sign returns the Sign of Currency 1 if positive, -1 if negative
func (c Currency) Sign() int {
	if c.m < 0 {
		return -1
	}
	return 1
}

// Localize for money type representation in a specific locale. Owns the return value.
func (c Currency) Localize() []byte {
	c.bufC = c.bufC[:0]
	c.formatPrice(&c.bufC)
	c.Formatter.Localize(&c.bufC)
	return c.bufC
}

// String for money type representation i a specific locale.
func (c Currency) String() string {
	return string(c.Localize())

}

// Unformatted prints the currency without any locale specific formatting. E.g. useful in JavaScript.
func (c Currency) Unformatted() string {
	return string(c.UnformattedByte())
}

// UnformattedByte prints the currency without any locale specific formatting. Owns the result.
func (c Currency) UnformattedByte() []byte {
	c.bufC = c.bufC[:0]
	c.formatPrice(&c.bufC)
	return c.bufC
}

func (c *Currency) formatPrice(buf *[]byte) {
	i, d := c.Geti(), c.Dec()
	if c.Sign() < 0 && i == 0 && d > 0 {
		*buf = append(*buf, '-') // because Dec is always positive ...
	}
	*buf = append(*buf, fmt.Sprintf("%d.%02d", i, d)...) // @todo remove Sprintf
}

// Add Adds two Currency types. Returns empty Currency on integer overflow.
// Errors will be logged and a trace is available when the level for tracing has been set.
func (c Currency) Add(d Currency) Currency {
	r := c.m + d.m
	if (r^c.m)&(r^d.m) < 0 {
		if log.IsTrace() {
			log.Trace("Currency=Add", "err", ErrOverflow, "m", c, "n", d)
		}
		log.Error("Currency=Add", "err", ErrOverflow, "m", c, "n", d)
		return New()
	}
	c.m = r
	c.Valid = true
	return c
}

// Sub subtracts one Currency type from another. Returns empty Currency on integer overflow.
// Errors will be logged and a trace is available when the level for tracing has been set.
func (c Currency) Sub(d Currency) Currency {
	r := c.m - d.m
	if (r^c.m)&^(r^d.m) < 0 {
		if log.IsTrace() {
			log.Trace("Currency=Sub", "err", ErrOverflow, "m", c, "n", d)
		}
		log.Error("Currency=Sub", "err", ErrOverflow, "m", c, "n", d)
		return New()
	}
	c.m = r
	return c
}

// Mul Multiplies two Currency types. Both types must have the same precision.
func (c Currency) Mul(d Currency) Currency {
	return c.Set(c.m * d.m / c.dp)
}

// Div Divides one Currency type from another
func (c Currency) Div(d Currency) Currency {
	f := (c.guardf * c.dpf * float64(c.m)) / float64(d.m) / c.guardf
	i := int64(f)
	return c.Set(rnd(i, f-float64(i)))
}

// Mulf Multiplies a Currency with a float to return a money-stored type
func (c Currency) Mulf(f float64) Currency {
	i := c.m * int64(f*c.guardf*c.dpf)
	r := i / c.guard / c.dp
	return c.Set(rnd(r, float64(i)/c.guardf/c.dpf-float64(r)))
}

// Neg Returns the negative value of Currency
func (c Currency) Neg() Currency {
	if c.m != 0 {
		c.m *= -1
	}
	return c
}

// Pow is the power of Currency
func (c Currency) Pow(f float64) Currency {
	return c.Setf(math.Pow(c.Getf(), f))
}

// rnd rounds int64 remainder rounded half towards plus infinity
// trunc = the remainder of the float64 calc
// r     = the result of the int64 cal
func rnd(r int64, trunc float64) int64 {
	//fmt.Printf("RND 1 r = % v, trunc = %v RoundTo = %v\n", r, trunc, RoundTo)
	if trunc > 0 {
		if trunc >= RoundTo {
			r++
		}
	} else {
		if trunc < RoundToN {
			r--
		}
	}
	//fmt.Printf("RND 2 r = % v, trunc = %v RoundTo = %v\n", r, trunc, RoundTo)
	return r
}

// Roundx rounds a value. @todo check out to round negative numbers https://gist.github.com/pelegm/c48cff315cd223f7cf7b
func Round(f float64) float64 {
	return math.Floor(f + .5)
}

// Swedish applies the Swedish rounding. You may set the usual options.
func (c Currency) Swedish(opts ...OptionFunc) Currency {
	c.Option(opts...)
	switch c.Interval {
	case Interval005:
		// NL, SG, SA, CH, TR, CL, IE
		// 5 cent rounding
		return c.Setf(Round(c.Getf()*20) / 20) // base 5
	case Interval010:
		// New Zealand & Hong Kong
		// 10 cent rounding
		// In Sweden between 1985 and 1992, prices were rounded up for sales
		// ending in 5 öre.
		return c.Setf(Round(c.Getf()*10) / 10)
	case Interval015:
		// 10 cent rounding, special case
		// Special case: In NZ, it is up to the business to decide if they
		// will round 5¢ intervals up or down. The majority of retailers follow
		// government advice and round it down.
		if c.m%5 == 0 {
			c.m = c.m - 1
		}
		return c.Setf(Round(c.Getf()*10) / 10)
	case Interval025:
		// round to quarter
		return c.Setf(Round(c.Getf()*4) / 4)
	case Interval050:
		// 50 cent rounding
		// The system used in Sweden from 1992 to 2010, in Norway from 1993 to 2012,
		// and in Denmark since 1 October 2008 is the following:
		// Sales ending in 1–24 öre round down to 0 öre.
		// Sales ending in 25–49 öre round up to 50 öre.
		// Sales ending in 51–74 öre round down to 50 öre.
		// Sales ending in 75–99 öre round up to the next whole Krone/krona.
		return c.Setf(Round(c.Getf()*2) / 2)
	case Interval100:
		// The system used in Sweden since 30 September 2010 and used in Norway since 1 May 2012.
		// Sales ending in 1–49 öre/øre round down to 0 öre/øre.
		// Sales ending in 50–99 öre/øre round up to the next whole krona/krone.
		return c.Setf(Round(c.Getf()*1) / 1) // ;-)
	}
	return c
}

var (
	_          json.Unmarshaler = (*Currency)(nil)
	_          json.Marshaler   = (*Currency)(nil)
	_          sql.Scanner      = (*Currency)(nil)
	nullString                  = []byte("null")

//	_ driver.ValueConverter = (Currency)(nil)
//	_ driver.Valuer         = (Currency)(nil)
)

// jsonStrOrArray return true if we need an array output
func jsonStrOrArray(f int) bool {
	count := 0
	if f&JSONFormatNumber != 0 {
		count++
	}
	if f&JSONFormatSign != 0 {
		count++
	}
	if f&JSONFormatLocale != 0 {
		count++
	}
	return count > 1
}

// MarshalJSON generates a JSON string. @todo compatibility to ffjson ?
func (c Currency) MarshalJSON() ([]byte, error) {
	// @todo use interface JSONer in this package

	// @todo should be possible to output the value without the currency sign
	// or output it as an array e.g.: [1234.56, "1.234,56€", "€"]
	// hmmmm
	if false == c.Valid {
		return nullString, nil
	}

	isArray := jsonStrOrArray(c.JSONFormat)
	if false == isArray {
		if c.JSONFormat&JSONFormatNumber != 0 {
			return c.UnformattedByte(), nil
		}
		if c.JSONFormat&JSONFormatSign != 0 {
			return c.Formatter.Sign(), nil
		}
		if c.JSONFormat&JSONFormatLocale != 0 {
			// necessary to return a copy?
			// l := c.Localize()
			// ll := len(l)
			// b := make([]byte, ll, ll)
			// copy(b, l)
			// return b, nil
			return c.Localize(), nil
		}
	}

	// c.bufJ = c.bufJ[:0]

	return nil, nil
}

func (c *Currency) UnmarshalJSON(b []byte) error {
	// @todo rewrite and optimize unmarshalling but for now json.Unmarshal is fine
	var s interface{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return c.Scan(s)
}

// @todo quick write down without tests so add tests 8-)
// Errors will be logged.
func (c *Currency) Scan(value interface{}) error {
	if value == nil {
		c.m, c.Valid = 0, false
		return nil
	}
	if c.guard == 0 {
		c.Option(Guard(guardi))
	}
	if c.dp == 0 {
		c.Option(Precision(dpi))
	}

	if rb, ok := value.(*sql.RawBytes); ok {
		f, err := atof64([]byte(*rb))
		if err != nil {
			return log.Error("Currency=Scan", "err", err)
		}
		c.Valid = true
		c.Setf(f)
	}
	return nil
}

var colon = []byte(",")

func atof64(bVal []byte) (f float64, err error) {
	bVal = bytes.Replace(bVal, colon, nil, -1)
	//	s := string(bVal)
	//	s1 := strings.Replace(s, ",", "", -1)
	f, err = strconv.ParseFloat(string(bVal), 64)
	return f, err
}

//// ConvertValue @todo ?
//func (c Currency) ConvertValue(v interface{}) (driver.Value, error) {
//	return nil, nil
//}
//
//// Value @todo ?
//func (c Currency) Value() (driver.Value, error) {
//	return nil, nil
//}
