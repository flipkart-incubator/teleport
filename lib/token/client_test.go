package token

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
	"time"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		scheme           string
		expectedInjector AuthInjector
	}{
		{scheme: "NONE", expectedInjector: NoAuthInjector{}},
		{scheme: "", expectedInjector: NoAuthInjector{}},
		{scheme: "abcd", expectedInjector: NoAuthInjector{}},
	}
	for _, tc := range testCases {
		client := New(ClientConfig{
			UrlTmpl:           template.New("abcd"),
			ReadTimeout:       1 * time.Second,
			ConnectionTimeout: 1 * time.Second,
			AuthConfig: AuthConfig{
				Scheme: tc.scheme,
			},
		})
		require.IsTypef(t, client.authInjector, tc.expectedInjector, "")
	}
}

// TestGetTokenAuthCredentialsFailures cannot test http body read error because not sure how to simulate it
// Not testing injector because there is only one injector and it does not throw any error
func TestGetTokenAuthCredentialsFailures(t *testing.T) {
	resp := &struct {
		status   int
		response string
	}{
		response: "test",
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(resp.status)
		writer.Write([]byte(resp.response))
	}))
	defer server.Close()
	testCases := []struct {
		name     string
		urlTmpl  string
		scheme   string
		httpStat int
		httpResp string
		error    string
		expUser  string
		expPass  string
	}{
		{
			name:     "invalid template substitution",
			urlTmpl:  server.URL + "/{{.User}}",
			scheme:   NoneAuthScheme,
			httpStat: 200,
			httpResp: "{}",
			error:    "can't evaluate field User in type token.urlTmplParams",
			expUser:  "",
			expPass:  "",
		},
		{
			name:     "invalid request",
			urlTmpl:  string([]byte{0x7f}),
			scheme:   NoneAuthScheme,
			httpStat: 200,
			httpResp: "{}",
			error:    "invalid control character in URL",
			expUser:  "",
			expPass:  "",
		},
		{
			name:     "http error",
			urlTmpl:  "http://localhost:54323",
			scheme:   NoneAuthScheme,
			httpStat: 200,
			httpResp: "{}",
			error:    "connect: connection refused",
			expUser:  "",
			expPass:  "",
		},
		{
			name:     "unmarshal error",
			urlTmpl:  server.URL,
			scheme:   NoneAuthScheme,
			httpStat: 400,
			httpResp: "{]",
			error:    "invalid character ']'",
			expUser:  "",
			expPass:  "",
		},
		{
			name:     "400 error",
			urlTmpl:  server.URL,
			scheme:   NoneAuthScheme,
			httpStat: 400,
			httpResp: "{\"code\":\"123\", \"msg\": \"some error\"}",
			error:    "123: some error",
			expUser:  "",
			expPass:  "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := template.New("abcd").Parse(tc.urlTmpl)
			if err != nil {
				t.Errorf("error while creating template: %s", tc.urlTmpl)
				return
			}
			client := New(ClientConfig{
				UrlTmpl:           tmpl,
				ReadTimeout:       1 * time.Second,
				ConnectionTimeout: 1 * time.Second,
				AuthConfig: AuthConfig{
					Scheme: tc.scheme,
				},
			})
			resp.response = tc.httpResp
			resp.status = tc.httpStat
			user, pass, err := client.GetTokenAuthCredentials(context.Background(), "abcd", "abcd", "abcd")

			require.Equal(t, tc.expUser, user)
			require.Equal(t, tc.expPass, pass)
			require.ErrorContains(t, err, tc.error)
		})

	}
}

func TestGetTokenAuthCredentialsSuccess(t *testing.T) {
	expUser := "abcd"
	expPass := "abcd"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "{\"response\":{\"username\":\"%s\",\"password\":\"%s\"}}", expUser, expPass)
	}))
	defer server.Close()
	tmpl, err := template.New("abcd").Parse(server.URL)
	if err != nil {
		t.Errorf("error while creating template: %s", server.URL)
		return
	}
	client := New(ClientConfig{
		UrlTmpl:           tmpl,
		ReadTimeout:       1 * time.Second,
		ConnectionTimeout: 1 * time.Second,
		AuthConfig: AuthConfig{
			Scheme: NoneAuthScheme,
		},
	})
	user, pass, err := client.GetTokenAuthCredentials(context.Background(), "abcd", "abcd", "abcd")

	require.Equal(t, expUser, user)
	require.Equal(t, expPass, pass)
	require.Nil(t, err)
}
