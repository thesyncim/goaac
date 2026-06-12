// Package aac decodes AAC-LC elementary streams.
//
// The decoder core is a pure-Go translation of the pinned FAAD2 AAC-LC C
// reference, with Go parsing helpers for MPEG-4 AudioSpecificConfig and ADTS
// stream framing.
//
// The high-level helpers decode complete ADTS streams:
//
//	pcm, cfg, err := aac.DecodeADTS(data)
//
// For long streams or reusable buffers, create a Decoder and use the Into-style
// methods:
//
//	dec, err := aac.New(aac.Options{Transport: aac.TransportADTS})
//	if err != nil {
//		return err
//	}
//	defer dec.Close()
//
//	pcm = pcm[:0]
//	pcm, info, err := dec.Decode(pcm, adtsFrame)
//
// PCM samples are interleaved signed 16-bit values in native Go int16 form.
package aac
