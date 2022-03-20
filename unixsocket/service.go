package unixsocket

import (
	"github.com/ruixiaoedu/ota/interfaces"
	"github.com/ruixiaoedu/ota/unixsocket/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"strings"
)

func NewService(c interfaces.Core) *Service {
	return &Service{
		core: c,
	}
}

type Service struct {
	core interfaces.Core // 核心
}

func (s *Service) Update(ctx context.Context, up *pb.UpdateRequest) (*pb.UpdateReply, error) {

	us := strings.SplitN(up.Url, "://", 2)
	if len(us) < 2 {
		return &pb.UpdateReply{
			Ok:      false,
			Message: "this is not the correct URL",
		}, nil
	}

	switch us[0] {
	case "file":
		if err := s.core.UpdateFromLocalFile(us[1]); err != nil {
			return &pb.UpdateReply{
				Ok:      false,
				Message: err.Error(),
			}, nil
		}
	case "http", "https":
		if err := s.core.UpdateFromUrl(up.Url); err != nil {
			return &pb.UpdateReply{
				Ok:      false,
				Message: err.Error(),
			}, nil
		}
	default:
		return &pb.UpdateReply{
			Ok:      false,
			Message: "this is not the correct URL",
		}, nil
	}

	return &pb.UpdateReply{
		Ok:      true,
		Message: "OK",
	}, nil
}

func (s *Service) Server() error {

	// 创建gRPC服务器
	gs := grpc.NewServer()

	// 注册服务
	pb.RegisterOtaServer(gs, s)
	reflection.Register(gs)

	// 监听Unix Socket
	lis, err := s.listen()
	if err != nil {
		return err
	}

	log.Println("grpc service is starting...")
	if err := gs.Serve(lis); err != nil {
		log.Printf("run service fail: %s", err)
		return err
	}

	return nil
}

// listen 监听UNIX套接字
func (s *Service) listen() (*net.UnixListener, error) {
	_ = os.Remove(SockAddr)
	unixAddr, err := net.ResolveUnixAddr("unix", SockAddr)
	if err != nil {
		log.Fatalln("无效的套接字", err)
	}

	listener, err := net.ListenUnix("unix", unixAddr)
	if err != nil {
		log.Fatalln("监听套接字失败", err)
	}
	return listener, nil
}
