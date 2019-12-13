#!/usr/bin/env python
# -*- coding: utf-8 -*-
import json,logging,sys

# set the default encoding to utf-8
# reload sys model to enable the getdefaultencoding method.
reload(sys);
# using exec to set the encoding, to avoid error in IDE.
exec("sys.setdefaultencoding('utf-8')");

NotSure = "不确定，请参考[issues](https://github.com/ossrs/srs/issues?q=%s)或换个提问方式"
UnknownKnowledge = "不知道，请换个提问方式"

# 实体：依赖环境
def depends_env(env):
    if env in ["CentOS", "x86-64"]:
        return "很好，官方支持，建议直接用docker运行，参考[这里](https://github.com/ossrs/srs-docker/tree/srs3)"
    if env in ["Linux", "Unix"]:
        return "可以，建议用docker[编译调试](https://github.com/ossrs/srs-docker/tree/dev)和[运行](https://github.com/ossrs/srs-docker/tree/srs3)"
    if env in ["ARM"]:
        return "可以，需要替换ST(state-threads)，参考[这里](https://github.com/ossrs/state-threads/tree/srs#usage)"
    if env in ["Windows"]:
        return "不支持，不过可以用docker运行，参考[这里](https://github.com/ossrs/srs/wiki/v1_CN_WindowsSRS)"
    if env in ["thingOS"]:
        return "不支持"
    return NotSure%(env)


# 实体：AVCodec
def av_codecs(codec):
    if codec in ["H.264", "AAC"]:
        return "很好，各种浏览器和平台都支持"
    if codec in ["H.263", "SPEEX", "PCM"]:
        return "不支持，[太老的编码格式](https://github.com/ossrs/srs/blob/b4870a6d6f94ad26c7cc9c7fb39a4246180b5864/trunk/src/kernel/srs_kernel_codec.hpp#L35)，建议用[FFMPEG](https://github.com/ossrs/srs/wiki/v1_CN_SampleFFMPEG)转码为h.264"
    if codec in ["H.265"]:
        return "不支持，浏览器支持得不好，参考[#465](https://github.com/ossrs/srs/issues/465#issuecomment-562794207)"
    if codec in ["AV1"]:
        return "不支持，是未来的趋势，参考[#1070](https://github.com/ossrs/srs/issues/1070#issuecomment-562794926)"
    if codec in ["MP3"]:
        return "部分支持，不推荐，参考[#301](https://github.com/ossrs/srs/issues/301)和[#296](https://github.com/ossrs/srs/issues/296)"
    if codec in ["Opus"] :
        return "不支持，是WebRTC的音频编码，参考[#307](https://github.com/ossrs/srs/issues/307)"
    if codec in ["SRT"]:
        return "不支持，是广电常用的协议，参考[#1147](https://github.com/ossrs/srs/issues/1147)"
    return NotSure%(codec)

# 实体：传输协议
def protocols(proto):
    if proto in ['DASH']:
        return "支持，可能会有些问题，参考[#299](https://github.com/ossrs/srs/issues/299)"
    if proto in ['HDS']:
        return "支持，用得比较少，参考[wiki](https://github.com/ossrs/srs/wiki/v2_CN_DeliveryHDS)"
    if proto in ['CMAF', 'HSS']:
        return "不支持，用户还很少"
    if proto in ['MPEGTS', 'RTSP']:
        return "支持推流，不支持播放，参考[wiki](https://github.com/ossrs/srs/wiki/v2_CN_Streamer)"
    if proto in ['HTTP']:
        return "支持但不常用，HTTP-FLV最通用，HTTP-TS和HTTP-MP3偶然会有应用，参考[wiki](https://github.com/ossrs/srs/wiki/v2_CN_DeliveryHttpStream)"
    if proto in ['HLS']:
        return "支持且常用，参考[wiki](https://github.com/ossrs/srs/wiki/v3_CN_DeliveryHLS)"
    if proto in ['RTMP']:
        return "支持且常用，参考[wiki](https://github.com/ossrs/srs/wiki/v1_CN_DeliveryRTMP)"
    if proto in ['FLV']:
        return "支持且常用，参考[wiki](https://github.com/ossrs/srs/wiki/v2_CN_DeliveryHttpStream)"
    if proto in ['AXP']:
        return "不支持。AXP是阿里CDN的技术框架，可以让NGINX,Libevent等事件驱动框架快速集成QUIC,KCP,SRT等协议，取代TCP作为传输层，可能会(目前还没)开源"
    if proto in ['RTP']:
        return "不支持，RTP/SRTP是WebRTC的传输协议，参考[#307](https://github.com/ossrs/srs/issues/307)"
    if proto in ['PS']:
        return "不支持，国标/PS是监控摄像头的传输协议，参考[互联网监控](https://mp.weixin.qq.com/s/6NNIm5CxYg3t_tHg2TeyCQ)和[SRS和监控](https://mp.weixin.qq.com/s/FDwGLdtb5Rjo0aus02GHkQ)"
    return NotSure%(proto)

# 实体：录制点播
def dvr(category):
    if category in ['VOD']:
        return "不支持，直接用HTTP分发就可以"
    if category in ['TimeShift']:
        return "不支持，需要业务系统做内容管理"
    if category in ['DVR']:
        return "支持，可录制为[FLV或MP4](https://github.com/ossrs/srs/wiki/v3_CN_DVR)"
    return NotSure%(category)

# python -c 'import fc;print fc.handler(json.dumps({"key":"depends_env","arg0":"centos"}), "")'
def handler(event, context):
    logger = logging.getLogger()

    def fn_arg0(pfn, arg0):
        (rr, rrs) = ([], {})
        for arg0 in arg0.split(','):
            if arg0 in rrs:
                continue
            rrs[arg0] = True
            rr.append('**%s:** %s'%(arg0, pfn(arg0)))
        return rr

    o = json.loads(event)
    key = o['key']
    if key == 'depends_env':
        rr = fn_arg0(depends_env, o['arg0'])
    elif key == 'av_codecs':
        rr = fn_arg0(av_codecs, o['arg0'])
    elif key == 'protocol':
        rr = fn_arg0(protocols, o['arg0'])
    elif key == 'dvr':
        rr = fn_arg0(dvr, o['arg0'])
    else:
        rr = [UnknownKnowledge]

    logger.info('fc %s result is %s'%(o, rr))
    if len(rr) == 1:
        return rr[0]
    return '\n- ' + '\n- '.join(rr)

if __name__== "__main__":
    argv = []
    for arg in sys.argv:
        if arg == 'python' or arg.endswith('.py'):
            continue
        argv.append(arg)

    if len(argv) == 0:
        print "No parameters"
        sys.exit(-1)

    if argv[0].startswith('{'):
        print handler(argv[0], None)
        sys.exit(0)

    obj = {}
    for index in range(len(argv)):
        if index == 0:
            obj['key'] = argv[index]
        else:
            obj['arg%d'%(index-1)] = argv[index]
    print handler(json.dumps(obj), None)
    sys.exit(0)