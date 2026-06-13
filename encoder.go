package aac

import (
	"fmt"
	"sync"

	"github.com/thesyncim/goaac/internal/fdkaac"
)

const (
	encoderSamplesPerFrame           = 1024
	encoderMaxChannels               = 2
	encoderRawBufferBytes            = 8192
	encoderAACLCBlockSwitchLookahead = 4*(encoderSamplesPerFrame/8) + (encoderSamplesPerFrame/8)/2
	encoderAACLCDelaySamples         = encoderSamplesPerFrame + encoderAACLCBlockSwitchLookahead
)

type encoderOutputMode int

const (
	encoderOutputUnset encoderOutputMode = iota
	encoderOutputConfigured
	encoderOutputFLV
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
	flushing         bool
	flushZeros       int
	pendingSamples   int
	pendingOutput    encoderOutputMode
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
	errCode := fdkaac.AACEncOK
	if transport == TransportADTS {
		errCode = fdkaac.FDKaacEncInitADTSFrameState(&e.state, encCfg)
	} else {
		errCode = fdkaac.FDKaacEncInitRawFrameState(&e.state, encCfg)
	}
	if errCode != fdkaac.AACEncOK {
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
	if e.transport != TransportRaw {
		return dst, EncodedFrameInfo{}, fmt.Errorf("%w: encoder transport is %s, not %s", ErrInvalidConfig, e.transport, TransportRaw)
	}
	if e.flushing {
		return dst, EncodedFrameInfo{}, ErrClosed
	}
	if e.pendingSamples != 0 {
		return dst, EncodedFrameInfo{}, fmt.Errorf("%w: encoder has %d buffered samples per channel", ErrInvalidFrame, e.pendingSamples)
	}
	before := len(dst)
	out, result, err := e.encodeRawPayloadLocked(dst, pcm)
	if err != nil {
		return dst, EncodedFrameInfo{}, err
	}
	info := e.encodedFrameInfo(TransportRaw, len(pcm), len(out)-before, len(out)-before, 0, result)
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
	if e.transport != TransportADTS {
		return dst, EncodedFrameInfo{}, fmt.Errorf("%w: encoder transport is %s, not %s", ErrInvalidConfig, e.transport, TransportADTS)
	}
	if e.flushing {
		return dst, EncodedFrameInfo{}, ErrClosed
	}
	if e.pendingSamples != 0 {
		return dst, EncodedFrameInfo{}, fmt.Errorf("%w: encoder has %d buffered samples per channel", ErrInvalidFrame, e.pendingSamples)
	}
	raw, result, err := e.encodeRawPayloadLocked(e.raw[:0], pcm)
	if err != nil {
		return dst, EncodedFrameInfo{}, err
	}

	before := len(dst)
	dst, err = fdkaac.AppendADTSHeaderWithScratch(dst, e.coderConfig(), len(raw), e.adtsBufferFullnessLocked(), &e.transportScratch)
	if err != nil {
		return dst, EncodedFrameInfo{}, err
	}
	headerBytes := len(dst) - before
	dst = append(dst, raw...)
	info := e.encodedFrameInfo(TransportADTS, len(pcm), len(dst)-before, len(raw), headerBytes, result)
	return dst, info, nil
}

// EncodeSamplesInto consumes interleaved S16 PCM and appends at most one AAC
// frame using the encoder's configured transport. It reports how many input
// samples were consumed and ready=false when more PCM is needed before an
// access unit can be emitted.
func (e *Encoder) EncodeSamplesInto(dst []byte, pcm []int16) ([]byte, EncodedFrameInfo, int, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return dst, EncodedFrameInfo{}, 0, false, ErrClosed
	}
	if e.flushing {
		return dst, EncodedFrameInfo{}, 0, false, ErrClosed
	}
	consumed, ready, err := e.bufferPCMLocked(pcm, encoderOutputConfigured)
	if err != nil || !ready {
		return dst, EncodedFrameInfo{}, consumed, false, err
	}
	switch e.transport {
	case TransportRaw:
		before := len(dst)
		out, result, err := e.encodeBufferedPayloadLocked(dst)
		if err != nil {
			return dst, EncodedFrameInfo{}, consumed, false, err
		}
		info := e.encodedFrameInfo(TransportRaw, consumed, len(out)-before, len(out)-before, 0, result)
		return out, info, consumed, true, nil
	case TransportADTS:
		raw, result, err := e.encodeBufferedPayloadLocked(e.raw[:0])
		if err != nil {
			return dst, EncodedFrameInfo{}, consumed, false, err
		}
		before := len(dst)
		dst, err = fdkaac.AppendADTSHeaderWithScratch(dst, e.coderConfig(), len(raw), e.adtsBufferFullnessLocked(), &e.transportScratch)
		if err != nil {
			return dst, EncodedFrameInfo{}, consumed, false, err
		}
		headerBytes := len(dst) - before
		dst = append(dst, raw...)
		info := e.encodedFrameInfo(TransportADTS, consumed, len(dst)-before, len(raw), headerBytes, result)
		return dst, info, consumed, true, nil
	default:
		return dst, EncodedFrameInfo{}, consumed, false, fmt.Errorf("%w: unknown encoder transport %s", ErrInvalidConfig, e.transport)
	}
}

