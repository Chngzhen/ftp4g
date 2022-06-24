package ftp4g

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Chngzhen/log4g"
	"github.com/secsy/goftp"
)

type FTPClient struct {
	connection *goftp.Client
	// 远程文件存放的根目录。默认：用户主目录。
	remoteBootDir string
	// 本地文件存放的根目录。默认：当前目录。
	localBootDir string
	// 是否在检索远程文件时创建本地目录。默认：true。
	createAbsentParent bool
	// 远程文件的后缀。若不指定，则检索所有文件。默认：[]。
	filterFileExtends []string
}

func buildFtpClient(host string, port int, username, password string) (Client, error) {
	config := goftp.Config{
		User:               username,
		Password:           password,
		ConnectionsPerHost: port,
		Timeout:            10 * time.Second,
		Logger:             nil,
		ActiveTransfers:    false,
	}

	client, err := goftp.DialConfig(config, host)
	if err != nil {
		return nil, err
	}

	return &FTPClient{client, "", "", true, nil}, nil
}

func (c *FTPClient) SetRemoteBootDir(path string) {
	c.remoteBootDir = path
}

func (c *FTPClient) SetLocalBootDir(path string) {
	c.localBootDir = path
}

func (c *FTPClient) SetCreateAbsentParent(b bool) {
	c.createAbsentParent = b
}

func (c *FTPClient) SetFilterFileExtends(extends []string) {
	c.filterFileExtends = extends
}

func (c *FTPClient) RetrieveAll(relativeDirPath string, ftpFileChannel chan<- *SimpleFile) (bool, error) {
	// 获取文件
	var dirFullPath string
	if "" == c.remoteBootDir {
		// 若未指定远程根目录，则根据相对路径获取文件
		dirFullPath = relativeDirPath
	} else {
		// 若指定了远程根目录，则根据根目录和相对路径获取文件
		dirFullPath = c.remoteBootDir + "/" + relativeDirPath
	}
	dirSubFiles, err := c.connection.ReadDir(dirFullPath)
	if err != nil {
		return false, err
	}

	// 遍历文件夹的子目录
	needCheckExt := len(c.filterFileExtends) > 0
	for _, dirSubFile := range dirSubFiles {
		ftpFile := &SimpleFile{
			Name:        dirSubFile.Name(),
			RelativeDir: relativeDirPath,
		}

		if dirSubFile.IsDir() {
			// 目录类型
			ftpFile.IsFile = false
			ftpFileChannel <- ftpFile

			if c.createAbsentParent {
				// 创建本地目录
				var localFilePath string
				if "" == c.localBootDir {
					localFilePath = strings.TrimLeft(ftpFile.RelativeDir+"/"+ftpFile.Name, "/")
				} else {
					localFilePath = c.localBootDir + "/" + ftpFile.RelativeDir + "/" + ftpFile.Name
				}
				if ok, err := checkDir(localFilePath); !ok {
					return false, err
				}
			}

			// 递归检索当前远程目录。若失败，则中止检索。
			dirSubFileRelativePath := relativeDirPath + "/" + dirSubFile.Name()
			if ok, err := c.RetrieveAll(strings.TrimLeft(dirSubFileRelativePath, "/"), ftpFileChannel); !ok {
				return ok, err
			}
		} else if !needCheckExt || checkExt(ftpFile.Name, c.filterFileExtends) {
			// 文件类型，不需要校验文件后缀或文件后缀合法
			ftpFile.IsFile = true
			ftpFileChannel <- ftpFile
		}
	}
	return true, nil
}

func (c *FTPClient) Download(remoteFilePath string, localFilePath string) (bool, error) {
	// 开启本地文件的输出流
	localFile, err := os.OpenFile(localFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return false, err
	}
	defer func(localFile *os.File) {
		err = localFile.Close()
		if err != nil {
			log4g.Error("本地文件[%s]的输出流关闭失败：%+v", localFilePath, err)
		}
	}(localFile)

	// 检索远程文件并写入到本地文件
	err = c.connection.Retrieve(remoteFilePath, localFile)
	if err != nil {
		log4g.Error("%+v", err)
		// 若文件下载失败，删除已创建的本地文件
		if err = os.Remove(localFilePath); err != nil {
			log4g.Error("本地文件[%s]回滚失败", localFilePath)
		}
		return false, errors.New(fmt.Sprintf("本地文件[%s]写入失败", localFilePath))
	}
	return true, nil
}
