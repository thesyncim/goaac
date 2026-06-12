// Package aac decodes AAC-LC elementary streams.
//
// The production decoder path uses the native FFmpeg AAC decoder through cgo.
// Pure Go parsing is kept source-shaped against FFmpeg's MPEG-4 Audio and ADTS
// helpers so stream validation stays available outside the native boundary.
package aac
