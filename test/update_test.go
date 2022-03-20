package test

import (
	"fmt"
	"github.com/ruixiaoedu/ota/core"
	"github.com/ruixiaoedu/ota/utils"
	"os"
	"testing"
)

// TestUpdate 测试升级
func TestUpdate(t *testing.T) {

	pubKey, _ := utils.ParsePublicKey([]byte(publicKey))
	core := core.NewCore(pubKey)

	f, _ := os.Open("ota.tar.gz")
	defer f.Close()

	err := core.Update(f)
	fmt.Println(err)

}
