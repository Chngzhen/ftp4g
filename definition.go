package ftp4g

type Protocol string

const FTP = "ftp"
const SFTP = "sftp"

type SimpleFile struct {
	// 文件名（带后缀）
	Name string
	// 文件的相对于检索起始目录的路径
	RelativeDir string
	// 是否是文件
	IsFile bool
}
