package ftp4g

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Chngzhen/log4g"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPClient struct {
	connection *sftp.Client
	// 远程文件存放的根目录。默认：用户主目录。
	remoteBootDir string
	// 本地文件存放的根目录。默认：当前目录。
	localBootDir string
	// 是否在检索远程文件时创建本地目录。默认：true。
	createAbsentParent bool
	// 远程文件的后缀。若不指定，则检索所有文件。默认：[]。
	filterFileExtends []string
}

func buildSftpClient(host string, port int, username, password string) (Client, error) {
	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			//ssh.KeyboardInteractive(func(_, _ string, questions []string, _ []bool) (answers []string, err error) {
			//	answers = make([]string, len(questions))
			//	for n := range questions {
			//		answers[n] = password
			//	}
			//	return answers, nil
			//}),
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		ClientVersion:   "",
		Timeout:         10 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig)
	if err != nil {
		return nil, err
	}

	client, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, err
	}

	return &SFTPClient{client, "", "", true, nil}, nil
}

func (c *SFTPClient) SetRemoteBootDir(path string) {
	c.remoteBootDir = path
}

func (c *SFTPClient) SetLocalBootDir(path string) {
	c.localBootDir = path
}

func (c *SFTPClient) SetCreateAbsentParent(b bool) {
	c.createAbsentParent = b
}

func (c *SFTPClient) SetFilterFileExtends(extends []string) {
	c.filterFileExtends = extends
}

func (c *SFTPClient) RetrieveAll(relativeDirPath string, ftpFileChannel chan<- *SimpleFile) (bool, error) {
	// 获取文件夹
	var dirFullPath string
	if "" == c.remoteBootDir {
		// 若未指定远程根目录，则根据相对路径获取文件夹
		dirFullPath = relativeDirPath
	} else {
		// 若指定了远程根目录，则根据根目录和相对路径获取文件夹
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
				if yes, err := checkDir(localFilePath); !yes {
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

func (c *SFTPClient) Download(remoteFilePath string, localFilePath string) (bool, error) {
	// 开启本地文件的输出流
	localFile, err := os.OpenFile(localFilePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log4g.Error("%+v", err)
		return false, errors.New(fmt.Sprintf("本地文件[%s]的输出流创建失败", localFilePath))
	}
	defer func(localFile *os.File) {
		err = localFile.Close()
		if err != nil {
			log4g.Error("本地文件[%s]的输出流关闭失败：%+v", localFilePath, err)
		}
	}(localFile)

	// 开启远程文件的输入流
	remoteFile, err := c.connection.Open(remoteFilePath)
	if err != nil {
		log4g.Error("%+v", err)
		return false, errors.New(fmt.Sprintf("远程文件[%s]的输入流创建失败", remoteFilePath))
	}
	defer func(remoteFile *sftp.File) {
		err = remoteFile.Close()
		if err != nil {
			log4g.Error("远程文件[%s]的输入流关闭失败：%+v", remoteFilePath, err)
		}
	}(remoteFile)

	// 下载
	if _, err = remoteFile.WriteTo(localFile); err != nil {
		log4g.Error("%+v", err)
		// 若文件下载失败，删除已创建的本地文件
		if err = os.Remove(localFilePath); err != nil {
			log4g.Error("本地文件[%s]回滚失败", localFilePath)
		}
		return false, errors.New(fmt.Sprintf("本地文件[%s]写入失败", localFilePath))
	}
	return true, nil
}
