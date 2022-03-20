package test

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"github.com/ruixiaoedu/ota/models"
	"github.com/ruixiaoedu/ota/utils"
	"io"
	"os"

	"testing"
)

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDR6LMzOlH0rVwebznH13s+Nbu862AyIdqEzeJFq7OTQXMAAXOv
V5cgl6BnJy11F8QGMPW48rAkOfvEv8vHCnY3jpKMwCkFoSJvxBeG1Z0ca9OEyhEk
470uJPk+bwaB28WbsrD+lbZFHdth5L1iLmivKgnPhY3uJEB+5b8PD5BlfwIDAQAB
AoGBAKiRrUdYcHSDu9SdEdPQ0iI1WJzwkQHxeeDozeuRZda92rKId/S57J255pCw
P6sm+L7YFpz+GEIfZnasZ+NiHWgutWL5gG0coHd5gUJLMf9nAtf1xTvAbGFCtQZU
qMNi3K5rB5Pz1Ds83aaXJadiEpkwyogLD/5sFRGTi34vx9ABAkEA++JQ0sGJYkGh
dFZpuo/TEyymQ2i7New43Mk5dpUasFWcpDrUvq35BTao41XLAxBeTBLPx6V5tO3m
WWWzyrlJfwJBANVWyzEMR5TM67hl5HmehGvasmU+vxXXGSUYoyTx0PT6afjJdkct
5FQ3FmSpRfpzX7Smb1jFxChn5wr3oJ2e5AECQQDVnJHEmoNDQ7uD6QDTScPMwBHk
mv4hdcqnWzOTYFH49zHXiVkAuJO2GyvRV+HKIGiIBXAWtTvo99RhPkHii45LAkAV
Azp6N0JpppFlFSweyn0yflTp4fdCOHByle2juumg53U+muE6e4usu8xJ195bn7eC
fI4lCT2b2TgJfYBlZfwBAkEA1mrM9HVvvRNxm6DIPEcRdzT/ehuUh+z70+nyagCe
hQYohPdueDFCxrIQ99poWyV6TyNGJV/bUw7bIbR1zRbuqQ==
-----END RSA PRIVATE KEY-----`

const publicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDR6LMzOlH0rVwebznH13s+Nbu8
62AyIdqEzeJFq7OTQXMAAXOvV5cgl6BnJy11F8QGMPW48rAkOfvEv8vHCnY3jpKM
wCkFoSJvxBeG1Z0ca9OEyhEk470uJPk+bwaB28WbsrD+lbZFHdth5L1iLmivKgnP
hY3uJEB+5b8PD5BlfwIDAQAB
-----END PUBLIC KEY-----
`

// TestCreateFile 测试生成打包文件
func TestCreateFile(t *testing.T) {
	var scripts []models.Script

	f, _ := os.Open("preinstall.sh")
	md5, _ := utils.Md5FromReader(f)
	f.Close()

	f, _ = os.Open("preinstall.sh")
	sha256, _ := utils.Sha256FromReader(f)
	f.Close()

	scripts = append(scripts, models.Script{
		Filename: "preinstall.sh",
		Type:     "preinstall",
		Md5:      md5,
		Sha256:   sha256,
	})

	f, _ = os.Open("postinstall.sh")
	md5, _ = utils.Md5FromReader(f)
	f.Close()

	f, _ = os.Open("postinstall.sh")
	sha256, _ = utils.Sha256FromReader(f)
	f.Close()

	scripts = append(scripts, models.Script{
		Filename: "postinstall.sh",
		Type:     "postinstall",
		Md5:      md5,
		Sha256:   sha256,
	})

	var files []models.File

	f, _ = os.Open("otatest")
	md5, _ = utils.Md5FromReader(f)
	f.Close()

	f, _ = os.Open("otatest")
	sha256, _ = utils.Sha256FromReader(f)
	f.Close()

	files = append(files, models.File{
		Filename: "otatest",
		Path:     "/usr/sbin/otatest",
		Md5:      md5,
		Sha256:   sha256,
	})

	des := models.Description{
		Name:        "test",
		Version:     "1.0.0",
		Description: "test file",
		Files:       files,
		Scripts:     scripts,
	}

	bs, _ := json.Marshal(des)

	prv, _ := utils.ParsePrivateKey([]byte(privateKey))
	sign, _ := utils.SignWithSha256(bs, prv)

	d, _ := os.Create("ota.tar.gz")
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	tw.WriteHeader(&tar.Header{
		Name: "ota-description.json",
		Size: int64(len(bs)),
	})
	tw.Write(bs)

	tw.WriteHeader(&tar.Header{
		Name: "ota-description.sig",
		Size: int64(len(sign)),
	})
	tw.Write([]byte(sign))

	f, _ = os.Open("otatest")
	info, _ := f.Stat()
	header, _ := tar.FileInfoHeader(info, "")
	header.Name = "otatest"
	tw.WriteHeader(header)
	io.Copy(tw, f)
	f.Close()

	f, _ = os.Open("postinstall.sh")
	info, _ = f.Stat()
	header, _ = tar.FileInfoHeader(info, "")
	header.Name = "postinstall.sh"
	tw.WriteHeader(header)
	io.Copy(tw, f)
	f.Close()

	f, _ = os.Open("preinstall.sh")
	info, _ = f.Stat()
	header, _ = tar.FileInfoHeader(info, "")
	header.Name = "preinstall.sh"
	tw.WriteHeader(header)
	io.Copy(tw, f)
	f.Close()
}
