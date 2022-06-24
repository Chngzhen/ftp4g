package ftp4g

import (
	"os"
	"strings"
)

func checkExt(fileName string, extends []string) bool {
	for _, ext := range extends {
		if strings.HasSuffix(fileName, "."+ext) {
			return true
		}
	}
	return false
}

func checkDir(dirPath string) (bool, error) {
	if _, ok := existsFile(dirPath); !ok {
		if err := os.MkdirAll(dirPath, 0777); err != nil {
			return false, err
		}
	}
	return true, nil
}

// ExistsFile 判断文件是否存在。注意，若无权访问目标文件，依旧会返回true。
func existsFile(filePath string) (*os.FileInfo, bool) {
	fileInfo, err := os.Lstat(filePath)
	return &fileInfo, !os.IsNotExist(err)
}
