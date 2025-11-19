package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/hajimehoshi/go-mp3"
	opus "github.com/qrtc/opus-go"
)

// OpusDecoder 封装opus解码器
type OpusDecoder struct {
	decoder   *opus.OpusDecoder
	mu        sync.Mutex
	config    *OpusDecoderConfig
	outBuffer []byte
}

// OpusDecoderConfig 解码器配置
type OpusDecoderConfig struct {
	SampleRate  int
	MaxChannels int
}

// NewOpusDecoder 创建新的opus解码器
func NewOpusDecoder(config *OpusDecoderConfig) (*OpusDecoder, error) {
	if config == nil {
		config = &OpusDecoderConfig{
			SampleRate:  24000, // 默认使用24kHz采样率
			MaxChannels: 1,     // 默认单通道
		}
	}

	libConfig := &opus.OpusDecoderConfig{
		SampleRate:  config.SampleRate,
		MaxChannels: config.MaxChannels,
	}

	decoder, err := opus.CreateOpusDecoder(libConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Opus解码器失败: %v", err)
	}

	bufSize := max(config.SampleRate*2*config.MaxChannels*120/1000, 8192)

	return &OpusDecoder{
		decoder:   decoder,
		config:    config,
		outBuffer: make([]byte, bufSize),
	}, nil
}

// Decode 解码opus数据为PCM
func (d *OpusDecoder) Decode(opusData []byte) ([]byte, error) {
	if len(opusData) == 0 {
		return nil, nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// 使用预分配的缓冲区
	n, err := d.decoder.Decode(opusData, d.outBuffer)
	if err != nil {
		return nil, fmt.Errorf("Opus解码失败: %v", err)
	}

	// 返回解码后的PCM数据的副本
	result := make([]byte, n)
	copy(result, d.outBuffer[:n])
	return result, nil
}

// Close 关闭解码器
func (d *OpusDecoder) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.decoder != nil {
		if err := d.decoder.Close(); err != nil {
			return fmt.Errorf("关闭Opus解码器失败: %v", err)
		}
		d.decoder = nil
	}
	return nil
}

func SaveAudioToWavFile(
	data []byte,
	fileName string,
	sampleRate int,
	channels int,
	bitsPerSample int,
	append bool, // 新增参数：是否追加写入，默认为false
) (string, error) {
	// 处理文件名
	if fileName == "" {
		fileName = "output.wav"
	}

	var file *os.File
	var err error
	var currentDataSize int64 = 0

	// 检查文件是否存在
	_, err = os.Stat(fileName)
	fileExists := !os.IsNotExist(err)

	if append && fileExists {
		// 追加模式：打开现有文件
		file, err = os.OpenFile(fileName, os.O_RDWR, 0644)
		if err != nil {
			return "", fmt.Errorf("打开文件失败: %v", err)
		}
		defer file.Close()

		// 获取当前数据大小
		fileInfo, err := file.Stat()
		if err != nil {
			return "", fmt.Errorf("获取文件信息失败: %v", err)
		}
		currentDataSize = max(fileInfo.Size()-44, 0)

		// 定位到文件末尾准备追加数据
		_, err = file.Seek(0, io.SeekEnd)
		if err != nil {
			return "", fmt.Errorf("定位文件末尾失败: %v", err)
		}
	} else {
		// 覆写模式：删除现有文件（如果存在）并创建新文件
		if fileExists {
			if err := os.Remove(fileName); err != nil {
				return "", fmt.Errorf("删除现有文件失败: %v", err)
			}
		}

		// 创建新文件
		file, err = os.Create(fileName)
		if err != nil {
			return "", fmt.Errorf("创建文件失败: %v", err)
		}
		defer file.Close()

		// 写入WAV文件头
		if err := writeWavHeader(file, 0, sampleRate, channels, bitsPerSample); err != nil {
			return "", fmt.Errorf("写入WAV头失败: %v", err)
		}
	}

	// 打开现有文件进行追加
	file, err = os.OpenFile(fileName, os.O_WRONLY, 0o644)
	// 写入音频数据
	_, err = file.Write(data)
	if err != nil {
		return "", fmt.Errorf("写入数据失败: %v", err)
	}

	// 更新WAV头中的数据大小
	newDataSize := currentDataSize + int64(len(data))
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("定位文件开头失败: %v", err)
	}

	if err := writeWavHeader(file, int(newDataSize), sampleRate, channels, bitsPerSample); err != nil {
		return "", fmt.Errorf("更新WAV头失败: %v", err)
	}

	return fileName, nil
}

