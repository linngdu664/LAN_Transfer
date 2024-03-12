package service

import (
	"errors"
	"strconv"
	"strings"
)

type State uint8

const (
	Running State = iota
	Stopped
)

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
