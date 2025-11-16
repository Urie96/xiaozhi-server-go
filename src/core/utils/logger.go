package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

const (
	LogRetentionDays = 7 // 日志保留天数，硬编码7天
)

var DefaultLogger *Logger

type LogCfg struct {
	LogLevel string `yaml:"log_level" json:"log_level"`
}

// CustomTextHandler 自定义文本处理器，支持彩色输出和格式化
type CustomTextHandler struct {
	writer io.Writer
	level  slog.Level
}

var (
	colorReset  = "\x1b[0m"
	colorTime   = "\x1b[93m" // 时间：浅黄色 (Bright Yellow)
	colorDebug  = "\x1b[36m" // DEBUG：青色
	colorInfo   = "\x1b[32m" // INFO：绿色
	colorWarn   = "\x1b[33m" // WARN：黄色
	colorError  = "\x1b[31m" // ERROR：红色
	colorASR    = "\x1b[35m" // ASR：品红
	colorLLM    = "\x1b[34m" // LLM：蓝色
	colorTTS    = "\x1b[95m" // TTS：亮品红
	colorTiming = "\x1b[92m" // Timing：亮绿色
)

func (h *CustomTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CustomTextHandler) Handle(ctx context.Context, r slog.Record) error {
	// 获取时间戳
	timeStr := r.Time.Format("2006-01-02 15:04:05.000")

	// 获取日志级别
	levelStr := r.Level.String()

	// 应用颜色
	var levelColor string
	switch r.Level {
	case slog.LevelDebug:
		levelColor = colorDebug
	case slog.LevelInfo:
		levelColor = colorInfo
	case slog.LevelWarn:
		levelColor = colorWarn
	case slog.LevelError:
		levelColor = colorError
	default:
		levelColor = colorReset
	}

	// 检查是否是特殊阶段日志
	var stageColor string
	var isStageLog bool
	msg := r.Message

	if strings.HasPrefix(msg, "[ASR]") {
		stageColor = colorASR
		isStageLog = true
	} else if strings.HasPrefix(msg, "[LLM]") {
		stageColor = colorLLM
		isStageLog = true
	} else if strings.HasPrefix(msg, "[TTS]") {
		stageColor = colorTTS
		isStageLog = true
	} else if strings.HasPrefix(msg, "[TIMING]") {
		stageColor = colorTiming
		isStageLog = true
	}

	// 构建输出
	var output string
	if isStageLog {
		// 阶段日志格式: [时间] [阶段] 消息
		output = fmt.Sprintf("%s[%s]%s %s%s%s",
			colorTime, timeStr, colorReset,
			stageColor, msg, colorReset)
	} else {
		// 普通日志格式: [时间] [级别] 消息
		output = fmt.Sprintf("%s[%s]%s %s[%s]%s %s",
			colorTime, timeStr, colorReset,
			levelColor, levelStr, colorReset,
			msg)
	}

	// 添加属性（如果有）
	if r.NumAttrs() > 0 {
		output += " {"
		r.Attrs(func(a slog.Attr) bool {
			output += fmt.Sprintf(" %s=%v", a.Key, a.Value)
			return true
		})
		output += " }"
	}
	output += "\n"

	_, err := h.writer.Write([]byte(output))
	return err
}

func (h *CustomTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h // 简化实现
}

func (h *CustomTextHandler) WithGroup(name string) slog.Handler {
	return h // 简化实现
}

// Logger 日志接口实现
type Logger struct {
	config      *LogCfg
	textLogger  *slog.Logger // 控制台文本输出
	currentDate string       // 当前日期 YYYY-MM-DD
}

