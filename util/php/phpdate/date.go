// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package phpdate

import "bytes"

// Day	---	---
// d	Day of the month, 2 digits with leading zeros	01 to 31
// D	A textual representation of a day, three letters	Mon through Sun
// j	Day of the month without leading zeros	1 to 31
// l (lowercase 'L')	A full textual representation of the day of the week	Sunday through Saturday
// N	ISO-8601 numeric representation of the day of the week (added in PHP 5.1.0)	1 (for Monday) through 7 (for Sunday)
// S	English ordinal suffix for the day of the month, 2 characters	st, nd, rd or th. Works well with j
// w	Numeric representation of the day of the week	0 (for Sunday) through 6 (for Saturday)
// z	The day of the year (starting from 0)	0 through 365
// Week	---	---
// W	ISO-8601 week number of year, weeks starting on Monday (added in PHP 4.1.0)	Example: 42 (the 42nd week in the year)
// Month	---	---
// F	A full textual representation of a month, such as January or March	January through December
// m	Numeric representation of a month, with leading zeros	01 through 12
// M	A short textual representation of a month, three letters	Jan through Dec
// n	Numeric representation of a month, without leading zeros	1 through 12
// t	Number of days in the given month	28 through 31
// Year	---	---
// L	Whether it's a leap year	1 if it is a leap year, 0 otherwise.
// o	ISO-8601 year number. This has the same value as Y, except that if the ISO week number (W) belongs to the previous or next year, that year is used instead. (added in PHP 5.1.0)	Examples: 1999 or 2003
// Y	A full numeric representation of a year, 4 digits	Examples: 1999 or 2003
// y	A two digit representation of a year	Examples: 99 or 03
// Time	---	---
// a	Lowercase Ante meridiem and Post meridiem	am or pm
// A	Uppercase Ante meridiem and Post meridiem	AM or PM
// B	Swatch Internet time	000 through 999
// g	12-hour format of an hour without leading zeros	1 through 12
// G	24-hour format of an hour without leading zeros	0 through 23
// h	12-hour format of an hour with leading zeros	01 through 12
// H	24-hour format of an hour with leading zeros	00 through 23
// i	Minutes with leading zeros	00 to 59
// s	Seconds, with leading zeros	00 through 59
// u	Microseconds (added in PHP 5.2.2). Note that date() will always generate 000000 since it takes an integer parameter, whereas DateTime::format() does support microseconds if DateTime was created with microseconds.	Example: 654321
// Timezone	---	---
// e	Timezone identifier (added in PHP 5.1.0)	Examples: UTC, GMT, Atlantic/Azores
// I (capital i)	Whether or not the date is in daylight saving time	1 if Daylight Saving Time, 0 otherwise.
// O	Difference to Greenwich time (GMT) in hours	Example: +0200
// P	Difference to Greenwich time (GMT) with colon between hours and minutes (added in PHP 5.1.3)	Example: +02:00
// T	Timezone abbreviation	Examples: EST, MDT ...
// Z	Timezone offset in seconds. The offset for timezones west of UTC is always negative, and for those east of UTC is always positive.	-43200 through 50400
// Full Date/Time	---	---
// c	ISO 8601 date (added in PHP 5)	2004-02-12T15:19:21+00:00
// r	» RFC 2822 formatted date	Example: Thu, 21 Dec 2000 16:01:07 +0200
// U	Seconds since the Unix Epoch (January 1 1970 00:00:00 GMT)	See also time()

var phpToGoMap = map[rune]string{
	// it is absolutely safe to read from this map in concurrency cases
	'd': "02",
	'D': "Mon",
	'j': "2",
	'l': "Monday",
	// 'S':"only possible if we extend the time.Format function",
	'w': "TODO",
	'z': "TODO",
	'W': "TODO",

	'F': "January",
	'm': "01",
	'M': "Jan",
	'n': "1",
	't': "TODO",

	'L': "TODO",
	'Y': "2006",
	'y': "06",

	'a': "pm",
	'A': "PM",
	'g': "3",
	'G': "15",
	'h': "03",
	'H': "15",
	'i': "04",
	's': "05",
	'u': ".9",

	'e': "MST",
	'T': "MST",
	'O': "-0700",
	'P': "-07:00",
	'Z': "-070000",
}

const escape = '\\'

// ToGoFormat maps the PHP date format to the Go format.
// Not all PHP formats are supported
// http://php.net/manual/en/function.date.php
func ToGoFormat(phpFormat string) string {
	// tutorial: http://jeremy.marzhillstudios.com/io/
	var buf bytes.Buffer
	var isEscape bool
	for _, r := range phpFormat {
		if isEscape {
			// this is the following character after an escape
			buf.WriteRune(r)
			isEscape = false
			continue
		}
		if r == escape {
			isEscape = true
			continue
		}

		if gf, ok := phpToGoMap[r]; ok {
			buf.WriteString(gf)
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
