package token

import "net/http"

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
	injectAuth(request *http.Request) error
}

type NoAuthInjector struct {
}

func (i NoAuthInjector) injectAuth(r *http.Request) error {
	return nil
}

func newNoAuthInjector(config AuthConfig) AuthInjector {
	return NoAuthInjector{}
}
