package aac

import (
	"fmt"
	"sync"
)

type Transport int

const (
	// TransportAuto selects raw AAC when Options.Config contains enough raw
	// stream configuration, otherwise ADTS.
	TransportAuto Transport = iota
	// TransportRaw decodes raw AAC-LC access units configured by Config.
	TransportRaw
	// TransportADTS decodes AAC-LC frames with ADTS headers.
	TransportADTS
)

func (t Transport) String() string {
	switch t {
	case TransportAuto:
		return "auto"
	case TransportRaw:
		return "raw"
	case TransportADTS:
		return "adts"
	default:
		return fmt.Sprintf("transport(%d)", int(t))
	}
}

type Options struct {
	// Config is required for TransportRaw. For TransportAuto it selects raw
	// AAC when it contains ExtraData, sample rate, or channel information.
	Config Config
	// Transport selects the expected input framing. The zero value is
	// TransportAuto.
	Transport Transport
}

// FrameInfo describes one decoded AAC access unit.
type FrameInfo struct {
	// Config is the decoder configuration after this frame.
	Config Config
	// Transport is the framing used for this decode call.
	Transport Transport
	// InputBytes is the number of input bytes consumed for this frame.
	InputBytes int
	// OutputSamples is the number of interleaved int16 samples appended.
	OutputSamples int
	// SampleRate is the output sample rate reported for this frame.
	SampleRate int
	// Channels is the output channel count reported for this frame.
	Channels int
	// ObjectType is the decoded MPEG-4 audio object type.
	ObjectType AudioObjectType
	// ADTSHeader is populated when Transport is TransportADTS.
	ADTSHeader ADTSHeader
}

// Decoder decodes AAC-LC access units. A Decoder is safe for concurrent use,
// although decode calls are serialized because AAC decoder state is ordered.
type Decoder struct {
	mu     sync.Mutex
	pure   *pureDecoder
	cfg    Config
	adts   bool
	closed bool
}

// New creates a decoder from Options.
func New(opts Options) (*Decoder, error) {
	switch opts.Transport {
	case TransportAuto:
		if hasRawConfig(opts.Config) {
			return NewDecoder(opts.Config)
		}
		return NewADTSDecoder()
	case TransportRaw:
		return NewDecoder(opts.Config)
	case TransportADTS:
		return NewADTSDecoder()
	default:
		return nil, fmt.Errorf("%w: unknown transport %s", ErrInvalidConfig, opts.Transport)
	}
}

// NewDecoder creates a raw AAC-LC decoder from Config.
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

// NewADTSDecoder creates an AAC-LC decoder for ADTS-framed input.
func NewADTSDecoder() (*Decoder, error) {
	pure, err := newPureDecoder(Config{})
	if err != nil {
		return nil, err
	}
	return &Decoder{pure: pure, adts: true}, nil
}

