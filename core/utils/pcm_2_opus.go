package utils

import (
	"fmt"
	"io"
	"slices"

	opus "github.com/qrtc/opus-go"
)

type funcReader struct {
	f func(p []byte) (n int, err error)
}

func (r funcReader) Read(p []byte) (n int, err error) {
	return r.f(p)
}

func StreamPCM2Opus(pcmStream io.Reader, sampleRate int, channels int, bitrate int) (io.Reader, error) {
	// 检查采样率是否支持
	if !slices.Contains([]int{8000, 12000, 16000, 24000, 48000}, bitrate) {
		return nil, fmt.Errorf("采样率 %dHz 不被Opus支持，仅支持8000/12000/16000/24000/48000Hz", sampleRate)
	}

	// 创建Opus编码器
	encoder, err := opus.CreateOpusEncoder(&opus.OpusEncoderConfig{
		SampleRate:    sampleRate,
		MaxChannels:   channels,
		Application:   opus.AppVoIP,
		FrameDuration: opus.Framesize60Ms, // 使用60ms帧长
	})
	if err != nil {
		return nil, fmt.Errorf("创建Opus编码器失败: %v", err)
	}

	// 计算每帧样本数 (60ms帧)
	samplesPerFrame := (sampleRate * 60) / 1000 // 60ms帧
	// 每个样本的字节数 (16位 = 2字节)
	bytesPerSample := 2 * channels
	// 每帧字节数
	bytesPerFrame := samplesPerFrame * bytesPerSample

	readPcmFrame := func() ([]byte, error) {
		framePCM := make([]byte, bytesPerFrame)
		_, err := io.ReadFull(pcmStream, framePCM)
		if err == io.ErrUnexpectedEOF {
			paddedFrame := make([]byte, bytesPerFrame)
			copy(paddedFrame, framePCM)
			return paddedFrame, nil
		} else if err != nil {
			return nil, err
		} else {
			return framePCM, nil
		}
	}

	return funcReader{
		f: func(p []byte) (n int, err error) {
			framePCM, err := readPcmFrame()
			if err != nil {
				if err == io.EOF {
					encoder.Close()
				}
				return 0, err
			}
			return encoder.Encode(framePCM, p)
		},
	}, nil
}
