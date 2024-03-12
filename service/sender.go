package service

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

/**
send端口:tcp传输文件
send端口+1:udp接收ip
*/

type SendHandler struct {
	State   State
	ip      string
	port    uint16
	fileSrc string
}

var Sender = SendHandler{}

var connR net.PacketConn

// InitSetting 初始化设置
func (r *SendHandler) InitSetting() {
	Log("Init Sender")
	r.State = Stopped
	//设置默认接收端口
	r.port = 32000
	SenderPortInput.SetText(strconv.Itoa(int(r.port)))
	Log("Init Sender Succeed")
}

// RunIpSearcher 启动ip获取
func (r *SendHandler) RunIpSearcher() error {
	Log("Run IP Searcher")
	if r.State == Running {
		err := errors.New("repeated start")
		LogErr(err.Error())
		return err
	}
	//端口检查
	port, err := PortCheck(SenderPortInput.Text)
	if err != nil {
		LogErr("Run Receiver Error:" + err.Error())
		return err
	}
	r.port = port
	r.RunIpReceiver()
	r.State = Running
	Log("Run IP Searcher Succeed")
	return nil
}

// StopIpSearcher 停止ip获取
func (r *SendHandler) StopIpSearcher() {
	Log("Stop IP Searcher")
	if r.State == Stopped {
		err := errors.New("repeated stop")
		LogErr(err.Error())
		return
	}
	if connR != nil {
		connR.Close()
	}
	r.State = Stopped
	Log("Stop IP Succeed")
}
func (r *SendHandler) PortS(offset uint16) string {
	return strconv.FormatUint(uint64(r.port+offset), 10)
}
func (r *SendHandler) SetFileSrc(src string) error {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		r.fileSrc = src
	} else {
		return errors.New("file path is not a file")
	}
	return nil
}

func (r *SendHandler) RunIpReceiver() {
	Log("Run Ip Receiver...")
	var err error
	connR, err = net.ListenPacket("udp", ":"+r.PortS(1))
	if err != nil {
		LogErr("Link error with port: " + r.PortS(1))
		return
	}
	go func() {
		defer func(connR net.PacketConn) {
			if connR != nil {
				connR.Close()
			}
		}(connR)
		for {
			_, addr, err := connR.ReadFrom(make([]byte, 8))
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					Log("RunIpReceiver closed")
				} else {
					LogErr("RunIpReceiver ReadFrom Error: " + err.Error())
				}
				return
			}
			ip, err := ExtractIPPartOfAddress(addr.String())
			if err != nil {
				LogErr("ExtractIPPartOfAddress Error:" + err.Error())
				return
			}
			AddSList(ip)
		}
	}()
}
func (r *SendHandler) SendFile() {
	Log("Start sending files...")
	//检查ip
	err := IpCheck(SIpInput.Text)
	if err != nil {
		LogErr("IP is illegal:" + err.Error())
		return
	}
	//检查文件
	err = r.SetFileSrc(SenderFileSrcInput.Text)
	if err != nil {
		LogErr("Wrong file path:" + err.Error())
		return
	}
	//监听端口
	conn, err := net.Dial("tcp", SIpInput.Text+":"+r.PortS(0))
	if err != nil {
		LogErr("Link error with " + SIpInput.Text + ":" + r.PortS(0) + err.Error())
		return
	}
	defer conn.Close()
	//打开文件
	file, err := os.Open(r.fileSrc)
	if err != nil {
		LogErr("Fail to open file:" + err.Error())
		return
	}
	defer file.Close()
	//发送文件名
	_, err = conn.Write([]byte(filepath.Base(file.Name())))
	if err != nil {
		LogErr("Wrong file name sent:" + err.Error())
		return
	}
	//发送文件大小
	stat, err := file.Stat()
	if err != nil {
		LogErr("Failed to obtain file information:" + err.Error())
		return
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(stat.Size()))
	_, err = conn.Write(buf)
	if err != nil {
		LogErr("Wrong file name sent:" + err.Error())
		return
	}
	//发送文件内容
	_, err = io.Copy(conn, file)
	if err != nil {
		LogErr("Error sending file:" + err.Error())
		return
	}
	Log("Send file completed")
}
