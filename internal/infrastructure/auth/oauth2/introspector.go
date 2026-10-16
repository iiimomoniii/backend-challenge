package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	domainauth "github.com/iiimomoniii/backend-challenge/internal/domain/auth"
)

// Introspector verifies bearer access tokens through an OAuth 2.0 token
// introspection endpoint (RFC 7662). The authorization server issues tokens;
// this API only validates them.
type Introspector struct {
	endpoint     string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewIntrospector(endpoint, clientID, clientSecret string, allowInsecureHTTP bool, httpClient *http.Client) (*Introspector, error) {
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil || parsed.Host == "" {
		return nil, errors.New("OAuth2 introspection URL must be an absolute URL")
	}
	if parsed.Scheme != "https" && !(allowInsecureHTTP && parsed.Scheme == "http" && isLoopbackHost(parsed.Hostname())) {
		return nil, errors.New("OAuth2 introspection URL must use HTTPS; HTTP is allowed only for loopback when OAUTH2_ALLOW_INSECURE_HTTP=true")
	}
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("OAuth2 introspection client ID and secret are required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Introspector{
		endpoint: endpoint, clientID: clientID, clientSecret: clientSecret, httpClient: httpClient,
	}, nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

var _ domainauth.TokenVerifier = (*Introspector)(nil)

type introspectionResponse struct {
	Active   bool   `json:"active"`
	Subject  string `json:"sub"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Expires  int64  `json:"exp"`
}

func (i *Introspector) Verify(ctx context.Context, accessToken string) (*domainauth.AuthPayload, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, errors.New("empty OAuth2 access token")
	}

	form := url.Values{}
	form.Set("token", accessToken)
	form.Set("token_type_hint", "access_token")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, i.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create introspection request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(i.clientID, i.clientSecret)

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call OAuth2 introspection endpoint: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("OAuth2 introspection endpoint returned status %d", resp.StatusCode)
	}

	var result introspectionResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode OAuth2 introspection response: %w", err)
	}
	if !result.Active {
		return nil, errors.New("OAuth2 access token is inactive")
	}

	userID := result.Subject
	if userID == "" {
		userID = result.Username
	}
	if userID == "" {
		return nil, errors.New("OAuth2 introspection response has no sub or username")
	}

	payload := &domainauth.AuthPayload{UserID: userID, Email: result.Email}
	if result.Expires > 0 {
		payload.ExpiresAt = time.Unix(result.Expires, 0).UTC()
	}
	return payload, nil
}
