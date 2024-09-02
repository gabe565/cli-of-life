package util

import "net/http"

func NewUserAgentTransport(userAgent string) *UserAgentTransport {
	return &UserAgentTransport{
		transport: http.DefaultTransport,
		userAgent: userAgent,
	}
}

type UserAgentTransport struct {
	transport http.RoundTripper
	userAgent string
}

func (u *UserAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", u.userAgent)
	return u.transport.RoundTrip(r)
}