// FlushFrameInto appends one delayed encoder frame using the configured
// transport. It returns more=false once the AAC-LC encoder delay has been fully
// drained. After flushing starts, normal Encode calls return ErrClosed.
func (e *Encoder) FlushFrameInto(dst []byte) ([]byte, EncodedFrameInfo, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return dst, EncodedFrameInfo{}, false, ErrClosed
	}
	switch e.transport {
	case TransportRaw:
		before := len(dst)
		out, result, more, err := e.flushRawPayloadLocked(dst, encoderOutputConfigured)
		if err != nil || !more {
			return dst, EncodedFrameInfo{}, more, err
		}
		info := e.encodedFrameInfo(TransportRaw, 0, len(out)-before, len(out)-before, 0, result)
		return out, info, true, nil
	case TransportADTS:
		raw, result, more, err := e.flushRawPayloadLocked(e.raw[:0], encoderOutputConfigured)
		if err != nil || !more {
			return dst, EncodedFrameInfo{}, more, err
		}
		before := len(dst)
		dst, err = fdkaac.AppendADTSHeaderWithScratch(dst, e.coderConfig(), len(raw), e.adtsBufferFullnessLocked(), &e.transportScratch)
		if err != nil {
			return dst, EncodedFrameInfo{}, false, err
		}
		headerBytes := len(dst) - before
		dst = append(dst, raw...)
		info := e.encodedFrameInfo(TransportADTS, 0, len(dst)-before, len(raw), headerBytes, result)
		return dst, info, true, nil
	default:
		return dst, EncodedFrameInfo{}, false, fmt.Errorf("%w: unknown encoder transport %s", ErrInvalidConfig, e.transport)
	}
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
	if e.transport != TransportRaw {
		return dst, EncodedFrameInfo{}, fmt.Errorf("%w: encoder transport is %s, not %s", ErrInvalidConfig, e.transport, TransportRaw)
	}
	if e.flushing {
		return dst, EncodedFrameInfo{}, ErrClosed
	}
	if e.pendingSamples != 0 {
		return dst, EncodedFrameInfo{}, fmt.Errorf("%w: encoder has %d buffered samples per channel", ErrInvalidFrame, e.pendingSamples)
	}
	raw, result, err := e.encodeRawPayloadLocked(e.raw[:0], pcm)
	if err != nil {
		return dst, EncodedFrameInfo{}, err
	}
	before := len(dst)
	dst = AppendFLVAACRawTag(dst, raw)
	info := e.encodedFrameInfo(TransportRaw, len(pcm), len(dst)-before, len(raw), 0, result)
	return dst, info, nil
}

// EncodeRTMPMessageInto encodes one PCM frame and appends an RTMP AAC audio
// message payload. RTMP uses the same body layout as an FLV AAC audio tag.
func (e *Encoder) EncodeRTMPMessageInto(dst []byte, pcm []int16) ([]byte, EncodedFrameInfo, error) {
	return e.EncodeFLVTagInto(dst, pcm)
}

