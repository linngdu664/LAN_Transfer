package service

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func bufGet(size int64) []byte {
	if size < 1<<16 {
		return make([]byte, 1<<10)
	} else if size < 1<<23 {
		return make([]byte, 1<<17)
	} else {
		return make([]byte, 1<<24)
	}
}

func SendFile(src string, writer io.Writer) error {
	startTime := time.Now()
	//打开文件
	file, err := os.Open(src)
	if err != nil {
		return errors.New("Fail to open file:" + err.Error())
	}
	defer file.Close()
	//发送文件名大小与文件名
	fileNameBytes := []byte(filepath.Base(file.Name()))
	if len(fileNameBytes) > 255 {
		return errors.New("file name too long")
	}
	if _, err = writer.Write([]byte{byte(len(fileNameBytes))}); err != nil {
		return errors.New("Wrong file name size sent:" + err.Error())
	}
	if _, err = writer.Write(fileNameBytes); err != nil {
		return errors.New("Wrong file name sent:" + err.Error())
	}
	//发送文件大小
	stat, err := file.Stat()
	if err != nil {
		return errors.New("Failed to obtain file information:" + err.Error())
	}
	fileSize := make([]byte, 8)
	binary.BigEndian.PutUint64(fileSize, uint64(stat.Size()))
	if _, err = writer.Write(fileSize); err != nil {
		return errors.New("Wrong file name sent:" + err.Error())
	}
	//计算并发送文件内容与文件md5
	buf := bufGet(stat.Size())
	hash := md5.New()
	hook := NewProgressBarHook(SenderProgressBar, SenderSpeedText, stat.Size())
	multiWriter := io.MultiWriter(writer, hash, hook)
	if _, err = io.CopyBuffer(multiWriter, file, buf); err != nil {
		return errors.New("Error sending file:" + err.Error())
	}
	fileMD5 := hash.Sum(nil)
	if _, err = writer.Write(fileMD5); err != nil {
		return errors.New("Error sending md5:" + err.Error())
	}
	Log("Send file:" + string(fileNameBytes) + " size:" + strconv.FormatInt(stat.Size(), 10) + " totalTime:" + strconv.FormatFloat(float64(time.Now().Sub(startTime).Milliseconds()), 'f', -1, 64) + "ms md5:" + hex.EncodeToString(fileMD5))
	buf = nil
	hook.Close()
	return nil
}
func ReceiveFile(src string, reader io.Reader, pbHook *MultipleProgressBarHook) error {
	startTime := time.Now()
	var err error
	//读取文件名大小
	fileNameLen := make([]byte, 1)
	if _, err = reader.Read(fileNameLen); err != nil {
		return errors.Join(errors.New("error reading file name length"), err)
	}
	//读取文件名
	fileName := make([]byte, fileNameLen[0])
	if _, err = reader.Read(fileName); err != nil {
		return errors.Join(errors.New("error reading file name"), err)
	}
	//fileName = bytes.TrimRight(fileName, "\x00")
	//读取文件大小
	fileSize := make([]byte, 8)
	if _, err = reader.Read(fileSize); err != nil {
		return errors.Join(errors.New("error reading file size"), err)
	}
	num := int64(binary.BigEndian.Uint64(fileSize))
	//读取文件内容
	buf := bufGet(num)
	hash := md5.New()
	fPath := src + string(filepath.Separator) + string(fileName)
	newFile, err := os.Create(fPath)
	if err != nil {
		return errors.Join(errors.New("error creating file"), err)
	}
	defer newFile.Close()
	pbHook.AddPB(num)
	multiWriter := io.MultiWriter(newFile, hash, pbHook)
	if _, err = CopyNBuffer(multiWriter, reader, num, buf); err != nil {
		errF := os.Remove(fPath)
		return errors.Join(errors.New("error reading file"), errF, err)
	}
	pbHook.RemovePb(num)
	//读取并比较md5
	hashSum := hash.Sum(nil)
	fileMD5 := make([]byte, 16)
	if _, err = reader.Read(fileMD5); err != nil {
		errF := os.Remove(fPath)
		return errors.Join(errors.New("error reading file md5"), errF, err)
	}
	if !bytes.Equal(fileMD5, hashSum) {
		errF := os.Remove(fPath)
		return errors.Join(errors.New("error equal file md5"), errF)
	}
	Log("Received file:" + string(fileName) + " size:" + strconv.FormatInt(num, 10) + " totalTime:" + strconv.FormatFloat(float64(time.Now().Sub(startTime).Milliseconds()), 'f', -1, 64) + "ms md5:" + hex.EncodeToString(fileMD5))
	buf = nil
	return nil
}

func CopyNBuffer(dst io.Writer, src io.Reader, n int64, buf []byte) (written int64, err error) {
	written, err = io.CopyBuffer(dst, io.LimitReader(src, n), buf)
	if written == n {
		return n, nil
	}
	if written < n && err == nil {
		// src stopped early; must have been EOF.
		err = io.EOF
	}
	return
}
