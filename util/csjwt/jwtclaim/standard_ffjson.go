/*
Sniperkit-Bot
- Status: analyzed
*/

// DO NOT EDIT!
// Code generated by ffjson <https://github.com/pquerna/ffjson>
// source: standard.go
// DO NOT EDIT!

package jwtclaim

import (
	"bytes"
	"fmt"

	fflib "github.com/pquerna/ffjson/fflib/v1"
)

const (
	ffj_t_Standardbase = iota
	ffj_t_Standardno_such_key

	ffj_t_Standard_Audience

	ffj_t_Standard_ExpiresAt

	ffj_t_Standard_ID

	ffj_t_Standard_IssuedAt

	ffj_t_Standard_Issuer

	ffj_t_Standard_NotBefore

	ffj_t_Standard_Subject
)

var ffj_key_Standard_Audience = []byte("aud")

var ffj_key_Standard_ExpiresAt = []byte("exp")

var ffj_key_Standard_ID = []byte("jti")

var ffj_key_Standard_IssuedAt = []byte("iat")

var ffj_key_Standard_Issuer = []byte("iss")

var ffj_key_Standard_NotBefore = []byte("nbf")

var ffj_key_Standard_Subject = []byte("sub")

func (uj *Standard) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return uj.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

func (uj *Standard) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error = nil
	currentKey := ffj_t_Standardbase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffj_t_Standardno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'a':

					if bytes.Equal(ffj_key_Standard_Audience, kn) {
						currentKey = ffj_t_Standard_Audience
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'e':

					if bytes.Equal(ffj_key_Standard_ExpiresAt, kn) {
						currentKey = ffj_t_Standard_ExpiresAt
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'i':

					if bytes.Equal(ffj_key_Standard_IssuedAt, kn) {
						currentKey = ffj_t_Standard_IssuedAt
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffj_key_Standard_Issuer, kn) {
						currentKey = ffj_t_Standard_Issuer
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'j':

					if bytes.Equal(ffj_key_Standard_ID, kn) {
						currentKey = ffj_t_Standard_ID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'n':

					if bytes.Equal(ffj_key_Standard_NotBefore, kn) {
						currentKey = ffj_t_Standard_NotBefore
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 's':

					if bytes.Equal(ffj_key_Standard_Subject, kn) {
						currentKey = ffj_t_Standard_Subject
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.EqualFoldRight(ffj_key_Standard_Subject, kn) {
					currentKey = ffj_t_Standard_Subject
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Standard_NotBefore, kn) {
					currentKey = ffj_t_Standard_NotBefore
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffj_key_Standard_Issuer, kn) {
					currentKey = ffj_t_Standard_Issuer
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Standard_IssuedAt, kn) {
					currentKey = ffj_t_Standard_IssuedAt
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Standard_ID, kn) {
					currentKey = ffj_t_Standard_ID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Standard_ExpiresAt, kn) {
					currentKey = ffj_t_Standard_ExpiresAt
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Standard_Audience, kn) {
					currentKey = ffj_t_Standard_Audience
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffj_t_Standardno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffj_t_Standard_Audience:
					goto handle_Audience

				case ffj_t_Standard_ExpiresAt:
					goto handle_ExpiresAt

				case ffj_t_Standard_ID:
					goto handle_ID

				case ffj_t_Standard_IssuedAt:
					goto handle_IssuedAt

				case ffj_t_Standard_Issuer:
					goto handle_Issuer

				case ffj_t_Standard_NotBefore:
					goto handle_NotBefore

				case ffj_t_Standard_Subject:
					goto handle_Subject

				case ffj_t_Standardno_such_key:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_Audience:

	/* handler: uj.Audience type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.Audience = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_ExpiresAt:

	/* handler: uj.ExpiresAt type=int64 kind=int64 quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int64", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			uj.ExpiresAt = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_ID:

	/* handler: uj.ID type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.ID = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_IssuedAt:

	/* handler: uj.IssuedAt type=int64 kind=int64 quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int64", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			uj.IssuedAt = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Issuer:

	/* handler: uj.Issuer type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.Issuer = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_NotBefore:

	/* handler: uj.NotBefore type=int64 kind=int64 quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int64", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			uj.NotBefore = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Subject:

	/* handler: uj.Subject type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.Subject = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:
	return nil
}
