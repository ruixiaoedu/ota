package main

import (
	"github.com/ruixiaoedu/ota/core"
	"github.com/ruixiaoedu/ota/unixsocket"
	"github.com/ruixiaoedu/ota/unixsocket/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strings"
)

// update 升级操作
func update(c *core.Core) {
	// 是否使用独立模式升级
	if standAloneFlag != nil && *standAloneFlag {
		updateStandAlone(c)
		return
	}

	conn, err := grpc.Dial("unix://"+unixsocket.SockAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln("open the grpc unix socket fail: " + err.Error())
		return
	}
	defer conn.Close()

	client := pb.NewOtaClient(conn)

	var (
		url  string
		file string
	)

	if updateUrlFlag != nil {
		url = *updateUrlFlag
	}

	if updateFileFlag != nil {
		file = *updateFileFlag
	}

	// 本地文件模式和线上模式只能任选其一
	var updateURL = ""
	if url != "" && file != "" {
		log.Fatalln("File and URL cannot coexist")
		return
	} else if url == "" && file == "" {
		log.Fatalln("File or URL Mandatory Select either")
		return
	} else if url != "" {
		// 此时为URL模式
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			log.Fatalln("The URL is not the correct address")
			return
		}
		updateURL = url
	} else if file != "" {
		updateURL = "file://" + file
	}

	updateReply, err := client.Update(context.Background(), &pb.UpdateRequest{Url: updateURL})
	if err != nil {
		log.Fatalln("Update fail: " + err.Error())
		return
	} else if !updateReply.Ok {
		log.Fatalln("Update fail: " + updateReply.Message)
		return
	}

}

// updateStandAlone 使用独立模式进行升级
func updateStandAlone(c *core.Core) {

}
