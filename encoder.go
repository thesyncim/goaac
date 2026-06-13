package aac

import (
	"fmt"
	"sync"

	"github.com/thesyncim/goaac/internal/fdkaac"
)

const (
	encoderSamplesPerFrame = 1024
	encoderMaxChannels     = 2
	encoderRawBufferBytes  = 8192
)

// EncoderOptions configures an AAC-LC encoder.
type EncoderOptions struct {
	// Config selects the output sample rate and mono/stereo channel layout.
	Config Config
	// BitRate is the target stream bitrate in bits per second. The zero value
	// selects 64 kbit/s per channel.
	BitRate int
	// Transport selects the framing returned by Encode. TransportAuto uses raw
	// AAC access units.
	Transport Transport
}

// EncodedFrameInfo describes one encoded AAC-LC frame.
type EncodedFrameInfo struct {
	Config               Config
	Transport            Transport
	InputSamples         int
	OutputBytes          int
	PayloadBytes         int
	SampleRate           int
	Channels             int
	ObjectType           AudioObjectType
	BitRate              int
	TotalBits            int
	UsedDynBits          int
	FillBits             int
	AlignBits            int
	BitReservoir         int
	ADTSHeaderBytes      int
	TransportStaticBits  int
	QuantizationPasses   int
	QuantizationDone     bool
	QuantizedElements    int
	ChannelElements      int
	GlobalFillExtensions int
}

// Encoder encodes interleaved signed 16-bit PCM into AAC-LC. Calls are safe for
// concurrent use, although encoding is serialized because AAC encoder state is
// ordered.
type Encoder struct {
	mu               sync.Mutex
	state            fdkaac.AACEncFrameState
	scratch          fdkaac.AACEncFrameScratch
	transportScratch fdkaac.TransportScratch
	cfg              Config
	mode             fdkaac.ChannelMode
	bitRate          int
	transport        Transport
	closed           bool
	planar           [encoderMaxChannels * encoderSamplesPerFrame]int16
	raw              [encoderRawBufferBytes]byte
}

// NewEncoder creates a reusable AAC-LC encoder.
func NewEncoder(opts EncoderOptions) (*Encoder, error) {
	transport := opts.Transport
	switch transport {
	case TransportAuto:
		transport = TransportRaw
	case TransportRaw, TransportADTS:
	default:
		return nil, fmt.Errorf("%w: unknown encoder transport %s", ErrInvalidConfig, opts.Transport)
	}

	cfg, err := normalizeRawConfig(opts.Config)
	if err != nil {
		return nil, err
	}
	if cfg.Channels < 1 || cfg.Channels > encoderMaxChannels {
		return nil, fmt.Errorf("%w: encoder supports mono/stereo AAC-LC, got %d channels", ErrUnsupportedFormat, cfg.Channels)
	}
	mode, err := fdkaacEncoderChannelMode(cfg.ChannelConfig)
	if err != nil {
		return nil, err
	}

	bitRate := opts.BitRate
	if bitRate == 0 {
		bitRate = 64000 * cfg.Channels
	}
	if bitRate <= 0 {
		return nil, fmt.Errorf("%w: invalid encoder bitrate %d", ErrInvalidConfig, bitRate)
	}

	var encCfg fdkaac.AACEncConfig
	fdkaac.FDKaacEncAacInitDefaultConfig(&encCfg)
	encCfg.SampleRate = cfg.SampleRate
	encCfg.BitRate = bitRate
	encCfg.AOT = fdkaac.AOTAACLC
	encCfg.NChannels = cfg.Channels
	encCfg.ChannelMode = mode
	encCfg.FrameLength = encoderSamplesPerFrame
	encCfg.UseRequant = 1

	e := &Encoder{
		cfg:       cfg,
		mode:      mode,
		bitRate:   bitRate,
		transport: transport,
	}
	if errCode := fdkaac.FDKaacEncInitRawFrameState(&e.state, encCfg); errCode != fdkaac.AACEncOK {
		return nil, encoderCodeError(errCode)
	}
	return e, nil
}

// Transport reports the framing returned by Encode.
func (e *Encoder) Transport() Transport {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.transport
}

// Config returns a copy of the encoder's MPEG-4 audio config.
func (e *Encoder) Config() Config {
	e.mu.Lock()
	defer e.mu.Unlock()
	return copyConfig(e.cfg)
}

// BitRate reports the configured target bitrate in bits per second.
func (e *Encoder) BitRate() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.bitRate
}

// Encode encodes one PCM frame using the encoder's configured transport.
func (e *Encoder) Encode(dst []byte, pcm []int16) ([]byte, EncodedFrameInfo, error) {
	if e.Transport() == TransportADTS {
		return e.EncodeADTSFrameInto(dst, pcm)
	}
	return e.EncodeRawInto(dst, pcm)
}

