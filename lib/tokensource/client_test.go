package tokensource

import (
	"context"
	"fmt"
	"github.com/gravitational/teleport/api/defaults"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
	"time"
)

// TestGetTokenAuthCredentialsFailures not testing injector because there is only one injector and it does not throw any error
func TestGetTokenAuthCredentialsFailures(t *testing.T) {
	resp := &struct {
		status   int
		response string
	}{
		response: "test",
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "application/json")
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
	}{
		{
			name:     "invalid template substitution",
			urlTmpl:  server.URL + "/{{.User}}",
			scheme:   NoneAuthScheme,
			httpStat: 200,
			httpResp: "{}",
			error:    "can't evaluate field User in type tokensource.urlTmplParams",
		},
		{
			name:     "400 error",
			urlTmpl:  server.URL,
			scheme:   NoneAuthScheme,
			httpStat: 400,
			httpResp: "{\"code\":\"123\", \"msg\": \"some error\"}",
			error:    "123: some error",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := template.New("abcd").Parse(tc.urlTmpl)
			if err != nil {
				t.Errorf("error while creating template: %s", tc.urlTmpl)
				return
			}
			client := NewClient(ClientConfig{
				Enabled:     true,
				UrlTemplate: tmpl,
				Timeout:     1 * time.Second,
				AuthConfig: AuthConfig{
					Scheme: tc.scheme,
				},
			})
			resp.response = tc.httpResp
			resp.status = tc.httpStat
			_, _, err = client.GetCredentials(context.Background(), "abcd", "abcd", "abcd")

			require.ErrorContains(t, err, tc.error)
		})

	}
}

func TestGetTokenAuthCredentialsSuccess(t *testing.T) {
	expUser := "abcd"
	expPass := "abcd"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(writer, "{\"response\":{\"username\":\"%s\",\"password\":\"%s\"}}", expUser, expPass)
	}))
	defer server.Close()
	tmpl, err := template.New("abcd").Parse(server.URL)
	if err != nil {
		t.Errorf("error while creating template: %s", server.URL)
		return
	}
	client := NewClient(ClientConfig{
		Enabled:     true,
		UrlTemplate: tmpl,
		Timeout:     1 * time.Second,
		AuthConfig: AuthConfig{
			Scheme: NoneAuthScheme,
		},
	})
	user, pass, err := client.GetCredentials(context.Background(), "abcd", "abcd", "abcd")

	require.Equal(t, expUser, user)
	require.Equal(t, expPass, pass)
	require.Nil(t, err)
}

func TestCheckAndSetDefaults(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		authScheme    string
		expTimeout    time.Duration
		expAuthScheme string
	}{
		{
			name:          "all defaults",
			timeout:       0,
			authScheme:    "",
			expTimeout:    defaults.DefaultIOTimeout,
			expAuthScheme: DefaultAuthScheme,
		},
		{
			name:          "default timeout",
			timeout:       0,
			authScheme:    "NONE",
			expTimeout:    defaults.DefaultIOTimeout,
			expAuthScheme: "NONE",
		},
		{
			name:          "default auth scheme",
			timeout:       1 * time.Second,
			authScheme:    "",
			expTimeout:    1 * time.Second,
			expAuthScheme: DefaultAuthScheme,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := ClientConfig{
				Timeout: test.timeout,
				AuthConfig: AuthConfig{
					Scheme: test.authScheme,
				},
			}
			config.CheckAndSetDefaults()
			require.Equal(t, test.expTimeout, config.Timeout)
			require.Equal(t, test.expAuthScheme, config.AuthConfig.Scheme)
		})
	}
}
