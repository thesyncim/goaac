#include "ffmpeg_bridge.h"

#include <limits.h>
#include <stdio.h>
#include <string.h>

#include <libavcodec/avcodec.h>
#include <libavutil/channel_layout.h>
#include <libavutil/error.h>
#include <libavutil/mathematics.h>
#include <libavutil/mem.h>
#include <libavutil/samplefmt.h>
#include <libavutil/version.h>
#include <libswresample/swresample.h>

struct AACLCDecoder {
    const AVCodec *codec;
    AVCodecContext *ctx;
    AVPacket *pkt;
    AVFrame *frame;
    SwrContext *swr;
    AVChannelLayout swr_layout;
    enum AVSampleFormat swr_format;
    int swr_rate;
    int swr_valid;
};

static int set_error(int ret, const char *op, char *errbuf, int errbuf_size)
{
    if (errbuf && errbuf_size > 0) {
        char detail[AV_ERROR_MAX_STRING_SIZE] = {0};
        av_strerror(ret, detail, sizeof(detail));
        snprintf(errbuf, (size_t)errbuf_size, "%s: %s", op, detail);
    }
    return ret;
}

static int set_message(int ret, const char *msg, char *errbuf, int errbuf_size)
{
    if (errbuf && errbuf_size > 0)
        snprintf(errbuf, (size_t)errbuf_size, "%s", msg);
    return ret;
}

int aaclc_decoder_open(AACLCDecoder **out,
                       const uint8_t *extradata,
                       int extradata_size,
                       char *errbuf,
                       int errbuf_size)
{
    AACLCDecoder *d;
    int ret;

    if (!out)
        return set_message(AVERROR(EINVAL), "missing decoder output", errbuf, errbuf_size);
    *out = NULL;

    d = av_mallocz(sizeof(*d));
    if (!d)
        return set_message(AVERROR(ENOMEM), "alloc decoder", errbuf, errbuf_size);

    d->codec = avcodec_find_decoder(AV_CODEC_ID_AAC);
    if (!d->codec) {
        aaclc_decoder_close(d);
        return set_message(AVERROR_DECODER_NOT_FOUND, "AAC decoder not found", errbuf, errbuf_size);
    }

    d->ctx = avcodec_alloc_context3(d->codec);
    d->pkt = av_packet_alloc();
    d->frame = av_frame_alloc();
    if (!d->ctx || !d->pkt || !d->frame) {
        aaclc_decoder_close(d);
        return set_message(AVERROR(ENOMEM), "alloc codec context", errbuf, errbuf_size);
    }

    if (extradata_size > 0) {
        if (!extradata)
            return set_message(AVERROR(EINVAL), "missing extradata", errbuf, errbuf_size);
        d->ctx->extradata = av_mallocz((size_t)extradata_size + AV_INPUT_BUFFER_PADDING_SIZE);
        if (!d->ctx->extradata) {
            aaclc_decoder_close(d);
            return set_message(AVERROR(ENOMEM), "alloc extradata", errbuf, errbuf_size);
        }
        memcpy(d->ctx->extradata, extradata, (size_t)extradata_size);
        d->ctx->extradata_size = extradata_size;
    }

    ret = avcodec_open2(d->ctx, d->codec, NULL);
    if (ret < 0) {
        aaclc_decoder_close(d);
        return set_error(ret, "open AAC decoder", errbuf, errbuf_size);
    }

    *out = d;
    return 0;
}

static int ensure_swr(AACLCDecoder *d, const AVFrame *frame, char *errbuf, int errbuf_size)
{
    AVChannelLayout out_layout = {0};
    int ret;

    if (d->swr_valid &&
        d->swr &&
        d->swr_format == (enum AVSampleFormat)frame->format &&
        d->swr_rate == frame->sample_rate &&
        av_channel_layout_compare(&d->swr_layout, &frame->ch_layout) == 0) {
        return 0;
    }

    swr_free(&d->swr);
    if (d->swr_valid)
        av_channel_layout_uninit(&d->swr_layout);
    d->swr_valid = 0;

    if (frame->ch_layout.nb_channels <= 0)
        return set_message(AVERROR_INVALIDDATA, "decoded frame has no channel layout", errbuf, errbuf_size);

    ret = av_channel_layout_copy(&d->swr_layout, &frame->ch_layout);
    if (ret < 0)
        return set_error(ret, "copy channel layout", errbuf, errbuf_size);
    ret = av_channel_layout_copy(&out_layout, &frame->ch_layout);
    if (ret < 0) {
        av_channel_layout_uninit(&d->swr_layout);
        return set_error(ret, "copy output channel layout", errbuf, errbuf_size);
    }

    ret = swr_alloc_set_opts2(&d->swr,
                              &out_layout,
                              AV_SAMPLE_FMT_S16,
                              frame->sample_rate,
                              &frame->ch_layout,
                              (enum AVSampleFormat)frame->format,
                              frame->sample_rate,
                              0,
                              NULL);
    av_channel_layout_uninit(&out_layout);
    if (ret < 0) {
        av_channel_layout_uninit(&d->swr_layout);
        return set_error(ret, "alloc resampler", errbuf, errbuf_size);
    }

    ret = swr_init(d->swr);
    if (ret < 0) {
        swr_free(&d->swr);
        av_channel_layout_uninit(&d->swr_layout);
        return set_error(ret, "init resampler", errbuf, errbuf_size);
    }

    d->swr_format = (enum AVSampleFormat)frame->format;
    d->swr_rate = frame->sample_rate;
    d->swr_valid = 1;
    return 0;
}