// 写入WAV文件头
func writeWavHeader(file *os.File, dataSize int, sampleRate, channels, bitsPerSample int) error {
	// RIFF块
	header := make([]byte, 44)
	copy(header[0:4], []byte("RIFF"))

	// 文件总长度 = 数据大小 + 头部大小(36) - 8
	fileSize := uint32(dataSize + 36)
	header[4] = byte(fileSize)
	header[5] = byte(fileSize >> 8)
	header[6] = byte(fileSize >> 16)
	header[7] = byte(fileSize >> 24)

	// 文件类型
	copy(header[8:12], []byte("WAVE"))

	// 格式块
	copy(header[12:16], []byte("fmt "))

	// 格式块大小(16字节)
	header[16] = 16
	header[17] = 0
	header[18] = 0
	header[19] = 0

	// 音频格式(1表示PCM)
	header[20] = 1
	header[21] = 0

	// 通道数
	header[22] = byte(channels)
	header[23] = 0

	// 采样率
	header[24] = byte(sampleRate)
	header[25] = byte(sampleRate >> 8)
	header[26] = byte(sampleRate >> 16)
	header[27] = byte(sampleRate >> 24)

	// 字节率 = 采样率 × 通道数 × 位深度/8
	byteRate := uint32(sampleRate * channels * bitsPerSample / 8)
	header[28] = byte(byteRate)
	header[29] = byte(byteRate >> 8)
	header[30] = byte(byteRate >> 16)
	header[31] = byte(byteRate >> 24)

	// 块对齐 = 通道数 × 位深度/8
	blockAlign := uint16(channels * bitsPerSample / 8)
	header[32] = byte(blockAlign)
	header[33] = byte(blockAlign >> 8)

	// 位深度
	header[34] = byte(bitsPerSample)
	header[35] = byte(bitsPerSample >> 8)

	// 数据块
	copy(header[36:40], []byte("data"))

	// 数据大小
	header[40] = byte(dataSize)
	header[41] = byte(dataSize >> 8)
	header[42] = byte(dataSize >> 16)
	header[43] = byte(dataSize >> 24)

	_, err := file.Write(header)
	return err
}

// 保留原来的函数，但使用新函数
func SaveAudioToFile(data []byte, fileName string) (string, error) {
	// 默认使用16kHz, 单声道, 16位
	return SaveAudioToWavFile(data, fileName, 24000, 1, 16, false)
}

func ReadPCMDataFromWavFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开WAV文件失败: %v", err)
	}
	defer file.Close()

	// 跳过WAV头
	header := make([]byte, 44)
	if _, err := file.Read(header); err != nil {
		return nil, fmt.Errorf("读取WAV头失败: %v", err)
	}

	// 读取PCM数据
	pcmData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取PCM数据失败: %v", err)
	}

	return pcmData, nil
}

