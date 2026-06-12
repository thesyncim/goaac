package aac

import "errors"

var (
	ErrClosed             = errors.New("aac: decoder is closed")
	ErrNeedMoreData       = errors.New("aac: need more data")
	ErrInvalidConfig      = errors.New("aac: invalid MPEG-4 audio config")
	ErrInvalidADTS        = errors.New("aac: invalid ADTS frame")
	ErrUnsupportedProfile = errors.New("aac: unsupported profile")
	ErrNativeUnavailable  = errors.New("aac: decoder unavailable")
)
