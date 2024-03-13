package service

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func SendFile(src string, writer io.Writer) error {
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
	//计算并发送文件md5
	fileMD5, err := GetFileMD5(file)
	if err != nil {
		return errors.New("Failed to getting file md5:" + err.Error())
	}
	if _, err = writer.Write(fileMD5); err != nil {
		return errors.New("Failed to sent file md5:" + err.Error())
	}
	//发送文件内容
	if _, err = io.Copy(writer, file); err != nil {
		return errors.New("Error sending file:" + err.Error())
	}
	Log("Send file:" + string(fileNameBytes) + " size:" + strconv.FormatInt(stat.Size(), 10) + " md5:" + hex.EncodeToString(fileMD5))
	return nil
}
func ReceiveFile(src string, reader io.Reader) error {
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
	//获取文件md5
	fileMD5 := make([]byte, 16)
	if _, err = reader.Read(fileMD5); err != nil {
		return errors.Join(errors.New("error reading file md5"), err)
	}
	//创建文件
	fPath := src + string(filepath.Separator) + string(fileName)
	newFile, err := os.Create(fPath)
	if err != nil {
		return errors.Join(errors.New("error creating file"), err)
	}
	defer newFile.Close()
	//获取文件内容
	fReader := bufio.NewReader(reader)
	if _, err = io.Copy(newFile, fReader); err != nil {
		//出错删除文件
		err2 := os.Remove(fPath)
		if err2 != nil {
			return errors.Join(errors.New("error occurred while removing incomplete files"), err, err2)
		}
		return errors.Join(errors.New("io copy error"), err)
	}
	//验证文件md5
	if ok, err := CompareFileMD5(newFile, fileMD5); !ok {
		err2 := os.Remove(fPath)
		if err2 != nil {
			return errors.Join(errors.New("error occurred while removing incomplete files"), err, err2)
		}
		if err != nil {
			return errors.Join(errors.New("compare file md5 error"), err)
		}
		return errors.New("file md5 verification failed")
	}
	Log("Received file:" + string(fileName) + " size:" + strconv.FormatInt(num, 10) + " md5:" + hex.EncodeToString(fileMD5))
	return nil
}
func GetFileMD5(file *os.File) ([]byte, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
func CompareFileMD5(file *os.File, md5str []byte) (bool, error) {
	fileMD5, err := GetFileMD5(file)
	if err != nil {
		return false, err
	}
	if bytes.Equal(fileMD5, md5str) {
		return true, nil
	} else {
		return false, nil
	}
}