func AudioToPCMData(audioFile string) ([][]byte, float64, error) {
	file, err := os.Open(audioFile)
	if err != nil {
		return nil, 0, fmt.Errorf("打开音频文件失败: %v", err)
	}
	defer file.Close()

	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, 0, fmt.Errorf("创建MP3解码器失败: %v", err)
	}

	mp3SampleRate := decoder.SampleRate()
	// fmt.Println("AudioToPCMData 原始MP3采样率:", mp3SampleRate)
	// 目标采样率设为24kHz
	targetSampleRate := 24000

	// decoder.Length() 返回解码后的PCM数据总字节数 (16-bit little-endian stereo)
	pcmBytes := make([]byte, decoder.Length())
	// ReadFull确保读取所有请求的字节，否则返回错误
	if _, err := io.ReadFull(decoder, pcmBytes); err != nil {
		// 如果 decoder.Length() 为 0, pcmBytes 为空, ReadFull 读取 0 字节, 返回 nil 错误，这是正常的。
		// 如果 decoder.Length() > 0 且 ReadFull 返回错误, 表示未能读取完整的PCM数据。
		return nil, 0, fmt.Errorf("读取PCM数据失败: %v", err)
	}

	// go-mp3 解码为 16-bit little-endian stereo PCM.
	// pcmBytes 包含交错的立体声数据 (LRLRLR...).
	// 每个立体声样本对 (左16位, 右16位) 占用4字节.
	// numMonoSamples 是转换后得到的16位单声道样本的数量.
	numMonoSamples := len(pcmBytes) / 4

	if numMonoSamples == 0 {
		// 处理 pcmBytes 为空或数据不足以形成一个单声道样本的情况 (即少于4字节).
		return [][]byte{}, 0, nil // 返回空数据
	}

	pcmMonoInt16 := make([]int16, numMonoSamples)
	for i := range numMonoSamples {
		// 从pcmBytes中提取16位小端序的左右声道样本
		// pcmBytes[i*4+0] = 左声道低字节, pcmBytes[i*4+1] = 左声道高字节
		// pcmBytes[i*4+2] = 右声道低字节, pcmBytes[i*4+3] = 右声道高字节
		leftSample := int16(uint16(pcmBytes[i*4+0]) | (uint16(pcmBytes[i*4+1]) << 8))
		rightSample := int16(uint16(pcmBytes[i*4+2]) | (uint16(pcmBytes[i*4+3]) << 8))

		// 通过平均值混合为单声道样本
		// 使用int32进行中间求和以防止在除法前溢出
		pcmMonoInt16[i] = int16((int32(leftSample) + int32(rightSample)) / 2)
	}

	// 重采样到目标采样率（如果需要）
	var resampledPcmInt16 []int16
	var finalSampleRate int

	if mp3SampleRate != targetSampleRate {
		fmt.Printf("重采样从 %dHz 到 %dHz\n", mp3SampleRate, targetSampleRate)
		resampledPcmInt16 = resamplePCM(pcmMonoInt16, mp3SampleRate, targetSampleRate)
		finalSampleRate = targetSampleRate
	} else {
		resampledPcmInt16 = pcmMonoInt16
		finalSampleRate = mp3SampleRate
	}

	// 将 []int16 类型的单声道PCM数据转换为 []byte (仍然是16位小端序)
	monoPcmDataBytes := make([]byte, len(resampledPcmInt16)*2) // 每个int16样本占用2字节
	for i, sample := range resampledPcmInt16 {
		monoPcmDataBytes[i*2] = byte(sample)        // 低字节 (LSB)
		monoPcmDataBytes[i*2+1] = byte(sample >> 8) // 高字节 (MSB)
	}

	// 音频播放时长（基于重采样后的数据）
	duration := float64(len(resampledPcmInt16)) / float64(finalSampleRate) // 单声道PCM数据的时长 (秒)

	// 函数签名要求返回 [][]byte.
	// 将整个单声道PCM数据作为外部切片中的单个段/切片返回.
	return [][]byte{monoPcmDataBytes}, duration, nil
}

// AudioToOpusData 将音频文件转换为Opus数据块
func AudioToOpusData(audioFile string) ([][]byte, float64, error) {
	var pcmData [][]byte
	var err error
	var duration float64

	// 获取采样率 (固定使用24000Hz作为Opus编码的采样率)
	// 如果采样率不是24000Hz，PCMSlicesToOpusData会处理重采样
	opusSampleRate := 24000
	channels := 1

	if strings.HasSuffix(audioFile, ".mp3") {
		// 先将MP3转为PCM
		pcmData, duration, err = AudioToPCMData(audioFile)
		if err != nil {
			return nil, 0, fmt.Errorf("PCM转换失败: %v", err)
		}

		if len(pcmData) == 0 {
			return nil, 0, fmt.Errorf("PCM转换结果为空")
		}

	} else {
		var singlePcmData []byte
		singlePcmData, _ = ReadPCMDataFromWavFile(audioFile)
		pcmData = [][]byte{singlePcmData}
	}

	// 将PCM转换为Opus
	opusData, err := PCMSlicesToOpusData(pcmData, opusSampleRate, channels, 0)
	if err != nil {
		return nil, 0, fmt.Errorf("PCM转Opus失败: %v", err)
	}

	return opusData, duration, nil
}

