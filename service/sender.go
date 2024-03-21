package service

import (
	"errors"
	"net"
	"os"
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

var SListItemEnable = true

var Sender = SendHandler{}

var connR net.PacketConn
var connSendFile net.Conn

// InitSetting 初始化设置
func (r *SendHandler) InitSetting() {
	Log("Init Sender")
	r.State = Stopped
	//设置默认接收端口
	r.port = 32000
	SenderPortInput.SetText(strconv.Itoa(int(r.port)))
	StopSendFileBtn.Disable()
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
			if AddSList(ip) {
				Log("Get ip:" + ip)
			}
		}
	}()
}
func (r *SendHandler) SendFile() {
	go func() {
		SIpInput.Disable()
		SenderPortInput.Disable()
		SenderFileSelectBtn.Disable()
		SenderFileSrcInput.Disable()
		SendFileBtn.Disable()
		StopSendFileBtn.Enable()
		SListItemEnable = false
		defer func() {
			SIpInput.Enable()
			SenderPortInput.Enable()
			SenderFileSelectBtn.Enable()
			SenderFileSrcInput.Enable()
			SendFileBtn.Enable()
			StopSendFileBtn.Disable()
			SListItemEnable = true
		}()
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
		connSendFile, err = net.Dial("tcp", SIpInput.Text+":"+r.PortS(0))
		if err != nil {
			LogErr("Link error with " + SIpInput.Text + ":" + r.PortS(0) + err.Error())
			return
		}
		defer connSendFile.Close()
		err = SendFile(r.fileSrc, connSendFile)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				Log("Send File Stopped")
			} else {
				LogErr(err.Error())
			}
		}
	}()
}
func (r *SendHandler) StopSendFile() {
	conn, err := net.Dial("udp", SIpInput.Text+":"+r.PortS(2))
	if err != nil {
		LogErr("Link error with " + SIpInput.Text + ":" + r.PortS(2) + err.Error())
		return
	}
	_, err = conn.Write([]byte{'1'})
	if err != nil {
		LogErr("Link write with " + SIpInput.Text + ":" + r.PortS(2) + err.Error())
		return
	}
	if connSendFile != nil {
		connSendFile.Close()
	}
	Log("Stop Send File")
}
