/*
Sniperkit-Bot
- Status: analyzed
*/

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

package ratelimit

import (
	"math"
	"net/http"
	"strconv"

	"github.com/corestoreio/errors"
	"github.com/corestoreio/log"
	loghttp "github.com/corestoreio/log/http"
	"gopkg.in/throttled/throttled.v2"
)

// WithRateLimit wraps an http.Handler to limit incoming requests. Requests that
// are not limited will be passed to the handler unchanged.  Limited requests
// will be passed to the DeniedHandler. X-RateLimit-Limit,
// X-RateLimit-Remaining, X-RateLimit-Reset and Retry-After headers will be
// written to the response based on the values in the RateLimitResult. The next
// handler may check an error with FromContextRateLimit().
func (s *Service) WithRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scpCfg, err := s.configByContext(r.Context())
		if err != nil {
			if s.Log.IsDebug() {
				s.Log.Debug("ratelimit.Service.WithRateLimit.configByContext", log.Err(err), loghttp.Request("request", r))
			}
			s.ErrorHandler(errors.Wrap(err, "ratelimit.Service.WithRateLimit.configFromContext")).ServeHTTP(w, r)
			return
		}
		if scpCfg.Disabled {
			if s.Log.IsDebug() {
				s.Log.Debug("ratelimit.Service.WithRateLimit.Disabled", log.Stringer("scope", scpCfg.ScopeID), log.Object("scpCfg", scpCfg), loghttp.Request("request", r))
			}
			next.ServeHTTP(w, r)
			return
		}

		isLimited, rlResult, err := scpCfg.requestRateLimit(r)
		if s.Log.IsDebug() {
			s.Log.Debug("ratelimit.Service.WithRateLimit.requestRateLimit",
				log.Err(err),
				log.Bool("is_limited", isLimited),
				log.Object("rate_limit_result", rlResult),
				log.Stringer("requested_scope", scpCfg.ScopeID),
				loghttp.Request("request", r),
			)
		}
		if err != nil {
			scpCfg.ErrorHandler(errors.Wrap(err, "[ratelimit] scpCfg.RateLimit")).ServeHTTP(w, r)
			return
		}

		setRateLimitHeaders(w, rlResult)

		if isLimited {
			// prevents a race condition in tests when calling DeniedHandler this way.
			scpCfg.DeniedHandler.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func setRateLimitHeaders(w http.ResponseWriter, rlr throttled.RateLimitResult) {
	if v := rlr.Limit; v >= 0 {
		w.Header().Add("X-RateLimit-Limit", strconv.Itoa(v))
	}

	if v := rlr.Remaining; v >= 0 {
		w.Header().Add("X-RateLimit-Remaining", strconv.Itoa(v))
	}

	if v := rlr.ResetAfter; v >= 0 {
		vi := int(math.Ceil(v.Seconds()))
		w.Header().Add("X-RateLimit-Reset", strconv.Itoa(vi))
	}

	if v := rlr.RetryAfter; v >= 0 {
		vi := int(math.Ceil(v.Seconds()))
		w.Header().Add("Retry-After", strconv.Itoa(vi))
	}
}