// EncodeFLVSamplesInto consumes interleaved S16 PCM and appends at most one FLV
// AAC raw audio tag body. It reports ready=false when more PCM is needed before
// a message can be emitted.
func (e *Encoder) EncodeFLVSamplesInto(dst []byte, pcm []int16) ([]byte, EncodedFrameInfo, int, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return dst, EncodedFrameInfo{}, 0, false, ErrClosed
	}
	if e.transport != TransportRaw {
		return dst, EncodedFrameInfo{}, 0, false, fmt.Errorf("%w: encoder transport is %s, not %s", ErrInvalidConfig, e.transport, TransportRaw)
	}
	if e.flushing {
		return dst, EncodedFrameInfo{}, 0, false, ErrClosed
	}
	consumed, ready, err := e.bufferPCMLocked(pcm, encoderOutputFLV)
	if err != nil || !ready {
		return dst, EncodedFrameInfo{}, consumed, false, err
	}
	raw, result, err := e.encodeBufferedPayloadLocked(e.raw[:0])
	if err != nil {
		return dst, EncodedFrameInfo{}, consumed, false, err
	}
	before := len(dst)
	dst = AppendFLVAACRawTag(dst, raw)
	info := e.encodedFrameInfo(TransportRaw, consumed, len(dst)-before, len(raw), 0, result)
	return dst, info, consumed, true, nil
}

// EncodeRTMPSamplesInto consumes interleaved S16 PCM and appends at most one
// RTMP AAC audio message payload.
func (e *Encoder) EncodeRTMPSamplesInto(dst []byte, pcm []int16) ([]byte, EncodedFrameInfo, int, bool, error) {
	return e.EncodeFLVSamplesInto(dst, pcm)
}

// FlushFLVTagInto appends one delayed FLV AAC raw audio tag body. It returns
// more=false once the AAC-LC encoder delay has been fully drained.
func (e *Encoder) FlushFLVTagInto(dst []byte) ([]byte, EncodedFrameInfo, bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return dst, EncodedFrameInfo{}, false, ErrClosed
	}
	if e.transport != TransportRaw {
		return dst, EncodedFrameInfo{}, false, fmt.Errorf("%w: encoder transport is %s, not %s", ErrInvalidConfig, e.transport, TransportRaw)
	}
	raw, result, more, err := e.flushRawPayloadLocked(e.raw[:0], encoderOutputFLV)
	if err != nil || !more {
		return dst, EncodedFrameInfo{}, more, err
	}
	before := len(dst)
	dst = AppendFLVAACRawTag(dst, raw)
	info := e.encodedFrameInfo(TransportRaw, 0, len(dst)-before, len(raw), 0, result)
	return dst, info, true, nil
}

// FlushRTMPMessageInto appends one delayed RTMP AAC audio message payload.
func (e *Encoder) FlushRTMPMessageInto(dst []byte) ([]byte, EncodedFrameInfo, bool, error) {
	return e.FlushFLVTagInto(dst)
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
	return e.encodePreparedPayloadLocked(dst)
}

func (e *Encoder) flushRawPayloadLocked(dst []byte, output encoderOutputMode) ([]byte, fdkaac.AACEncFrameResult, bool, error) {
	if e.pendingSamples > 0 {
		if e.pendingOutput != output {
			return dst, fdkaac.AACEncFrameResult{}, false, fmt.Errorf("%w: encoder has buffered samples for another output shape", ErrInvalidConfig)
		}
		e.flushing = true
		zeros := encoderSamplesPerFrame - e.pendingSamples
		e.clearPlanarRange(e.pendingSamples, encoderSamplesPerFrame)
		out, result, err := e.encodePreparedPayloadLocked(dst)
		if err != nil {
			return dst, result, false, err
		}
		e.pendingSamples = 0
		e.pendingOutput = encoderOutputUnset
		e.flushZeros += zeros
		return out, result, true, nil
	}
	e.flushing = true
	if e.flushZeros >= encoderAACLCDelaySamples {
		return dst, fdkaac.AACEncFrameResult{}, false, nil
	}
	clear(e.planar[:e.cfg.Channels*encoderSamplesPerFrame])
	out, result, err := e.encodePreparedPayloadLocked(dst)
	if err != nil {
		return dst, result, false, err
	}
	e.flushZeros += encoderSamplesPerFrame
	return out, result, true, nil
}

