package eval

import (
	"bufio"
	"fmt"
	"github.com/google/logger"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type sandboxArgs struct {
	WorkingDirectory string
	InputPath        string
	OutputPath       string
	ErrorPath        string
	ExtraReadPaths   []string
	ExtraWritePaths  []string
	TimeLimitMs      int
	MemoryLimitKb    int
}

type sandboxWrapper struct {
	cmd        *exec.Cmd
	sandboxIn  io.WriteCloser
	sandboxOut *bufio.Scanner
	sandboxErr strings.Builder
	waited     bool
}

func newSandbox(id int, args sandboxArgs) *sandboxWrapper {
	sandboxArgs := []string{
		"--sandbox-id", strconv.Itoa(id),
		"--time-lim-ms", strconv.Itoa(args.TimeLimitMs),
		"--wall-time-lim-ms", strconv.Itoa(args.TimeLimitMs*2 + 1000),
		"--memory-mb", strconv.Itoa(args.MemoryLimitKb),
		"--pid-limit", "10",
		"--inodes", "1000",
		"--blocks", strconv.Itoa(1_000_000_000 / 4096),
	}
	readPaths := args.ExtraReadPaths
	writePaths := args.ExtraWritePaths
	if args.WorkingDirectory != "" {
		sandboxArgs = append(sandboxArgs, "--working-dir", args.WorkingDirectory)
	}
	if args.InputPath != "" {
		sandboxArgs = append(sandboxArgs, "--stdin", args.InputPath)
		readPaths = append(readPaths, filepath.Dir(args.InputPath))
	}
	if args.OutputPath != "" {
		sandboxArgs = append(sandboxArgs, "--stdout", args.OutputPath)
		writePaths = append(writePaths, filepath.Dir(args.OutputPath))
	}
	if args.ErrorPath != "" {
		sandboxArgs = append(sandboxArgs, "--stderr", args.ErrorPath)
		writePaths = append(writePaths, filepath.Dir(args.ErrorPath))
	}
	sandboxArgs = append(sandboxArgs, mountArgs(readPaths, writePaths)...)
	logger.Infof("Sandbox %d running with args %v", id, sandboxArgs)
	cmd := exec.Command("/usr/bin/omogenexec", sandboxArgs...)
	cmd.Env = []string{
		"PATH=/bin:/usr/bin",
	}

	sandbox := &sandboxWrapper{
		cmd: cmd,
	}
	inPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil
	}
	sandbox.sandboxIn = inPipe
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil
	}
	sandbox.sandboxOut = bufio.NewScanner(outPipe)
	sandbox.sandboxOut.Split(bufio.ScanWords)
	cmd.Stderr = &sandbox.sandboxErr
	return sandbox
}

func (s *sandboxWrapper) Start() error {
	return s.cmd.Start()
}

func (s *sandboxWrapper) Run(cmdAndArgs []string) (*execResult, error) {
	logger.Infof("Sandbox executing command %v", cmdAndArgs)
	// Binary format is [#of commands + args] command [0x0] arg [0x0] arg ... [0x0]
	fields := len(cmdAndArgs)
	if fields > 255 {
		logger.Fatal("Too many args to sandbox (%v)", cmdAndArgs)
	}
	if fields == 0 {
		logger.Fatal("Too few args to sandbox (%v)", cmdAndArgs)
	}
	msg := []byte{byte(fields)}
	for _, arg := range cmdAndArgs {
		msg = append(msg, []byte(arg)...)
		msg = append(msg, 0x0)
	}
	// Pipes internally write all the data at once, so no need for retries here
	if _, err := s.sandboxIn.Write(msg); err != nil {
		logger.Infof("Failed writing to the sandbox: %v", err)
	}
	res := &execResult{}
	for {
		tok := s.sandboxToken()
		if tok == "done" {
			break
		} else if tok == "killed" {
			killReason := s.sandboxToken()
			if killReason == "tle" {
				res.ExitType = timedOut
			} else if killReason == "setup" {
				return res, fmt.Errorf("sandbox died during setup: %v", s.logs())
			} else {
				logger.Fatalf("Unrecognized output from sandbox (killed %s)", killReason)
			}
		} else if tok == "code" {
			res.ExitType = exited
			codeStr := s.sandboxToken()
			exitCode, err := strconv.Atoi(codeStr)
			if err != nil {
				logger.Fatalf("Unrecognized output from sandbox (code %s)", codeStr)
			}
			res.ExitCode = exitCode
		} else if tok == "signal" {
			res.ExitType = signaled
			signalStr := s.sandboxToken()
			signal, err := strconv.Atoi(signalStr)
			if err != nil {
				logger.Fatalf("Unrecognized output from sandbox (signal %s)", signalStr)
			}
			res.Signal = signal
		} else if tok == "mem" {
			memStr := s.sandboxToken()
			mem, err := strconv.Atoi(memStr)
			if err != nil {
				logger.Fatalf("Unrecognized output from sandbox (mem %s)", memStr)
			}
			res.MemoryUsageKb = mem / 1000
		} else if tok == "cpu" {
			cpuStr := s.sandboxToken()
			cpu, err := strconv.ParseInt(cpuStr, 10, 64)
			if err != nil {
				logger.Fatalf("Unrecognized output from sandbox (cpu %s)", cpuStr)
			}
			res.TimeUsageMs = cpu
		}
	}
	return res, nil
}

func (s *sandboxWrapper) sandboxToken() string {
	if !s.sandboxOut.Scan() {
		s.Finish()
		logger.Fatalf("Failed reading to the sandbox: %v", s.sandboxErr.String())
	}
	return s.sandboxOut.Text()
}

func (s *sandboxWrapper) Finish() {
	if !s.waited {
		s.waited = true
		if err := s.sandboxIn.Close(); err != nil {
			panic(err)
		}
		s.cmd.Wait()
	}
}

func (s *sandboxWrapper) logs() string {
	return s.sandboxErr.String()
}

func mountArgs(readable, writable []string) []string {
	seen := make(map[string]bool)
	var args []string
	// Writable first, in case a path exists in both
	for _, path := range writable {
		if len(path) == 0 {
			continue
		}
		if seen[path] {
			continue
		}
		seen[path] = true
		args = append(args, "--writable", path)
	}
	for _, path := range readable {
		if len(path) == 0 {
			continue
		}
		if seen[path] {
			continue
		}
		seen[path] = true
		args = append(args, "--readable", path)
	}
	return args
}
