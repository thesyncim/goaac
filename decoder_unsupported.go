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

type coreFrameInfo struct {
	InputBytes    int
	OutputSamples int
	Channels      int
	SampleRate    int
	ObjectType    AudioObjectType
}

func (d *pureDecoder) decodeInto(dst []int16, _ []byte) ([]int16, coreFrameInfo, error) {
	return dst, coreFrameInfo{}, ErrClosed
}

func (d *pureDecoder) close() {}

func PureGoVersion() string {
	return "unavailable"
}
