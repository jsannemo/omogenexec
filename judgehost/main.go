package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/logger"
	apipb "github.com/jsannemo/omogenexec/api"
	"google.golang.org/grpc"
	"io/ioutil"
)

var (
	log     = logger.Init("omogenexec-judgehost", true, false, ioutil.Discard)
	address = flag.String("listen_addr", "127.0.0.1:61811", "The run server address to listen to in the format host:port")
)

type runServer struct {
}

func (s *runServer) GetLanguages(ctx context.Context, _ *apipb.GetLanguagesRequest) (*apipb.GetLanguagesResponse, error) {
	return nil, nil
}

func (s *runServer) Compile(ctx context.Context, _ *apipb.CompileRequest) (*apipb.CompileResponse, error) {
	return nil, nil
}

func (s *runServer) Evaluate(req *apipb.EvaluateRequest, stream apipb.RunService_EvaluateServer) error {
	return nil
}

func newServer() (*runServer, error) {
	s := &runServer{}
	return s, nil
}

// Register registers a new RunService with the given server.
func Register(grpcServer *grpc.Server) error {
	server, err := newServer()
	if err != nil {
		return err
	}
	apipb.RegisterRunServiceServer(grpcServer, server)
	return nil
}

func main() {
	res, err := compile(&apipb.Program{
		Sources: []*apipb.SourceFile{
			{Path: "hello.py", Contents: []byte("print('Hello World!')")},
		},
		Language: apipb.LanguageGroup_PYTHON_3,
	}, "/var/lib/omogen/submissions/13123123/compile")
	if err != nil {
		logger.Fatalf("err: %v", err)
	}
	fmt.Printf("res: %v", res)
}
