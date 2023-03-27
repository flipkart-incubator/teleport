package tokensource

import (
	"github.com/go-resty/resty/v2"
)

const (
	NoneAuthScheme = "NONE"
)

const DefaultAuthScheme = NoneAuthScheme

type AuthCreator func(config AuthConfig) AuthInjector

var AuthCreators = make(map[string]AuthCreator)

func init() {
	AuthCreators[NoneAuthScheme] = newNoAuthInjector
}

type AuthInjector interface {
	injectAuth(c *resty.Client)
}

type NoAuthInjector struct {
}

func (i NoAuthInjector) injectAuth(c *resty.Client) {
}

func newNoAuthInjector(config AuthConfig) AuthInjector {
	return NoAuthInjector{}
}
