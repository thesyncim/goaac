//go:build !((darwin && arm64) || (linux && arm64))

package aac

import "fmt"

type pureDecoder struct{}

func newPureDecoder(Config) (*pureDecoder, error) {
	return nil, fmt.Errorf("%w: pure Go FAAD2 port is generated for darwin/arm64 and linux/arm64", ErrNativeUnavailable)
}

func (d *pureDecoder) decode([]byte) ([]int16, error) {
	return nil, ErrClosed
}

func (d *pureDecoder) close() {}

func PureGoVersion() string {
	return "unavailable"
}
