package service

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"
)

/**
Receive端口:tcp接收文件
Receive端口+1:循环udp广播ip
*/

type ReceiveHandler struct {
	state   State
	port    uint16
	fileSrc string
}

var Receiver = ReceiveHandler{}

var listener net.Listener
var stopBroadcastIp = make(chan struct{})

// InitSetting 初始化设置
func (r *ReceiveHandler) InitSetting() {
	Log("Init Receiver")
	r.state = Stopped
	//设置默认接收端口
	r.port = 32000
	ReceiverPortInput.SetText(strconv.Itoa(int(r.port)))
	//设置默认下载路径
	currentUser, err := user.Current()
	if err != nil {
		LogErr("Unable to obtain current user information" + err.Error())
	} else {
		// 获取默认下载目录路径
		r.fileSrc = filepath.Join(currentUser.HomeDir, "Downloads")
		ReceiverFileSrcInput.SetText(r.fileSrc)
	}
	r.GetLanIp()
	//r.AutofillIp()
	Log("Init Receiver Succeed")
}

// Run 启动接收
func (r *ReceiveHandler) Run() error {
	Log("Run Receiver")
	if r.state == Running {
		err := errors.New("repeated start")
		LogErr(err.Error())
		return err
	}
	//端口检查
	port, err := PortCheck(ReceiverPortInput.Text)
	if err != nil {
		LogErr("Run Receiver Error:" + err.Error())
		return err
	}
	r.port = port
	//文件路径检查
	err = r.SetFileSrc(ReceiverFileSrcInput.Text)
	if err != nil {
		LogErr("File directory check failed:" + err.Error())
		return err
	}

	r.StartBroadcastIp()
	r.RunReceiveFile()
	ReceiverPortInput.Disable()
	ReceiverFileSrcInput.Disable()
	RIpInput.Disable()
	ReceiverFileSelectBtn.Disable()
	r.state = Running
	Log("Run Receiver Succeed")
	return nil
}

// Stop 停止接收
func (r *ReceiveHandler) Stop() {
	if r.state == Stopped {
		err := errors.New("repeated stop")
		LogErr(err.Error())
		return
	}
	Log("Stop Receiver")
	stopBroadcastIp <- struct{}{}
	if listener != nil {
		listener.Close()
	}
	ReceiverPortInput.Enable()
	ReceiverFileSrcInput.Enable()
	RIpInput.Enable()
	ReceiverFileSelectBtn.Enable()
	r.state = Stopped
	Log("Stop Receiver Succeed")
}

func (r *ReceiveHandler) PortS(offset uint16) string {
	return strconv.FormatUint(uint64(r.port+offset), 10)
}

func (r *ReceiveHandler) SetFileSrc(src string) error {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		r.fileSrc = src
	} else {
		return errors.New("file path is not a folder")
	}
	return nil
}

// GetLanIp 获取局域网ip到列表
func (r *ReceiveHandler) GetLanIp() {
	localIpList := make([]string, 0)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		LogErr("Error obtaining local IP address:" + err.Error())
		return
	}
	Log("Get LAN address:")
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.IsPrivate() && ipnet.IP.To4() != nil {
			lanIp := ReplaceLastOctet(ipnet.IP.String(), "255")
			Log("LAN:" + lanIp)
			localIpList = append(localIpList, lanIp)
		}
	}
	RefreshRList(localIpList)
}

