package aac

import "errors"

var (
	ErrClosed             = errors.New("aac: closed")
	ErrNeedMoreData       = errors.New("aac: need more data")
	ErrInvalidConfig      = errors.New("aac: invalid MPEG-4 audio config")
	ErrInvalidADTS        = errors.New("aac: invalid ADTS frame")
	ErrInvalidFLV         = errors.New("aac: invalid FLV audio tag")
	ErrInvalidFrame       = errors.New("aac: invalid audio frame")
	ErrUnsupportedFormat  = errors.New("aac: unsupported media format")
	ErrUnsupportedProfile = errors.New("aac: unsupported profile")
	ErrNativeUnavailable  = errors.New("aac: decoder unavailable")
)
