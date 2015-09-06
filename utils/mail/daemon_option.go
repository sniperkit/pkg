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

package mail

import (
	"time"

	"crypto/tls"
	"errors"
	"github.com/corestoreio/csfw/config"
	"github.com/go-gomail/gomail"
)

// DaemonOption can be used as an argument in NewDaemon to configure a daemon.
type DaemonOption func(*Daemon) DaemonOption

// SetMessageChannel sets your custom channel to listen to.
func SetMessageChannel(mailChan chan *gomail.Message) DaemonOption {
	return func(da *Daemon) DaemonOption {
		previous := da.msgChan
		da.msgChan = mailChan
		da.closed = false
		return SetMessageChannel(previous)
	}
}

// SetDialer sets a custom dialer, e.g. for TLS use.
func SetDialer(di *gomail.Dialer) DaemonOption {
	return func(da *Daemon) DaemonOption {
		previous := da.dialer
		if di == nil {
			da.lastErrs = append(da.lastErrs, errors.New("gomail.Dialer cannot be nil"))
		}
		da.dialer = di
		da.sendFunc = nil
		return SetDialer(previous)
	}
}

// SetSendFunc lets you implements your email-sending function for e.g.
// to use any other third party API provider. Setting this option
// will remove the dialer. Your implementation must handle timeouts, etc.
func SetSendFunc(sf gomail.SendFunc) DaemonOption {
	return func(da *Daemon) DaemonOption {
		previous := da.sendFunc
		if sf == nil {
			da.lastErrs = append(da.lastErrs, errors.New("gomail.SendFunc cannot be nil"))
		}
		da.sendFunc = sf
		da.dialer = nil
		return SetSendFunc(previous)
	}
}

// SetStoreConfig sets the config.Reader to the daemon.
// Default reader is config.DefaultManager
func SetConfig(cr config.Reader) DaemonOption {
	return func(da *Daemon) DaemonOption {
		previous := da.config
		if cr == nil {
			da.lastErrs = append(da.lastErrs, errors.New("config.Reader cannot be nil"))
		}
		da.config = cr
		return SetConfig(previous)
	}
}

// SetSMTPTimeout sets the time when the daemon should closes the connection
// to the SMTP server if no email was sent in the last default 30 seconds.
func SetSMTPTimeout(t time.Duration) DaemonOption {
	return func(da *Daemon) DaemonOption {
		previous := da.smtpTimeout
		if t == 0 {
			da.lastErrs = append(da.lastErrs, errors.New("Time.Duration cannot be 0")) // really?
		}
		da.smtpTimeout = t
		return SetSMTPTimeout(previous)
	}
}

// SetTLSConfig represents the TLS configuration used for the TLS (when the
// STARTTLS extension is used) or SSL connection.
var SetTLSConfig = func(c *tls.Config) DaemonOption {
	return func(da *Daemon) DaemonOption {

		if nil == da.dialer {
			da.lastErrs = append(da.lastErrs, errors.New("Dialer is nil."))
			return SetTLSConfig(nil)
		}

		if false == da.dialer.SSL {
			da.lastErrs = append(da.lastErrs, errors.New("SSL not active."))
			return SetTLSConfig(nil)
		}

		previous := da.dialer.TLSConfig

		if nil == c {
			da.lastErrs = append(da.lastErrs, errors.New("*tls.Config cannot be nil"))
		}
		da.dialer.TLSConfig = c
		return SetTLSConfig(previous)
	}
}

// SetScope sets the config scope which can be default, website or store.
// Default scope is 0 = admin.
func SetScope(s config.ScopeIDer) DaemonOption {
	// if we have 50 stores ... each with a different mail setting then you
	// have 50 daemons. Or if we have 50 stores we must figure out which
	// stores uses the same smtp settings to avoid spinning up 50 daemons
	// and each for the same SMTP setting.
	return func(da *Daemon) DaemonOption {
		previous := da.scopeID
		if s == nil {
			da.lastErrs = append(da.lastErrs, errors.New("config.ScopeIDer cannot be nil"))
		}
		da.scopeID = s
		return SetScope(previous)
	}
}
