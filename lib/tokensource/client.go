package tokensource

import (
	"bytes"
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/gravitational/teleport/api/defaults"
	tracehttp "github.com/gravitational/teleport/api/observability/tracing/http"
	"github.com/gravitational/trace"
	"net/http"
	"text/template"
	"time"
)

type Client struct {
	urlTemplate  *template.Template
	httpClient   *resty.Client
	timeout      time.Duration
	authInjector AuthInjector
}

type ClientConfig struct {
	Enabled     bool
	UrlTemplate *template.Template
	Timeout     time.Duration
	AuthConfig  AuthConfig
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
	if !config.Enabled {
		return nil
	}

	config.CheckAndSetDefaults()

	client := resty.New().
		SetTransport(tracehttp.NewTransport(http.DefaultTransport)).
		SetTimeout(config.Timeout)

	scheme := config.AuthConfig.Scheme
	authInjector := AuthCreators[scheme](config.AuthConfig)
	authInjector.injectAuth(client)

	return &Client{
		urlTemplate:  config.UrlTemplate,
		httpClient:   client,
		authInjector: authInjector,
	}
}

func (c *Client) GetCredentials(ctx context.Context, dbName, username, token string) (string, string, error) {
	url, err := c.buildUrl(dbName, username, token)
	if err != nil {
		return "", "", trace.Wrap(err)
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetResult(&CredentialResponse{}).
		SetError(&CredentialResponse{}).
		Get(url)
	if err != nil {
		return "", "", trace.Wrap(err)
	}

	if resp.IsSuccess() {
		credentialResponse := resp.Result().(*CredentialResponse)
		return credentialResponse.Response.Username, credentialResponse.Response.Password, nil
	} else {
		errorResp := resp.Error().(*CredentialResponse)
		return "", "", trace.Errorf("%v: %v", errorResp.Code, errorResp.Msg)
	}
}

func (c *Client) buildUrl(dbName, username, token string) (string, error) {
	buf := &bytes.Buffer{}
	if err := c.urlTemplate.Execute(buf, urlTmplParams{DBName: dbName, Token: token, Username: username}); err != nil {
		return "", trace.Wrap(err)
	}
	return buf.String(), nil
}

func (c *ClientConfig) CheckAndSetDefaults() {
	if c.Timeout == 0 {
		c.Timeout = defaults.DefaultIOTimeout
	}
	if c.AuthConfig.Scheme == "" {
		c.AuthConfig.Scheme = DefaultAuthScheme
	}
}
