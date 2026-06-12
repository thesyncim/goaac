// Package aac decodes AAC-LC elementary streams.
//
// The decoder core is a pure-Go translation of the pinned FAAD2 AAC-LC C
// reference, with Go parsing helpers for MPEG-4 AudioSpecificConfig and ADTS
// stream framing.
package aac
