package fti4g

import (
	"errors"
	"fmt"
)

type Client interface {
	// SetRemoteBootDir 设置远程文件存储的根目录。
	SetRemoteBootDir(path string)
	// SetLocalBootDir 设置本地文件存储的根目录。
	SetLocalBootDir(path string)
	// SetCreateAbsentParent 设置是否在检索时创建本地目录。
	SetCreateAbsentParent(b bool)
	// SetFilterFileExtends 设置远程文件的后缀集合。
	SetFilterFileExtends(extends []string)
	// RetrieveAll 基于远程文件存储的根目录检索指定相对目录的所有子目录和文件。
	RetrieveAll(relativeDirPath string, ftpFileChannel chan<- *SimpleFile) (bool, error)
	// Download 将远程文件下载至本地。若下载失败，已创建的本地文件会被删除。
	Download(remoteFilePath string, localFilePath string) (bool, error)
}

func Build(protocol Protocol, host string, port int, user, password string) (Client, error) {
	switch protocol {
	case FTP:
		return buildFtpClient(host, port, user, password)
	case SFTP:
		return buildSftpClient(host, port, user, password)
	default:
		return nil, errors.New(fmt.Sprintf("未定义的协议：%s", protocol))
	}
}