// Deprecated: 没用
// AutofillIp 测试局域网并尝试填入ip
func (r *ReceiveHandler) AutofillIp() {
	Log("Start auto fill IP, port:" + r.PortS(2))
	readyReceive := make(chan struct{}, 1)
	go func() {
		connR, err := net.ListenPacket("udp", ":"+r.PortS(2))
		if err != nil {
			LogErr("Link error:" + err.Error())
			readyReceive <- struct{}{}
			return
		}
		readyReceive <- struct{}{}
		_, addr, err := connR.ReadFrom(make([]byte, 8))
		if err != nil {
			LogErr("Link Read error:" + err.Error())
			readyReceive <- struct{}{}
			return
		}
		lanIp := ReplaceLastOctet(addr.String(), "255")
		Log("Auto fill LAN IP:" + lanIp)
		RIpInput.SetText(lanIp)
		connR.Close()
	}()
	<-readyReceive

	for _, listIp := range RListItems {
		addr := listIp + ":" + r.PortS(2)
		connW, err := net.Dial("udp", addr)
		if err != nil {
			LogErr("Link error with " + addr + " " + err.Error())
			continue
		}
		_, err = fmt.Fprintf(connW, "x")
		if err != nil {
			LogErr("Link write error with " + addr + " " + err.Error())
			continue
		}
		connW.Close()
	}
	Log("Auto fill IP: send completed")
}

// StartBroadcastIp 开始向局域网广播ip
func (r *ReceiveHandler) StartBroadcastIp() {
	go func() {
		sendAllLAN := false
		Log("Start Broadcast Ip...")
		//ip检查

		if RIpInput.Text == "" && len(RListItems) > 0 {
			sendAllLAN = true
			Log("IP is not filled in, start LAN traversal sending mode")
		} else {
			err := IpCheck(RIpInput.Text)
			if err != nil {
				LogErr(err.Error())
				return
			}
		}
		for {
			select {
			case <-stopBroadcastIp:
				Log("Stop Broadcast Ip")
				return
			default:
				if sendAllLAN {
					for _, listIp := range RListItems {
						addr := listIp + ":" + r.PortS(1)
						connW, err := net.Dial("udp", addr)
						if err != nil {
							LogErr("Link error with " + addr + " " + err.Error())
							continue
						}
						_, err = fmt.Fprintf(connW, "x")
						if err != nil {
							LogErr("Link write error with " + addr + " " + err.Error())
							continue
						}
						connW.Close()
					}
				} else {
					addr := RIpInput.Text + ":" + r.PortS(1)
					connW, err := net.Dial("udp", addr)
					if err != nil {
						LogErr("Link error with " + addr + " " + err.Error())
						time.Sleep(time.Duration(rand.Float32()*100) * time.Millisecond)
						continue
					}
					_, err = fmt.Fprintf(connW, "x")
					if err != nil {
						LogErr("Link write error with " + addr + " " + err.Error())
						continue
					}
					connW.Close()
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
}

// RunReceiveFile 启动等待接收文件
func (r *ReceiveHandler) RunReceiveFile() {
	Log("Start listening to receive files...")
	logCloseOrErr := func(err error, errMsg string) {
		if errors.Is(err, net.ErrClosed) {
			Log(errMsg + " runReceiveFile Accept closed")
		} else {
			LogErr(errMsg + err.Error())
		}
	}
	var err error
	listener, err = net.Listen("tcp", ":"+r.PortS(0))
	if err != nil {
		LogErr("runReceiveFile Listen fail:" + err.Error())
		return
	}
	go func() {
		defer listener.Close()
		pbHook := NewMultipleProgressBarHook(ReceiverProgressBar, ReceiverSpeedText)
		for {
			conn, err := listener.Accept()
			if err != nil {
				logCloseOrErr(err, "runReceiveFile Accept stop:")
				break
			}
			Log("Start receiving files from:" + conn.RemoteAddr().String())
			go func(conn net.Conn) {
				defer conn.Close()
				//设置超时时间
				//err2 := conn.SetReadDeadline(time.Now().Add(time.Second))
				//if err2 != nil {
				//	logCloseOrErr(err2, "set time deadline error:")
				//}

				err2 := ReceiveFile(r.fileSrc, conn, pbHook)
				if err2 != nil {
					logCloseOrErr(err2, "receive file err:")
				}

			}(conn)
		}
		pbHook.Close()
	}()
}
