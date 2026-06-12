#ifndef GOAAC_FFMPEG_BRIDGE_H
#define GOAAC_FFMPEG_BRIDGE_H

#include <stdint.h>

typedef struct AACLCDecoder AACLCDecoder;

typedef struct AACLCPCM {
    uint8_t *data;
    int bytes;
    int samples;
    int channels;
    int sample_rate;
} AACLCPCM;

int aaclc_decoder_open(AACLCDecoder **out,
                       const uint8_t *extradata,
                       int extradata_size,
                       char *errbuf,
                       int errbuf_size);
int aaclc_decode(AACLCDecoder *d,
                 const uint8_t *data,
                 int size,
                 AACLCPCM *out,
                 char *errbuf,
                 int errbuf_size);
void aaclc_free_pcm(AACLCPCM *out);
void aaclc_decoder_close(AACLCDecoder *d);
const char *aaclc_ffmpeg_version(void);

#endif