// EncodeRaw encodes one PCM frame and returns a newly allocated raw AAC access
// unit.
func (e *Encoder) EncodeRaw(pcm []int16) ([]byte, error) {
	out, _, err := e.EncodeRawInto(nil, pcm)
	return out, err
}

// EncodeRawInto encodes one PCM frame and appends a raw AAC access unit to dst.
func (e *Encoder) EncodeRawInto(dst []byte, pcm []int16) ([]byte, EncodedFrameInfo, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return dst, EncodedFrameInfo{}, ErrClosed
	}
	before := len(dst)
	out, result, err := e.encodeRawPayloadLocked(dst, pcm)
	if err != nil {
		return dst, EncodedFrameInfo{}, err
	}
	info := e.encodedFrameInfo(TransportRaw, len(out)-before, len(out)-before, 0, result)
	return out, info, nil
}

// EncodeADTSFrame encodes one PCM frame and returns a newly allocated ADTS
// frame.
func (e *Encoder) EncodeADTSFrame(pcm []int16) ([]byte, error) {
	out, _, err := e.EncodeADTSFrameInto(nil, pcm)
	return out, err
}

// EncodeADTSFrameInto encodes one PCM frame and appends an ADTS-framed AAC
// frame to dst.
func (e *Encoder) EncodeADTSFrameInto(dst []byte, pcm []int16) ([]byte, EncodedFrameInfo, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return dst, EncodedFrameInfo{}, ErrClosed
	}
	raw, result, err := e.encodeRawPayloadLocked(e.raw[:0], pcm)
	if err != nil {
		return dst, EncodedFrameInfo{}, err
	}

	before := len(dst)
	dst, err = fdkaac.AppendADTSHeaderWithScratch(dst, e.coderConfig(), len(raw), 0x7ff, &e.transportScratch)
	if err != nil {
		return dst, EncodedFrameInfo{}, err
	}
	headerBytes := len(dst) - before
	dst = append(dst, raw...)
	info := e.encodedFrameInfo(TransportADTS, len(dst)-before, len(raw), headerBytes, result)
	return dst, info, nil
}

// AppendFLVSequenceHeader appends the FLV/RTMP AAC sequence header for this
// encoder.
func (e *Encoder) AppendFLVSequenceHeader(dst []byte) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return dst, ErrClosed
	}
	return AppendFLVAACSequenceHeader(dst, e.cfg)
}

// AppendRTMPSequenceHeader appends the RTMP AAC sequence-header payload for
// this encoder.
func (e *Encoder) AppendRTMPSequenceHeader(dst []byte) ([]byte, error) {
	return e.AppendFLVSequenceHeader(dst)
}

// EncodeFLVTagInto encodes one PCM frame and appends an FLV AAC raw tag body.
func (e *Encoder) EncodeFLVTagInto(dst []byte, pcm []int16) ([]byte, EncodedFrameInfo, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return dst, EncodedFrameInfo{}, ErrClosed
	}
	raw, result, err := e.encodeRawPayloadLocked(e.raw[:0], pcm)
	if err != nil {
		return dst, EncodedFrameInfo{}, err
	}
	before := len(dst)
	dst = AppendFLVAACRawTag(dst, raw)
	info := e.encodedFrameInfo(TransportRaw, len(dst)-before, len(raw), 0, result)
	return dst, info, nil
}

// EncodeRTMPMessageInto encodes one PCM frame and appends an RTMP AAC audio
// message payload. RTMP uses the same body layout as an FLV AAC audio tag.
func (e *Encoder) EncodeRTMPMessageInto(dst []byte, pcm []int16) ([]byte, EncodedFrameInfo, error) {
	return e.EncodeFLVTagInto(dst, pcm)
}

// Close releases encoder state. It is valid to call Close more than once.
func (e *Encoder) Close() error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closed = true
	return nil
}

func (e *Encoder) encodeRawPayloadLocked(dst []byte, pcm []int16) ([]byte, fdkaac.AACEncFrameResult, error) {
	if err := e.preparePCM(pcm); err != nil {
		return dst, fdkaac.AACEncFrameResult{}, err
	}
	out, result, errCode := fdkaac.FDKaacEncEncodeFrameRaw(
		&e.state,
		dst,
		e.planar[:e.cfg.Channels*encoderSamplesPerFrame],
		encoderSamplesPerFrame,
		&e.scratch,
	)
	if errCode != fdkaac.AACEncOK {
		return dst, result, encoderCodeError(errCode)
	}
	return out, result, nil
}

