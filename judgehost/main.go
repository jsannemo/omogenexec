package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/logger"
	apipb "github.com/jsannemo/omogenexec/api"
	"google.golang.org/grpc"
	"io/ioutil"
	"sync"
)

var (
	log     = logger.Init("omogenexec-judgehost", true, false, ioutil.Discard)
	address = flag.String("listen_addr", "127.0.0.1:61811", "The Run server address to listen to in the format host:port")
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
	res, err := Compile(&apipb.Program{
		Sources: []*apipb.SourceFile{
			{Path: "hello.py", Contents: []byte("print('Hello World!')")},
		},
		Language: apipb.LanguageGroup_PYTHON_3,
	}, "/var/lib/omogen/submissions/13123123/compile")
	if err != nil {
		logger.Fatalf("err: %v", err)
	}
	fmt.Printf("res: %v", res)

	ch := make(chan *apipb.Result)
	evaluator, err := NewEvaluator("/var/lib/omogen/submissions/13123123", &apipb.EvaluationPlan{
		Program: res.Program,
		RootGroup: &apipb.TestGroup{
			Cases: []*apipb.TestCase{
				{
					Name:       "01",
					InputPath:  "/var/lib/omogen/problems/helloworld/data/01.in",
					OutputPath: "/var/lib/omogen/problems/helloworld/data/01.ans",
				},
				{
					Name:       "02",
					InputPath:  "/var/lib/omogen/problems/helloworld/data/01.in",
					OutputPath: "/var/lib/omogen/problems/helloworld/data/01.ans",
				},
			},
			Groups:               nil,
			Name:                 "",
			Score:                0,
			OutputValidatorFlags: nil,
			BreakOnFail:          true,
			ScoringMode:          0,
			VerdictMode:          apipb.VerdictMode_FIRST_ERROR,
			AcceptIfAnyAccepted:  false,
		},
		TimeLimitMs: 1000,
		MemLimitKb:  1000 * 1000,
	}, ch)
	if err != nil {
		logger.Fatalf("eval setup err: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for res := range ch {
			logger.Infof("result: %v", res)
		}
		wg.Done()
	}()
	err = evaluator.Evaluate()
	if err != nil {
		logger.Fatalf("eval err: %v", err)
	}
	wg.Wait()
}
