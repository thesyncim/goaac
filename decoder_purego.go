//go:build (darwin && arm64) || (linux && arm64)

package aac

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/thesyncim/goaac/internal/faad2ccgo"
	"modernc.org/libc"
)

type pureDecoder struct {
	tls         *libc.TLS
	handle      faad2ccgo.NeAACDecHandle
	cfg         Config
	raw         bool
	initialized bool
}

func newPureDecoder(cfg Config) (*pureDecoder, error) {
	tls := libc.NewTLS()
	handle := faad2ccgo.NeAACDecOpen(tls)
	if handle == 0 {
		tls.Close()
		return nil, fmt.Errorf("%w: FAAD2 decoder open failed", ErrNativeUnavailable)
	}

	d := &pureDecoder{
		tls:    tls,
		handle: handle,
		cfg:    cfg,
		raw:    len(cfg.ExtraData) > 0,
	}
	if err := d.setConfig(); err != nil {
		d.close()
		return nil, err
	}
	if d.raw {
		if err := d.initRaw(cfg.ExtraData); err != nil {
			d.close()
			return nil, err
		}
	}
	runtime.SetFinalizer(d, (*pureDecoder).close)
	return d, nil
}

func (d *pureDecoder) setConfig() error {
	cfgPtr := faad2ccgo.NeAACDecGetCurrentConfiguration(d.tls, d.handle)
	if cfgPtr == 0 {
		return fmt.Errorf("%w: FAAD2 decoder configuration unavailable", ErrNativeUnavailable)
	}
	cfg := (*faad2ccgo.NeAACDecConfiguration)(unsafe.Pointer(cfgPtr))
	cfg.FdefObjectType = uint8(AOTAACLC)
	if d.cfg.SampleRate > 0 {
		cfg.FdefSampleRate = uint64(d.cfg.SampleRate)
	}
	cfg.FoutputFormat = uint8(faad2ccgo.FAAD_FMT_16BIT)
	cfg.FdownMatrix = 0
	cfg.FuseOldADTSFormat = 0
	cfg.FdontUpSampleImplicitSBR = 1
	if faad2ccgo.NeAACDecSetConfiguration(d.tls, d.handle, cfgPtr) == 0 {
		return fmt.Errorf("%w: FAAD2 rejected decoder configuration", ErrInvalidConfig)
	}
	return nil
}

func (d *pureDecoder) initRaw(asc []byte) error {
	if len(asc) == 0 {
		return fmt.Errorf("%w: missing AudioSpecificConfig", ErrInvalidConfig)
	}
	ascPtr := d.allocBytes(asc)
	defer d.tls.Free(len(asc))
	metaPtr := d.tls.Alloc(16)
	defer d.tls.Free(16)
	*(*uint64)(unsafe.Pointer(metaPtr)) = 0
	*(*uint8)(unsafe.Pointer(metaPtr + 8)) = 0
	rc := faad2ccgo.NeAACDecInit2(
		d.tls,
		d.handle,
		ascPtr,
		uint64(len(asc)),
		metaPtr,
		metaPtr+8,
	)
	if rc != 0 {
		return d.frameError(uint8(rc), "raw decoder init")
	}
	d.initialized = true
	samplerate := *(*uint64)(unsafe.Pointer(metaPtr))
	channels := *(*uint8)(unsafe.Pointer(metaPtr + 8))
	if samplerate != 0 {
		d.cfg.SampleRate = int(samplerate)
	}
	if channels != 0 {
		d.cfg.Channels = int(channels)
	}
	return nil
}

func (d *pureDecoder) initADTS(frame []byte) error {
	if len(frame) == 0 {
		return ErrNeedMoreData
	}
	framePtr := d.allocBytes(frame)
	defer d.tls.Free(len(frame))
	metaPtr := d.tls.Alloc(16)
	defer d.tls.Free(16)
	*(*uint64)(unsafe.Pointer(metaPtr)) = 0
	*(*uint8)(unsafe.Pointer(metaPtr + 8)) = 0
	consumed := faad2ccgo.NeAACDecInit(
		d.tls,
		d.handle,
		framePtr,
		uint64(len(frame)),
		metaPtr,
		metaPtr+8,
	)
	if consumed < 0 {
		return fmt.Errorf("%w: FAAD2 ADTS decoder init failed", ErrInvalidADTS)
	}
	d.initialized = true
	samplerate := *(*uint64)(unsafe.Pointer(metaPtr))
	channels := *(*uint8)(unsafe.Pointer(metaPtr + 8))
	if samplerate != 0 {
		d.cfg.SampleRate = int(samplerate)
	}
	if channels != 0 {
		d.cfg.Channels = int(channels)
	}
	return nil
}

func (d *pureDecoder) decode(frame []byte) ([]int16, error) {
	if len(frame) == 0 {
		return nil, nil
	}
	if d == nil || d.handle == 0 || d.tls == nil {
		return nil, ErrClosed
	}
	if !d.initialized {
		if d.raw {
			return nil, fmt.Errorf("%w: raw decoder not initialized", ErrInvalidConfig)
		}
		if err := d.initADTS(frame); err != nil {
			return nil, err
		}
	}

	framePtr := d.allocBytes(frame)
	defer d.tls.Free(len(frame))
	infoSize := int(unsafe.Sizeof(faad2ccgo.NeAACDecFrameInfo{}))
	infoPtr := d.tls.Alloc(infoSize)
	defer d.tls.Free(infoSize)
	samples := faad2ccgo.NeAACDecDecode(
		d.tls,
		d.handle,
		infoPtr,
		framePtr,
		uint64(len(frame)),
	)
	info := *(*faad2ccgo.NeAACDecFrameInfo)(unsafe.Pointer(infoPtr))
	if info.Ferror1 != 0 {
		return nil, d.frameError(info.Ferror1, "decode")
	}
	if info.Fsamples == 0 || samples == 0 {
		return nil, nil
	}
	if info.Fobject_type != uint8(AOTAACLC) {
		return nil, fmt.Errorf("%w: FAAD2 decoded object type %d", ErrUnsupportedProfile, info.Fobject_type)
	}
	if info.Fsamplerate != 0 {
		d.cfg.SampleRate = int(info.Fsamplerate)
	}
	if info.Fchannels != 0 {
		d.cfg.Channels = int(info.Fchannels)
	}

	n := int(info.Fsamples)
	src := unsafe.Slice((*int16)(unsafe.Pointer(samples)), n)
	out := make([]int16, n)
	copy(out, src)
	return out, nil
}

func (d *pureDecoder) allocBytes(data []byte) uintptr {
	ptr := d.tls.Alloc(len(data))
	copy(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), len(data)), data)
	return ptr
}

func (d *pureDecoder) frameError(code uint8, op string) error {
	msgPtr := faad2ccgo.NeAACDecGetErrorMessage(d.tls, code)
	if msgPtr == 0 {
		return fmt.Errorf("aac: FAAD2 %s error %d", op, code)
	}
	return fmt.Errorf("aac: FAAD2 %s error %d: %s", op, code, libc.GoString(msgPtr))
}

func (d *pureDecoder) close() {
	if d == nil {
		return
	}
	runtime.SetFinalizer(d, nil)
	if d.handle != 0 && d.tls != nil {
		faad2ccgo.NeAACDecClose(d.tls, d.handle)
		d.handle = 0
	}
	if d.tls != nil {
		d.tls.Close()
		d.tls = nil
	}
}

func PureGoVersion() string {
	return "FAAD2 " + faad2ccgo.FAAD2_VERSION + " pure Go"
}
