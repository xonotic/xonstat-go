package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

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
