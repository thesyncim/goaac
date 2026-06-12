//go:build !cgo

package aac

type nativeDecoder struct{}

func newNativeDecoder(Config) (*nativeDecoder, error) {
	return nil, ErrNativeUnavailable
}

func (*nativeDecoder) decode([]byte) ([]int16, error) {
	return nil, ErrNativeUnavailable
}

func (*nativeDecoder) close() {}

func nativeVersion() string {
	return "unavailable"
}
