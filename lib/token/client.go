package token

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gravitational/trace"
	"io"
	"net"
	"net/http"
	"text/template"
	"time"
)

type Client struct {
	urlTemplate  *template.Template
	httpClient   http.Client
	timeout      time.Duration
	authInjector AuthInjector
}

type ClientConfig struct {
	UrlTmpl                        *template.Template
	ReadTimeout, ConnectionTimeout time.Duration
	AuthConfig                     AuthConfig
}

type AuthConfig struct {
	Scheme string
}

type CredentialResponse struct {
	Response Credential `json:"response"`
	Msg      string     `json:"msg"`
	Code     string     `json:"code"`
}

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type urlTmplParams struct {
	DBName   string
	Token    string
	Username string
}

func NewClient(config ClientConfig) *Client {
	dialer := net.Dialer{
		Timeout: config.ConnectionTimeout, // Connect Timeout
	}

	authInjector := AuthCreators[DefaultAuthScheme](config.AuthConfig)

	scheme := config.AuthConfig.Scheme
	if scheme != "" && AuthCreators[scheme] != nil {
		authInjector = AuthCreators[scheme](config.AuthConfig)
	}
	return &Client{
		urlTemplate: config.UrlTmpl,
		httpClient: http.Client{
			Transport: &http.Transport{
				DialContext: dialer.DialContext,
			},
		},
		timeout:      config.ReadTimeout,
		authInjector: authInjector,
	}
}

func (c *Client) GetTokenAuthCredentials(ctx context.Context, dbName, username, token string) (string, string, error) {
	buf := &bytes.Buffer{}
	if err := c.urlTemplate.Execute(buf, urlTmplParams{DBName: dbName, Token: token, Username: username}); err != nil {
		return "", "", trace.Wrap(err)
	}
	url := buf.String()

	credentialResponse, err := c.doRequest(ctx, url)
	if err != nil {
		return "", "", err
	}

	return credentialResponse.Response.Username, credentialResponse.Response.Password, nil
}

func (c *Client) doRequest(ctx context.Context, url string) (CredentialResponse, error) {
	var credentialResponse CredentialResponse

	request, err := http.NewRequest("GET", url, bytes.NewReader(nil))
	if err != nil {
		return credentialResponse, trace.Wrap(err)
	}

	timeoutContext, cancelFunction := context.WithTimeout(ctx, c.timeout)
	defer cancelFunction()

	err = c.authInjector.injectAuth(request)
	if err != nil {
		return credentialResponse, trace.Wrap(err)
	}

	/* Execute Request */
	response, err := c.httpClient.Do(request.WithContext(timeoutContext))
	if err != nil {
		return credentialResponse, trace.Wrap(err)
	}
	defer response.Body.Close()

	/* Read Body & Decode if Response came & unmarshal entity is supplied */
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return credentialResponse, trace.Wrap(err)
	}
	err = json.Unmarshal(responseBytes, &credentialResponse)
	if err != nil {
		return credentialResponse, trace.Wrap(err)
	}

	/* Check If Request was Successful */
	if response.StatusCode < http.StatusOK || response.StatusCode > http.StatusMultipleChoices {
		return credentialResponse, trace.Errorf("%v: %v", credentialResponse.Code, credentialResponse.Msg)
	}

	return credentialResponse, nil
}
