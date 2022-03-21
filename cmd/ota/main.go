package main

import (
	"github.com/ruixiaoedu/ota/config"
	"github.com/ruixiaoedu/ota/core"
	"github.com/ruixiaoedu/ota/models"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
)

var (
	app        = kingpin.New("ota", "over-the-air software updater")
	configFlag = app.Flag("config", "the config of software updater").Short('c').String()

	serviceCommand = app.Command("service", "start with service program")

	updateCommand  = app.Command("update", "run the update program")
	standAloneFlag = updateCommand.Flag("stand-alone", "update without use daemon mode").Bool()
	updateUrlFlag  = updateCommand.Flag("url", "update with web url").Short('u').String()
	updateFileFlag = updateCommand.Flag("file", "update with local file").Short('f').String()
)

func main() {
	app.Version(models.Version)
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	cfgFileName := configFilename()
	if configFlag != nil && *configFlag != "" {
		cfgFileName = *configFlag
	}

	cfg, err := config.NewConfig(cfgFileName)
	if err != nil {
		log.Fatalf("fail to read config file: %v", err)
		return
	}

	c := core.NewCore(cfg)

	switch command {
	case serviceCommand.FullCommand(): // 服务模式
		service(c)
	case updateCommand.FullCommand(): // 升级
		update(c)
	}
}
