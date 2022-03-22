package test

import (
	"fmt"
	"github.com/ruixiaoedu/ota/config"
	"github.com/ruixiaoedu/ota/core"
	"os"
	"testing"
)

// TestUpdate 测试升级
func TestUpdate(t *testing.T) {

	core := core.NewCore(&config.Config{
		Keyfile: "",
	})

	f, _ := os.Open("ota.tar.gz")
	defer f.Close()

	err := core.Update(f)
	fmt.Println(err)

}