func (e *Encoder) preparePCM(pcm []int16) error {
	want := e.cfg.Channels * encoderSamplesPerFrame
	if len(pcm) != want {
		return fmt.Errorf("%w: got %d interleaved samples, want %d", ErrInvalidFrame, len(pcm), want)
	}
	if e.cfg.Channels == 1 {
		copy(e.planar[:encoderSamplesPerFrame], pcm)
		return nil
	}
	for i := 0; i < encoderSamplesPerFrame; i++ {
		base := i * e.cfg.Channels
		e.planar[i] = pcm[base]
		e.planar[encoderSamplesPerFrame+i] = pcm[base+1]
	}
	return nil
}

func (e *Encoder) encodedFrameInfo(transport Transport, outputBytes int, payloadBytes int, headerBytes int, result fdkaac.AACEncFrameResult) EncodedFrameInfo {
	return EncodedFrameInfo{
		Config:               e.cfg,
		Transport:            transport,
		InputSamples:         e.cfg.Channels * encoderSamplesPerFrame,
		OutputBytes:          outputBytes,
		PayloadBytes:         payloadBytes,
		SampleRate:           e.cfg.SampleRate,
		Channels:             e.cfg.Channels,
		ObjectType:           e.cfg.ObjectType,
		BitRate:              e.bitRate,
		TotalBits:            result.TotalBits,
		UsedDynBits:          result.UsedDynBits,
		FillBits:             result.FillBits,
		AlignBits:            result.AlignBits,
		BitReservoir:         result.BitReservoir,
		ADTSHeaderBytes:      headerBytes,
		TransportStaticBits:  result.TransportStaticBits,
		QuantizationPasses:   result.QCMain.QuantizationPasses,
		QuantizationDone:     result.QCMain.QuantizationDone != 0,
		QuantizedElements:    result.QCMain.QuantizedElements,
		ChannelElements:      result.Write.ChannelElements,
		GlobalFillExtensions: result.Write.GlobalExtensions,
	}
}

func (e *Encoder) coderConfig() fdkaac.CoderConfig {
	return fdkaac.CoderConfig{
		AOT:             fdkaac.AOTAACLC,
		ChannelMode:     e.mode,
		SamplingRate:    e.cfg.SampleRate,
		SamplesPerFrame: encoderSamplesPerFrame,
		NSubFrames:      1,
		Flags:           fdkaac.ConfigFlagMPEGID,
	}
}

func fdkaacEncoderChannelMode(channelConfig int) (fdkaac.ChannelMode, error) {
	switch channelConfig {
	case 1:
		return fdkaac.Mode1, nil
	case 2:
		return fdkaac.Mode2, nil
	default:
		return fdkaac.ModeInvalid, fmt.Errorf("%w: encoder supports mono/stereo AAC-LC, got channel config %d", ErrUnsupportedFormat, channelConfig)
	}
}

func encoderCodeError(code int) error {
	switch code {
	case fdkaac.AACEncInvalidHandle:
		return fmt.Errorf("%w: invalid encoder state", ErrEncode)
	case fdkaac.AACEncInvalidFrameLength:
		return fmt.Errorf("%w: unsupported frame length", ErrInvalidConfig)
	case fdkaac.AACEncUnsupportedBitrate, fdkaac.AACEncUnsupportedBitrateMode, fdkaac.AACEncInvalidChannelBitrate:
		return fmt.Errorf("%w: unsupported encoder bitrate", ErrInvalidConfig)
	case fdkaac.AACEncUnsupportedChannelConf, fdkaac.AACEncInvalidElementInfoType:
		return fmt.Errorf("%w: unsupported channel layout", ErrUnsupportedFormat)
	case fdkaac.AACEncUnsupportedSamplingRate:
		return fmt.Errorf("%w: unsupported sample rate", ErrInvalidConfig)
	case fdkaac.AACEncNoMemory:
		return fmt.Errorf("%w: fixed encoder storage exhausted", ErrEncode)
	case fdkaac.AACEncQuantError:
		return fmt.Errorf("%w: quantization failed", ErrEncode)
	case fdkaac.AACEncWrittenBitsError, fdkaac.AACEncWriteScalError, fdkaac.AACEncWriteSecError, fdkaac.AACEncWriteSpecError:
		return fmt.Errorf("%w: bitstream write failed", ErrEncode)
	case fdkaac.AACEncUnsupportedAOT, fdkaac.AACEncUnsupportedFilterbank, fdkaac.AACEncUnsupportedERFormat:
		return fmt.Errorf("%w: unsupported encoder profile", ErrUnsupportedProfile)
	case fdkaac.AACEncPNSTableError:
		return fmt.Errorf("%w: PNS table lookup failed", ErrEncode)
	default:
		return fmt.Errorf("%w: internal code %#x", ErrEncode, code)
	}
}
