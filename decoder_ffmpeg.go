//go:build cgo

package aac

/*
#cgo pkg-config: libavcodec libavutil libswresample
#include <stdlib.h>
#include "ffmpeg_bridge.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

type nativeDecoder struct {
	ptr *C.AACLCDecoder
}

func newNativeDecoder(cfg Config) (*nativeDecoder, error) {
	var ptr *C.AACLCDecoder
	var errbuf [256]C.char
	var extra *C.uint8_t
	extraLen := C.int(0)
	if len(cfg.ExtraData) > 0 {
		extra = (*C.uint8_t)(unsafe.Pointer(&cfg.ExtraData[0]))
		extraLen = C.int(len(cfg.ExtraData))
	}
	ret := C.aaclc_decoder_open(&ptr, extra, extraLen, &errbuf[0], C.int(len(errbuf)))
	if ret < 0 {
		return nil, nativeErr("open decoder", ret, &errbuf[0])
	}
	n := &nativeDecoder{ptr: ptr}
	runtime.SetFinalizer(n, (*nativeDecoder).close)
	return n, nil
}

func (n *nativeDecoder) decode(packet []byte) ([]int16, error) {
	if len(packet) == 0 {
		return nil, nil
	}
	if n == nil || n.ptr == nil {
		return nil, ErrClosed
	}
	var out C.AACLCPCM
	var errbuf [256]C.char
	ret := C.aaclc_decode(n.ptr, (*C.uint8_t)(unsafe.Pointer(&packet[0])), C.int(len(packet)), &out, &errbuf[0], C.int(len(errbuf)))
	if ret < 0 {
		return nil, nativeErr("decode", ret, &errbuf[0])
	}
	if out.data == nil || out.bytes == 0 {
		return nil, nil
	}
	defer C.aaclc_free_pcm(&out)
	if int(out.bytes)%2 != 0 {
		return nil, fmt.Errorf("aac: native decoder returned odd PCM byte count %d", int(out.bytes))
	}
	nSamples := int(out.bytes) / 2
	src := unsafe.Slice((*int16)(unsafe.Pointer(out.data)), nSamples)
	pcm := make([]int16, nSamples)
	copy(pcm, src)
	return pcm, nil
}

func (n *nativeDecoder) close() {
	if n == nil || n.ptr == nil {
		return
	}
	C.aaclc_decoder_close(n.ptr)
	n.ptr = nil
	runtime.SetFinalizer(n, nil)
}

func nativeVersion() string {
	return C.GoString(C.aaclc_ffmpeg_version())
}

func nativeErr(op string, ret C.int, cmsg *C.char) error {
	msg := C.GoString(cmsg)
	if msg == "" {
		msg = fmt.Sprintf("native error %d", int(ret))
	}
	return fmt.Errorf("aac: %s: %s", op, msg)
}
