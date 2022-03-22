package core

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ruixiaoedu/ota/config"
	"github.com/ruixiaoedu/ota/models"
	"github.com/ruixiaoedu/ota/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

// NewCore 创建核心程序
func NewCore(cfg *config.Config) *Core {

	var publicKey *rsa.PublicKey = nil
	if cfg.Keyfile != "" {
		var err error
		publicKey, err = utils.ParsePublicKeyFromFile(cfg.Keyfile)
		if err != nil {
			log.Fatal("public key init fail: " + err.Error())
		}
	}

	return &Core{
		pubKey: publicKey,
	}
}

// Core 核心
type Core struct {
	pubKey *rsa.PublicKey // 验签用的公钥
}

// UpdateFromLocalFile 从本地文件中进行升级
func (core *Core) UpdateFromLocalFile(filename string) error {

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return core.Update(f)
}

// UpdateFromUrl 从网络进行升级
func (core *Core) UpdateFromUrl(url string) error {

	resp, err := http.Get(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return core.Update(resp.Body)
}

// Update OTA升级
func (core *Core) Update(reader io.Reader) error {
	// 读取压缩数据
	gr, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer gr.Close()

	// 创建临时目录
	tempDir, err := ioutil.TempDir("", "ota-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// 解析tar包内容
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		filename := path.Join(tempDir, hdr.Name)
		file, err := utils.CreateFile(filename)
		if err != nil {
			return err
		}
		_, err = io.Copy(file, tr)
		if err != nil {
			return err
		}
		if err = file.Chmod(os.FileMode(hdr.Mode) | os.FileMode(0444)); err != nil {
			log.Println("chmod file mode fail", err)
		}
	}

	return core.updateFromDir(tempDir)
}

// updateFromDir 从文件夹中升级
func (core *Core) updateFromDir(dir string) error {
	var err error

	// OTA描述文件是否存在
	var descriptionByte []byte
	var desFilePath = path.Join(dir, "ota-description.json")
	descriptionByte, err = ioutil.ReadFile(desFilePath)
	if err != nil {
		return err
	}

	// OTA签名是否存在，如果存在，则验证签名的正确性
	var sigFilePath = path.Join(dir, "ota-description.sig")
	if utils.FileExist(sigFilePath) {
		if core.pubKey == nil {
			return errors.New("update file has sign, but public key is empty")
		}

		var bs []byte
		bs, err = ioutil.ReadFile(sigFilePath)
		if err != nil {
			return err
		}

		if !utils.VerifySignWithSha256(descriptionByte, strings.TrimSpace(string(bs)), core.pubKey) {
			return errors.New("sign is not right")
		}
	}

	// 解析description文件
	var description models.Description
	err = json.Unmarshal(descriptionByte, &description)
	if err != nil {
		return err
	}

	// 验证文件
	var files []struct {
		Filename string
		Md5      string
		Sha256   string
	}

	var preinstalls []models.Script
	var postinstalls []models.Script

	for _, v := range description.Files {
		if !utils.FileExist(path.Join(dir, v.Filename)) {
			return errors.New("文件不存在")
		}

		files = append(files, struct {
			Filename string
			Md5      string
			Sha256   string
		}{Filename: v.Filename, Md5: v.Md5, Sha256: v.Sha256})
	}

	for _, v := range description.Scripts {
		file, err := os.Open(path.Join(dir, v.Filename))
		if err != nil {
			return err
		}
		fi, err := file.Stat()
		if err != nil {
			file.Close()
			return err
		}

		switch v.Type {
		case "preinstall":
			preinstalls = append(preinstalls, v)
		case "postinstall":
			postinstalls = append(postinstalls, v)
		default:
			file.Close()
			return errors.New("无效的type")
		}

		if err = file.Chmod(fi.Mode() | os.FileMode(0111)); err != nil {
			file.Close()
			log.Println("chmod file mode fail", err)
		}

		files = append(files, struct {
			Filename string
			Md5      string
			Sha256   string
		}{Filename: v.Filename, Md5: v.Md5, Sha256: v.Sha256})
	}

	for _, v := range files {

		// 验证MD5
		if v.Md5 != "" {
			f, err := os.Open(path.Join(dir, v.Filename))
			if err != nil {
				return err
			}
			hex, err := utils.Md5FromReader(f)
			if err != nil {
				f.Close()
				return err
			}
			if hex != v.Md5 {
				f.Close()
				return fmt.Errorf("%s md5 is not right", v.Filename)
			}
			f.Close()
		}

		// 验证SHA256
		if v.Sha256 != "" {
			f, err := os.Open(path.Join(dir, v.Filename))
			if err != nil {
				return err
			}
			hex, err := utils.Sha256FromReader(f)
			if err != nil {
				f.Close()
				return err
			}
			if hex != v.Sha256 {
				f.Close()
				return fmt.Errorf("%s sha256 is not right", v.Filename)
			}
			f.Close()
		}
	}

	// 执行预执行文件
	for _, v := range preinstalls {
		execute(path.Join(dir, v.Filename))
	}

	// 复制文件
	for _, v := range description.Files {
		destination, err := utils.CreateFile(v.Path)
		if err != nil {
			return err
		}
		source, err := os.Open(path.Join(dir, v.Filename))
		if err != nil {
			destination.Close()
			return err
		}
		if _, err = io.Copy(destination, source); err != nil {
			destination.Close()
			source.Close()
			return err
		}
		destination.Close()
		source.Close()
	}

	// 执行完成执行文件
	for _, v := range postinstalls {
		execute(path.Join(dir, v.Filename))
	}

	return err
}

// 打印输出
func asyncLog(reader io.ReadCloser) error {
	cache := ""
	buf := make([]byte, 1024, 1024)
	for {
		num, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "closed") {
				err = nil
			}
			return err
		}
		if num > 0 {
			oByte := buf[:num]
			oSlice := strings.Split(string(oByte), "\n")
			line := strings.Join(oSlice[:len(oSlice)-1], "\n")
			fmt.Printf("%s%s\n", cache, line)
			cache = oSlice[len(oSlice)-1]
		}
	}
}

// 执行文件
func execute(script string) error {

	// 检测脚本是否有可执行权限
	if fileInfo, err := os.Stat(script); err != nil {
		return err
	} else if uint32(fileInfo.Mode().Perm()&os.FileMode(73)) != uint32(73) {
		file, err := os.Open(script)
		if err != nil {
			return err
		}
		if err = file.Chmod(fileInfo.Mode() | os.FileMode(73)); err != nil {
			log.Println("chmod file mode fail", err)
		}
		file.Close()
	}

	cmd := exec.Command("sh", "-c", script)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %s......", err.Error())
		return err
	}

	go asyncLog(stdout)
	go asyncLog(stderr)

	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command execution: %s......", err.Error())
		return err
	}

	return nil
}