// PCMSlicesToOpusData 将PCM数据切片批量编码为Opus格式
func PCMSlicesToOpusData(pcmSlices [][]byte, sampleRate int, channels int, bitrate int) ([][]byte, error) {
	if len(pcmSlices) == 0 {
		return nil, fmt.Errorf("PCM数据切片为空")
	}

	// 检查采样率是否支持
	supportedRates := map[int]bool{8000: true, 12000: true, 16000: true, 24000: true, 48000: true}
	if !supportedRates[sampleRate] {
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
	defer encoder.Close()

	// 所有编码后的Opus数据包
	var allOpusPackets [][]byte

	// 计算每帧样本数 (60ms帧)
	samplesPerFrame := (sampleRate * 60) / 1000 // 60ms帧
	// 每个样本的字节数 (16位 = 2字节)
	bytesPerSample := 2 * channels
	// 每帧字节数
	bytesPerFrame := samplesPerFrame * bytesPerSample

	for _, pcmSlice := range pcmSlices {
		if len(pcmSlice) == 0 {
			continue
		}

		// 确保PCM数据长度是偶数
		if len(pcmSlice)%2 != 0 {
			pcmSlice = pcmSlice[:len(pcmSlice)-1] // 截断最后一个字节
			if len(pcmSlice) == 0 {
				continue
			}
		}

		// 计算这个PCM片段可以分成多少帧
		numFrames := len(pcmSlice) / bytesPerFrame
		if len(pcmSlice)%bytesPerFrame != 0 {
			numFrames++ // 如果有剩余数据，额外增加一帧
		}

		// 逐帧处理PCM数据
		for frameIdx := range numFrames {
			frameStart := frameIdx * bytesPerFrame
			frameEnd := min(frameStart+bytesPerFrame, len(pcmSlice))

			// 当前帧的PCM数据
			framePcm := pcmSlice[frameStart:frameEnd]

			// 如果最后一帧数据不足，需要填充静音数据到完整帧大小
			if len(framePcm) < bytesPerFrame {
				paddedFrame := make([]byte, bytesPerFrame)
				copy(paddedFrame, framePcm)
				framePcm = paddedFrame
			}

			// 分配输出缓冲区 (Opus编码后的数据通常比PCM小)
			outBuf := make([]byte, len(framePcm))

			// 编码这一帧PCM数据到Opus
			n, err := encoder.Encode(framePcm, outBuf)
			if err != nil {
				continue // 跳过这一帧，继续处理下一帧
			}

			if n == 0 {
				continue // 跳过空帧
			}

			// 将编码后的Opus数据添加到结果集
			allOpusPackets = append(allOpusPackets, outBuf[:n])
		}
	}

	if len(allOpusPackets) == 0 {
		return nil, fmt.Errorf("所有PCM切片编码后为空")
	}

	return allOpusPackets, nil
}

// resamplePCM 使用线性插值对PCM数据进行重采样
func resamplePCM(input []int16, inputSampleRate, outputSampleRate int) []int16 {
	if inputSampleRate == outputSampleRate {
		return input
	}

	inputLength := len(input)
	if inputLength == 0 {
		return []int16{}
	}

	// 计算重采样比率
	ratio := float64(inputSampleRate) / float64(outputSampleRate)
	outputLength := int(float64(inputLength) / ratio)

	if outputLength == 0 {
		return []int16{}
	}

	output := make([]int16, outputLength)

	for i := range outputLength {
		// 计算在输入数组中的位置
		srcIndex := float64(i) * ratio

		// 获取整数和小数部分
		index := int(srcIndex)
		fraction := srcIndex - float64(index)

		if index >= inputLength-1 {
			// 如果超出边界，使用最后一个样本
			output[i] = input[inputLength-1]
		} else {
			// 线性插值
			sample1 := float64(input[index])
			sample2 := float64(input[index+1])
			interpolated := sample1 + fraction*(sample2-sample1)
			output[i] = int16(interpolated)
		}
	}

	return output
}
