package main

import (
	"github.com/google/logger"
	"github.com/jsannemo/omogenexec/runner"
)

func main() {
	sandbox := runner.NewSandbox(runner.SandboxArgs{
		WorkingDirectory: "/tmp/play/working",
		InputPath:        "/tmp/play/in",
		OutputPath:       "/tmp/play/out",
		ErrorPath:        "/tmp/play/err",
		ExtraReadPaths:   nil,
		ExtraWritePaths:  nil,
		TimeLimitMs:      1000,
		MemoryLimitKb:    100 * 1000,
	})
	err := sandbox.Start()
	if err != nil {
		logger.Fatal("error starting sandbox: ", err)
	}
	res, err := sandbox.Run("/usr/bin/g++", []string{"fil.cpp"})
//	res, err := sandbox.Run("/usr/bin/env", []string{})
	if err != nil {
		logger.Infof("stderr %s", sandbox.Logs())
		logger.Fatal("error starting sandbox: ", err)
	}
	logger.Infof("Result: %v", res)
	err = sandbox.Finish()
	if err != nil {
		logger.Fatal("error quitting sandbox: ", err)
	}
}
