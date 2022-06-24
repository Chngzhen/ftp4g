package fti4g

import (
	"testing"

	"github.com/Chngzhen/log4g"
)

func TestFTPClient_RetrieveAll(t *testing.T) {
	client, err := buildFtpClient("192.168.21.155", 21, "ffadmin", "Q)ydrsr04Mff")
	if err != nil {
		log4g.Error("FTP客户端创建失败：%+v", err)
		return
	}
	client.SetLocalBootDir("D:/home/ffadmin")

	fileInfos := make(chan *SimpleFile, 100)
	ok, err := client.RetrieveAll("dddatasync/tone", fileInfos)
	if !ok {
		log4g.Error("FTP客户端创建失败：%+v", err)
		return
	}
	close(fileInfos)

	for fileInfo := range fileInfos{
		log4g.Info("%+v", fileInfo)
	}
}