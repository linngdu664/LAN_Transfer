package service

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"strings"
	"time"
)

type State uint8

const (
	Running State = iota
	Stopped
)

type IShowSpeed struct {
	meat []int64
	turn int
}

func NewIShowSpeed(meets int) *IShowSpeed {
	return &IShowSpeed{meat: make([]int64, meets), turn: 0}
}
func (r *IShowSpeed) Beat(nNow int64) (sp int64) {
	sp = nNow - r.meat[r.turn]
	r.meat[r.turn] = nNow
	r.nextTurn()
	return
}
func (r *IShowSpeed) nextTurn() {
	r.turn++
	if r.turn >= len(r.meat) {
		r.turn = 0
	}
}

// ProgressBarHook 用于显示把读取调用进度输出到进度条
type ProgressBarHook struct {
	progressBar *widget.ProgressBar
	speedText   *canvas.Text
	target      int64
	now         int64
	closeSignal chan struct{}
}

func NewProgressBarHook(progressBar *widget.ProgressBar, speedText *canvas.Text, target int64) *ProgressBarHook {
	p := &ProgressBarHook{
		progressBar: progressBar,
		speedText:   speedText,
		target:      target,
		now:         0,
		closeSignal: make(chan struct{}),
	}
	go func(pbh *ProgressBarHook) {
		for {
			select {
			case <-pbh.closeSignal:
				pbh.progressBar.SetValue(0)
				return
			case <-time.After(time.Millisecond * 100):
				if pbh.now < pbh.target {
					pbh.progressBar.SetValue(float64(pbh.now) / float64(pbh.target))
				} else if pbh.target == 0 && pbh.progressBar.Value != 0 {
					pbh.progressBar.SetValue(0)
				}
			}
		}
	}(p)
	go func(pbh *ProgressBarHook) {
		cycle := time.Millisecond * 250
		sampling := cycle * 4
		iShowSpeed := NewIShowSpeed(4)
		for {
			select {
			case <-pbh.closeSignal:
				pbh.speedText.Text = "  0.0B/s t:0s"
				pbh.speedText.Refresh()
				return
			case <-time.After(cycle):
				pbh.speedText.Text = FormatSpeedAndArrivalTime(iShowSpeed.Beat(pbh.now), 1, sampling.Milliseconds(), pbh.target-pbh.now)
				pbh.speedText.Refresh()
			}
		}
	}(p)

	return p
}
func (r *ProgressBarHook) Write(p []byte) (n int, err error) {
	r.now += int64(len(p))
	return len(p), nil
}
func (r *ProgressBarHook) Close() {
	close(r.closeSignal)
}

type MultipleProgressBarHook struct {
	*ProgressBarHook
}

func NewMultipleProgressBarHook(progressBar *widget.ProgressBar, speedText *canvas.Text) *MultipleProgressBarHook {
	return &MultipleProgressBarHook{NewProgressBarHook(progressBar, speedText, 0)}
}
func (r *MultipleProgressBarHook) AddPB(num int64) {
	r.target += num
}
func (r *MultipleProgressBarHook) RemovePb(nowN, num int64) {
	r.now -= nowN
	r.target -= num
}

// ReplaceLastOctet 替换ip后缀
func ReplaceLastOctet(ip, lastOctet string) string {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		LogErr("ReplaceLastOctet IP format error")
		return ""
	}
	parts[3] = lastOctet
	return strings.Join(parts, ".")
}

// IpCheck 检查IP是否合法
func IpCheck(ip string) error {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return errors.New("IP format error: Split")
	}
	for _, part := range parts {
		i, err := strconv.Atoi(part)
		if err != nil {
			return errors.New("IP format error: Atoi")
		}
		if i < 0 || i > 255 {
			return errors.New("IP format error: Range")
		}
	}
	return nil
}

// ExtractIPPartOfAddress 提取ip地址部分
func ExtractIPPartOfAddress(row string) (string, error) {
	parts := strings.Split(row, ":")
	err := IpCheck(parts[0])
	if err != nil {
		return "", err
	}
	return parts[0], nil
}

// PortCheck 检查端口是否合法
func PortCheck(port string) (uint16, error) {
	var out uint16
	atoi, err := strconv.Atoi(port)
	if err != nil {
		return 0, errors.New("port number format error")
	}
	if atoi >= 0 && atoi < 65536 {
		out = uint16(atoi)
	} else {
		return 0, errors.New("port number range error")
	}
	return out, nil
}

// FormatByteSpeed 格式化字节传输速率(每秒)
func FormatByteSpeed(bPerSeconds int64, precision int) string {
	return FormatByteSize(bPerSeconds, precision) + "/s"
}

// FormatByteSize 格式化字节数单位
func FormatByteSize(bPerSeconds int64, precision int) string {
	if bPerSeconds < 0 {
		return "Invalid Input"
	}

	// 定义单位
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

	// 将字节数转换为浮点数
	size := float64(bPerSeconds)

	// 获取单位对应的下标
	unitIndex := 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	// 格式化输出
	return fmt.Sprintf("%."+strconv.Itoa(precision)+"f%s", size, units[unitIndex])
}
func FormatSpeedAndArrivalTime(bytes int64, precision int, durationMS int64, surplusBytes int64) string {
	bPerSeconds := (bytes / durationMS) * 1000
	builder := strings.Builder{}
	builder.WriteString("  ")
	builder.WriteString(FormatByteSpeed(bPerSeconds, precision))
	builder.WriteString(" t:")
	builder.WriteString(FormatArrivalTime(bPerSeconds, surplusBytes))
	return builder.String()
}
func FormatArrivalTime(bPerSeconds int64, surplusBytes int64) string {
	if surplusBytes == 0 || bPerSeconds == 0 {
		return FormatSeconds(0)
	}
	second := surplusBytes / bPerSeconds
	if second < 0 {
		return "Invalid Input"
	}
	return FormatSeconds(second)
}
func FormatSeconds(seconds int64) string {
	h := seconds / 3600
	m := seconds % 3600 / 60
	s := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	} else {
		return fmt.Sprintf("%ds", s)
	}
}