// Transport reports the decoder input framing.
func (d *Decoder) Transport() Transport {
	if d == nil {
		return TransportAuto
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.adts {
		return TransportADTS
	}
	return TransportRaw
}

// Config returns a copy of the decoder's current stream configuration.
func (d *Decoder) Config() Config {
	if d == nil {
		return Config{}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return copyConfig(d.cfg)
}

// DecodeRaw decodes one raw AAC-LC access unit and returns newly allocated PCM.
func (d *Decoder) DecodeRaw(frame []byte) ([]int16, error) {
	if d == nil {
		return nil, ErrClosed
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil, ErrClosed
	}
	if d.adts {
		return nil, fmt.Errorf("%w: ADTS decoder cannot decode raw AAC access units", ErrInvalidConfig)
	}
	out, _, err := d.decodeRawLocked(nil, frame)
	return out, err
}

// Decode decodes one frame using the decoder's configured transport and appends
// interleaved signed 16-bit PCM samples to dst.
func (d *Decoder) Decode(dst []int16, frame []byte) ([]int16, FrameInfo, error) {
	if d == nil {
		return dst, FrameInfo{}, ErrClosed
	}
	if d.Transport() == TransportADTS {
		return d.DecodeADTSFrameInto(dst, frame)
	}
	return d.DecodeRawInto(dst, frame)
}

// DecodeRawInto decodes one raw AAC-LC access unit and appends PCM to dst.
func (d *Decoder) DecodeRawInto(dst []int16, frame []byte) ([]int16, FrameInfo, error) {
	if d == nil {
		return dst, FrameInfo{}, ErrClosed
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return dst, FrameInfo{}, ErrClosed
	}
	if d.adts {
		return dst, FrameInfo{}, fmt.Errorf("%w: ADTS decoder cannot decode raw AAC access units", ErrInvalidConfig)
	}
	return d.decodeRawLocked(dst, frame)
}

// DecodeADTSFrame decodes one ADTS frame and returns newly allocated PCM.
func (d *Decoder) DecodeADTSFrame(frame []byte) ([]int16, error) {
	out, _, err := d.DecodeADTSFrameInto(nil, frame)
	return out, err
}

// DecodeADTSFrameInto decodes one ADTS frame and appends PCM to dst.
func (d *Decoder) DecodeADTSFrameInto(dst []int16, frame []byte) ([]int16, FrameInfo, error) {
	if d == nil {
		return dst, FrameInfo{}, ErrClosed
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return dst, FrameInfo{}, ErrClosed
	}
	if !d.adts {
		return dst, FrameInfo{}, fmt.Errorf("%w: raw decoder cannot decode ADTS frames", ErrInvalidConfig)
	}
	h, err := ParseADTSHeader(frame)
	if err != nil {
		return dst, FrameInfo{}, err
	}
	if h.ObjectType != AOTAACLC {
		return dst, FrameInfo{}, fmt.Errorf("%w: %s", ErrUnsupportedProfile, h.ObjectType)
	}
	if h.RawDataBlockCount != 0 {
		return dst, FrameInfo{}, fmt.Errorf("%w: ADTS frames with %d raw data blocks are not supported", ErrInvalidADTS, h.RawDataBlockCount+1)
	}
	frameCfg := Config{
		ObjectType:      h.ObjectType,
		SampleRate:      h.SampleRate,
		SampleRateIndex: h.SampleRateIndex,
		ChannelConfig:   h.ChannelConfig,
		Channels:        h.Channels,
	}
	out, core, err := d.pure.decodeInto(dst, frame[:h.FrameLength])
	if err != nil {
		return dst, FrameInfo{}, err
	}
	d.cfg = frameCfg
	info := FrameInfo{
		Config:        copyConfig(d.cfg),
		Transport:     TransportADTS,
		InputBytes:    h.FrameLength,
		OutputSamples: core.OutputSamples,
		SampleRate:    h.SampleRate,
		Channels:      h.Channels,
		ObjectType:    h.ObjectType,
		ADTSHeader:    h,
	}
	if core.InputBytes > 0 {
		info.InputBytes = core.InputBytes
	}
	if core.SampleRate > 0 {
		info.SampleRate = core.SampleRate
	}
	if core.Channels > 0 {
		info.Channels = core.Channels
	}
	if core.ObjectType != 0 {
		info.ObjectType = core.ObjectType
	}
	return out, info, nil
}

// Close releases decoder state. It is valid to call Close more than once.
func (d *Decoder) Close() error {
	if d == nil {
		return nil
	}
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

// DecodeADTSInto decodes a complete ADTS stream and appends PCM to dst.
func DecodeADTSInto(dst []int16, data []byte) ([]int16, Config, error) {
	dec, err := NewADTSDecoder()
	if err != nil {
		return dst, Config{}, err
	}
	defer dec.Close()
	for off, frameIndex := 0, 0; off < len(data); frameIndex++ {
		h, err := ParseADTSHeader(data[off:])
		if err != nil {
			return dst, Config{}, fmt.Errorf("frame %d: %w", frameIndex, err)
		}
		frame := data[off : off+h.FrameLength]
		dst, _, err = dec.DecodeADTSFrameInto(dst, frame)
		if err != nil {
			return dst, Config{}, err
		}
		off += h.FrameLength
	}
	return dst, dec.Config(), nil
}

// DecodeADTS decodes a complete ADTS stream and returns newly allocated PCM.
func DecodeADTS(data []byte) ([]int16, Config, error) {
	return DecodeADTSInto(nil, data)
}

// NativeVersion reports the decoder core version.
func NativeVersion() string {
	return PureGoVersion()
}

func copyConfig(c Config) Config {
	c.ExtraData = append([]byte(nil), c.ExtraData...)
	return c
}

func (d *Decoder) decodeRawLocked(dst []int16, frame []byte) ([]int16, FrameInfo, error) {
	out, core, err := d.pure.decodeInto(dst, frame)
	if err != nil {
		return dst, FrameInfo{}, err
	}
	if core.SampleRate > 0 {
		d.cfg.SampleRate = core.SampleRate
	}
	if core.Channels > 0 {
		d.cfg.Channels = core.Channels
	}
	info := FrameInfo{
		Config:        copyConfig(d.cfg),
		Transport:     TransportRaw,
		InputBytes:    len(frame),
		OutputSamples: core.OutputSamples,
		SampleRate:    d.cfg.SampleRate,
		Channels:      d.cfg.Channels,
		ObjectType:    d.cfg.ObjectType,
	}
	if core.InputBytes > 0 {
		info.InputBytes = core.InputBytes
	}
	if core.ObjectType != 0 {
		info.ObjectType = core.ObjectType
	}
	return out, info, nil
}

func hasRawConfig(c Config) bool {
	return len(c.ExtraData) > 0 || c.SampleRate > 0 || c.SampleRateIndex > 0 || c.Channels > 0 || c.ChannelConfig > 0
}