// configLogLevelToSlogLevel 将配置中的日志级别转换为slog.Level
func configLogLevelToSlogLevel(configLevel string) slog.Level {
	switch strings.ToLower(configLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// NewLogger 创建新的日志记录器
func NewLogger(config *LogCfg) (*Logger, error) {
	// 设置slog级别
	slogLevel := configLogLevelToSlogLevel(config.LogLevel)

	// 创建自定义文本处理器（用于控制台输出）
	customHandler := &CustomTextHandler{
		writer: os.Stdout,
		level:  slogLevel,
	}

	textLogger := slog.New(customHandler)

	logger := &Logger{
		config:      config,
		textLogger:  textLogger,
		currentDate: time.Now().Format("2006-01-02"),
	}

	if DefaultLogger == nil {
		DefaultLogger = logger
	}

	return logger, nil
}

// log 通用日志记录函数（内部使用）
func (l *Logger) log(level slog.Level, msg string, fields ...any) {
	// 构建slog属性
	var attrs []slog.Attr
	if len(fields) > 0 && fields[0] != nil {
		// 处理fields参数
		if fieldsMap, ok := fields[0].(map[string]any); ok {
			// 提取并排序键
			keys := make([]string, 0, len(fieldsMap))
			for k := range fieldsMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// 按排序后的键顺序添加属性
			for _, k := range keys {
				attrs = append(attrs, slog.Any(k, fieldsMap[k]))
			}
		} else {
			// 如果不是map，直接作为fields字段
			attrs = append(attrs, slog.Any("fields", fields[0]))
		}
	}

	// 同时写入文件（JSON）和控制台（文本）
	ctx := context.Background()
	l.textLogger.LogAttrs(ctx, level, msg, attrs...)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(msg string, args ...any) {
	if l.config.LogLevel == "DEBUG" {
		if len(args) > 0 && containsFormatPlaceholders(msg) {
			formattedMsg := fmt.Sprintf(msg, args...)
			l.log(slog.LevelDebug, formattedMsg)
		} else {
			l.log(slog.LevelDebug, msg, args...)
		}
	}
}

func containsFormatPlaceholders(s string) bool {
	return strings.Contains(s, "%")
}

// Info 记录信息级别日志
func (l *Logger) Info(msg string, args ...any) {
	// 检测是否为格式化模式
	if len(args) > 0 && containsFormatPlaceholders(msg) {
		// 格式化模式：类似 Info
		formattedMsg := fmt.Sprintf(msg, args...)
		l.log(slog.LevelInfo, formattedMsg)
	} else {
		// 结构化模式：原有方式
		l.log(slog.LevelInfo, msg, args...)
	}
}

// Warn 记录警告级别日志
func (l *Logger) Warn(msg string, args ...any) {
	if len(args) > 0 && containsFormatPlaceholders(msg) {
		formattedMsg := fmt.Sprintf(msg, args...)
		l.log(slog.LevelWarn, formattedMsg)
	} else {
		l.log(slog.LevelWarn, msg, args...)
	}
}

// Error 记录错误级别日志
func (l *Logger) Error(msg string, args ...any) {
	if len(args) > 0 && containsFormatPlaceholders(msg) {
		formattedMsg := fmt.Sprintf(msg, args...)
		l.log(slog.LevelError, formattedMsg)
	} else {
		l.log(slog.LevelError, msg, args...)
	}
}

// InfoASR 记录ASR阶段信息日志
func (l *Logger) InfoASR(msg string, args ...any) {
	prefixedMsg := "[ASR] " + msg
	l.Info(prefixedMsg, args...)
}

// InfoLLM 记录LLM阶段信息日志
func (l *Logger) InfoLLM(msg string, args ...any) {
	prefixedMsg := "[LLM] " + msg
	l.Info(prefixedMsg, args...)
}

// InfoTTS 记录TTS阶段信息日志
func (l *Logger) InfoTTS(msg string, args ...any) {
	prefixedMsg := "[TTS] " + msg
	l.Info(prefixedMsg, args...)
}

// InfoTiming 记录计时信息日志
func (l *Logger) InfoTiming(msg string, args ...any) {
	prefixedMsg := "[TIMING] " + msg
	l.Info(prefixedMsg, args...)
}
