package interfaces

import "io"

type Core interface {

	// UpdateFromLocalFile 从本地文件中进行升级
	UpdateFromLocalFile(filename string) error

	// UpdateFromUrl 从网络进行升级
	UpdateFromUrl(url string) error

	// Update OTA升级
	Update(reader io.Reader) error
}
