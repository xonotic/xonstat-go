package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/d0"
)

// ContextKey is a type for the keys in contexts.
type ContextKey string

// D0VerifyResultKey is the HTTP context key value used to store verification results in middleware functions.
const D0VerifyResultKey ContextKey = "D0VerifyResult"

// D0Verify is an HTTP middleware handler that checks if the content
// of a request is d0_blind_id verified via its signature and POST body
func D0Verify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keygen := viper.GetString("D0BlindIDKeyGen")
		pubkey := viper.GetString("D0BlindIDPubKey")

		signature := r.Header.Get("X-D0-Blind-Id-Detached-Signature")
		if signature == "" {
			http.Error(w, "unable to verify request: missing X-D0-Blind-Id-Detached-Signature header", http.StatusBadRequest)
			return
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "unable to verify request: malformed body", http.StatusBadRequest)
			return
		}
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(data))

		result, err := d0.Verify(keygen, pubkey, signature, "", string(data))
		if err != nil {
			http.Error(w, "unverified request", http.StatusUnauthorized)
			return
		}

		if result.IDFP == "" {
			http.Error(w, "unverified request", http.StatusUnauthorized)
			return
		}

		// Add the result to the context of the request so it can be used later.
		ctx := context.WithValue(r.Context(), D0VerifyResultKey, result)

		// Also set headers for fetching via that method too.
		r.Header.Set("X-D0-Blind-Id-CAStatus", fmt.Sprintf("%v", result.CAStatus))
		r.Header.Set("X-D0-Blind-Id-IDFP", result.IDFP)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Cached is a Redis caching middleware wherein successful requests can get stored for faster retrieval later.
func (ae *AppEnv) Cached(duration time.Duration, handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Most of the HandlerFuncs do their own content negotiation by checking the
		// Accept header instead of having a different route/URI. To account for this
		// we will tack on the value of the accept header to the cache key itself to
		// differentiate.
		accept := r.Header.Get("Accept")

		var key string
		if accept == "application/json" || accept == "text/plain" {
			key = fmt.Sprintf("%s-%s", r.RequestURI, accept)
		} else {
			key = r.RequestURI
		}

		content := ae.cache.Get(key)
		if content != nil {
			w.Write(content)
		} else {
			c := httptest.NewRecorder()
			handler(c, r)

			for k, v := range c.HeaderMap {
				w.Header()[k] = v
			}

			w.WriteHeader(c.Code)
			content := c.Body.Bytes()

			if c.Code == 200 {
				ae.cache.Set(key, content, duration)
			}

			w.Write(content)
		}
	}
}