static int convert_frame(AACLCDecoder *d, AACLCPCM *out, char *errbuf, int errbuf_size)
{
    uint8_t *planes[1] = {NULL};
    int channels = d->frame->ch_layout.nb_channels;
    int max_samples;
    int max_bytes;
    int converted;
    int actual_bytes;
    int ret;

    if (d->frame->nb_samples <= 0 || channels <= 0)
        return 0;

    ret = ensure_swr(d, d->frame, errbuf, errbuf_size);
    if (ret < 0)
        return ret;

    max_samples = (int)av_rescale_rnd(swr_get_delay(d->swr, d->frame->sample_rate) + d->frame->nb_samples,
                                      d->frame->sample_rate,
                                      d->frame->sample_rate,
                                      AV_ROUND_UP);
    if (max_samples <= 0)
        return set_message(AVERROR_INVALIDDATA, "invalid output sample count", errbuf, errbuf_size);

    max_bytes = av_samples_get_buffer_size(NULL, channels, max_samples, AV_SAMPLE_FMT_S16, 1);
    if (max_bytes <= 0)
        return set_error(max_bytes, "get output buffer size", errbuf, errbuf_size);

    planes[0] = av_malloc((size_t)max_bytes);
    if (!planes[0])
        return set_message(AVERROR(ENOMEM), "alloc PCM", errbuf, errbuf_size);

    converted = swr_convert(d->swr,
                            planes,
                            max_samples,
                            (const uint8_t **)d->frame->extended_data,
                            d->frame->nb_samples);
    if (converted < 0) {
        av_freep(&planes[0]);
        return set_error(converted, "convert PCM", errbuf, errbuf_size);
    }

    actual_bytes = av_samples_get_buffer_size(NULL, channels, converted, AV_SAMPLE_FMT_S16, 1);
    if (actual_bytes < 0) {
        av_freep(&planes[0]);
        return set_error(actual_bytes, "get converted size", errbuf, errbuf_size);
    }

    out->data = planes[0];
    out->bytes = actual_bytes;
    out->samples = converted;
    out->channels = channels;
    out->sample_rate = d->frame->sample_rate;
    return 0;
}

int aaclc_decode(AACLCDecoder *d,
                 const uint8_t *data,
                 int size,
                 AACLCPCM *out,
                 char *errbuf,
                 int errbuf_size)
{
    int ret;

    if (!d || !d->ctx || !d->pkt || !d->frame)
        return set_message(AVERROR(EINVAL), "decoder is closed", errbuf, errbuf_size);
    if (!out)
        return set_message(AVERROR(EINVAL), "missing output", errbuf, errbuf_size);
    memset(out, 0, sizeof(*out));
    if (!data || size <= 0)
        return 0;

    av_packet_unref(d->pkt);
    ret = av_new_packet(d->pkt, size);
    if (ret < 0)
        return set_error(ret, "alloc packet", errbuf, errbuf_size);
    memcpy(d->pkt->data, data, (size_t)size);

    ret = avcodec_send_packet(d->ctx, d->pkt);
    av_packet_unref(d->pkt);
    if (ret < 0)
        return set_error(ret, "send packet", errbuf, errbuf_size);

    ret = avcodec_receive_frame(d->ctx, d->frame);
    if (ret == AVERROR(EAGAIN) || ret == AVERROR_EOF)
        return 0;
    if (ret < 0)
        return set_error(ret, "receive frame", errbuf, errbuf_size);

    ret = convert_frame(d, out, errbuf, errbuf_size);
    av_frame_unref(d->frame);
    if (ret < 0) {
        aaclc_free_pcm(out);
        return ret;
    }
    return 0;
}

void aaclc_free_pcm(AACLCPCM *out)
{
    if (!out)
        return;
    av_freep(&out->data);
    out->bytes = 0;
    out->samples = 0;
    out->channels = 0;
    out->sample_rate = 0;
}

void aaclc_decoder_close(AACLCDecoder *d)
{
    if (!d)
        return;
    swr_free(&d->swr);
    if (d->swr_valid)
        av_channel_layout_uninit(&d->swr_layout);
    av_frame_free(&d->frame);
    av_packet_free(&d->pkt);
    avcodec_free_context(&d->ctx);
    av_free(d);
}

const char *aaclc_ffmpeg_version(void)
{
    return av_version_info();
}
