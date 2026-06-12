package aac

import (
	"fmt"
	"sync"
)

type Decoder struct {
	mu     sync.Mutex
	pure   *pureDecoder
	cfg    Config
	adts   bool
	closed bool
}

func NewDecoder(cfg Config) (*Decoder, error) {
	normalized, err := normalizeRawConfig(cfg)
	if err != nil {
		return nil, err
	}
	pure, err := newPureDecoder(normalized)
	if err != nil {
		return nil, err
	}
	return &Decoder{pure: pure, cfg: normalized}, nil
}

func NewADTSDecoder() (*Decoder, error) {
	pure, err := newPureDecoder(Config{})
	if err != nil {
		return nil, err
	}
	return &Decoder{pure: pure, adts: true}, nil
}

func (d *Decoder) Config() Config {
	d.mu.Lock()
	defer d.mu.Unlock()
	return copyConfig(d.cfg)
}

func (d *Decoder) DecodeRaw(frame []byte) ([]int16, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil, ErrClosed
	}
	if d.adts {
		return nil, fmt.Errorf("%w: ADTS decoder cannot decode raw AAC access units", ErrInvalidConfig)
	}
	return d.pure.decode(frame)
}

func (d *Decoder) DecodeADTSFrame(frame []byte) ([]int16, error) {
	h, err := ParseADTSHeader(frame)
	if err != nil {
		return nil, err
	}
	if h.ObjectType != AOTAACLC {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProfile, h.ObjectType)
	}
	if h.RawDataBlockCount != 0 {
		return nil, fmt.Errorf("%w: ADTS frames with %d raw data blocks are not supported", ErrInvalidADTS, h.RawDataBlockCount+1)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil, ErrClosed
	}
	if !d.adts && d.cfg.ObjectType != AOTAACLC {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProfile, d.cfg.ObjectType)
	}
	d.cfg = Config{
		ObjectType:      h.ObjectType,
		SampleRate:      h.SampleRate,
		SampleRateIndex: h.SampleRateIndex,
		ChannelConfig:   h.ChannelConfig,
		Channels:        h.Channels,
	}
	return d.pure.decode(frame[:h.FrameLength])
}

func (d *Decoder) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	if d.pure != nil {
		d.pure.close()
		d.pure = nil
	}
	return nil
}

func DecodeADTS(data []byte) ([]int16, Config, error) {
	frames, err := SplitADTSFrames(data)
	if err != nil {
		return nil, Config{}, err
	}
	dec, err := NewADTSDecoder()
	if err != nil {
		return nil, Config{}, err
	}
	defer dec.Close()
	var pcm []int16
	for _, frame := range frames {
		out, err := dec.DecodeADTSFrame(frame.Data)
		if err != nil {
			return nil, Config{}, err
		}
		pcm = append(pcm, out...)
	}
	return pcm, dec.Config(), nil
}

func NativeVersion() string {
	return PureGoVersion()
}

func copyConfig(c Config) Config {
	c.ExtraData = append([]byte(nil), c.ExtraData...)
	return c
}