func (e *Encoder) encodePreparedPayloadLocked(dst []byte) ([]byte, fdkaac.AACEncFrameResult, error) {
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

func (e *Encoder) encodeBufferedPayloadLocked(dst []byte) ([]byte, fdkaac.AACEncFrameResult, error) {
	if e.pendingSamples != encoderSamplesPerFrame {
		return dst, fdkaac.AACEncFrameResult{}, fmt.Errorf("%w: got %d buffered samples per channel, want %d", ErrInvalidFrame, e.pendingSamples, encoderSamplesPerFrame)
	}
	out, result, err := e.encodePreparedPayloadLocked(dst)
	if err != nil {
		return dst, result, err
	}
	e.pendingSamples = 0
	e.pendingOutput = encoderOutputUnset
	return out, result, nil
}

func (e *Encoder) bufferPCMLocked(pcm []int16, output encoderOutputMode) (int, bool, error) {
	channels := e.cfg.Channels
	if len(pcm)%channels != 0 {
		return 0, false, fmt.Errorf("%w: got %d interleaved samples for %d channels", ErrInvalidFrame, len(pcm), channels)
	}
	if e.pendingSamples > 0 && e.pendingOutput != output {
		return 0, false, fmt.Errorf("%w: encoder has buffered samples for another output shape", ErrInvalidConfig)
	}
	available := len(pcm) / channels
	if available == 0 {
		return 0, e.pendingSamples == encoderSamplesPerFrame, nil
	}
	need := encoderSamplesPerFrame - e.pendingSamples
	if need <= 0 {
		return 0, true, nil
	}
	if available < need {
		need = available
	}
	if e.pendingSamples == 0 {
		e.pendingOutput = output
	}
	e.copyPCMToPlanar(e.pendingSamples, pcm[:need*channels], need)
	e.pendingSamples += need
	return need * channels, e.pendingSamples == encoderSamplesPerFrame, nil
}

func (e *Encoder) preparePCM(pcm []int16) error {
	want := e.cfg.Channels * encoderSamplesPerFrame
	if len(pcm) != want {
		return fmt.Errorf("%w: got %d interleaved samples, want %d", ErrInvalidFrame, len(pcm), want)
	}
	e.copyPCMToPlanar(0, pcm, encoderSamplesPerFrame)
	return nil
}

func (e *Encoder) copyPCMToPlanar(dstSample int, pcm []int16, samplesPerChannel int) {
	if e.cfg.Channels == 1 {
		copy(e.planar[dstSample:dstSample+samplesPerChannel], pcm[:samplesPerChannel])
		return
	}
	for i := 0; i < samplesPerChannel; i++ {
		base := i * e.cfg.Channels
		e.planar[dstSample+i] = pcm[base]
		e.planar[encoderSamplesPerFrame+dstSample+i] = pcm[base+1]
	}
}

func (e *Encoder) clearPlanarRange(startSample int, endSample int) {
	for ch := 0; ch < e.cfg.Channels; ch++ {
		base := ch * encoderSamplesPerFrame
		clear(e.planar[base+startSample : base+endSample])
	}
}

func (e *Encoder) encodedFrameInfo(transport Transport, inputSamples int, outputBytes int, payloadBytes int, headerBytes int, result fdkaac.AACEncFrameResult) EncodedFrameInfo {
	return EncodedFrameInfo{
		Config:               e.cfg,
		Transport:            transport,
		InputSamples:         inputSamples,
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

func (e *Encoder) adtsBufferFullnessLocked() int {
	ncc := e.state.Init.ChannelMapping.NChannelsEff
	if ncc <= 0 {
		return 0x7ff
	}
	bits := fdkaac.FDKaacEncEncBitresToTpBitres(
		&e.state.QC.Kernel,
		e.state.QC.Kernel.BitrateMode,
		e.state.Config.AudioMuxVersion,
		ncc,
	)
	fullness := bits / ncc / 32
	if fullness > 0x7ff {
		return 0x7ff
	}
	if fullness < 0 {
		return 0
	}
	return fullness
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
