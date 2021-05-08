package prosody_httpupload

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
)

const v1Arg = "v"
const v2Arg = "v2"

// authenticator validates that requests come from a trusted client by checking a shared secret, following the
//specification in https://modules.prosody.im/mod_http_upload_external.html.
type authenticator struct {
	Secret       string
}

// newAuthenticator returns an authenticator instance configured to check against the given secret
func newAuthenticator(secret string) *authenticator {
	return &authenticator{
		Secret: secret,
	}
}

// authenticate is a middleware that authenticates an http request. Upon success, the underlying handler is called.
// If authentication is not successful, http.StatusForbidden is returned. Other 4XX codes might also be returned in
// certain situations.
func (a *authenticator) authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		// First identify the version, then attempt to validate it. This prevents doubly-authenticated requests.
		if a.isV2(r) {
			if a.authenticateV2(r) {
				handler.ServeHTTP(rw, r)
				return
			}
		} else if a.isV1(r) {
			if a.authenticateV1(r) {
				handler.ServeHTTP(rw, r)
				return
			}
		}

		// Authentication did not succeed, return 403
		rw.WriteHeader(http.StatusForbidden)
	})
}

// isV1 returns whether there is a 'v' parameter in the request
func (a *authenticator) isV1(r *http.Request) bool {
	return r.Form.Get(v1Arg) != ""
}

// authenticateV1 performs the Version 1 authentication process
func (a *authenticator) authenticateV1(r *http.Request) bool {
	v := r.Form.Get(v1Arg)
	if v == "" {
		return false
	}

	cl := a.contentLength(r)
	if cl == "" {
		return false
	}

	hash := a.hmac(r.URL.Path + " " + cl)

	return v == hash
}

// isV1 returns whether there is a 'v2' parameter in the request
func (a *authenticator) isV2(r *http.Request) bool {
	return r.Form.Get(v2Arg) != ""
}

// authenticateV2 performs the Version 1 authentication process
func (a *authenticator) authenticateV2(r *http.Request) bool {
	v := r.Form.Get(v2Arg)
	if v == "" {
		return false
	}

	cl := a.contentLength(r)
	if cl == "" {
		return false
	}

	hash := a.hmac(r.URL.Path + "\000" + cl + "\000" + r.Header.Get("Content-Type"))

	return v == hash
}

func (a *authenticator) contentLength(r *http.Request) string {
	cl := r.Header.Get("Content-Length")
	_, err := strconv.ParseUint(cl, 10, 64)
	if err != nil {
		return ""
	}

	return cl
}

func (a *authenticator) hmac(payload string) string {
	h := hmac.New(sha256.New, []byte(a.Secret))
	h.Write([]byte(payload))

	return hex.EncodeToString(h.Sum(nil))
}
