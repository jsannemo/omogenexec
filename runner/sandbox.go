package runner

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

// An ExitType describes why a program exited.
type ExitType int

const (
	// Exited means the program exited normally with an exit code.
	Exited ExitType = iota
	// Signaled means the program was killed by a signal.
	Signaled
	// TimedOut means the program was killed due to exceeding its time limit.
	TimedOut
)

// An ExecResult describes the result of a single execution.
type ExecResult struct {
	// How how the program exited.
	ExitType ExitType
	// The exit code. Only set if the program exited with a code.
	ExitCode int
	// The termination signal. Only set if the program exited with a signal.
	Signal int
	// The time the execution used.
	TimeUsageMs int
	// The memory the execution used.
	MemoryUsageKb int
}

// CrashedWith checks whether the program exited normally with the given code.
func (res ExecResult) CrashedWith(code int) bool {
	return res.ExitType == Exited && res.ExitCode == code
}

// Crashed checks whether the program exited abnormally.
func (res ExecResult) Crashed() bool {
	return (res.ExitType == Exited && res.ExitCode != 0) || res.ExitType == Signaled
}

// TimedOut checks whether the program exceeded its time limit or not.
func (res ExecResult) TimedOut() bool {
	return res.ExitType == TimedOut
}

type SandboxArgs struct {
	WorkingDirectory string
	InputPath        string
	OutputPath       string
	ErrorPath        string
	ExtraReadPaths   []string
	ExtraWritePaths  []string
	TimeLimitMs      int
	MemoryLimitKb    int
}

type Sandbox struct {
	cmd        *exec.Cmd
	sandboxIn  io.WriteCloser
	sandboxOut *bufio.Scanner
	sandboxErr strings.Builder
}

func NewSandbox(args SandboxArgs) *Sandbox {
	sandboxArgs := []string{
		"--sandbox-id", "0",
		"--working-dir", args.WorkingDirectory,
		"--stdin", args.InputPath,
		"--stdout", args.OutputPath,
		"--stderr", args.ErrorPath,
		"--time-lim-ms", strconv.Itoa(args.TimeLimitMs),
		"--wall-time-lim-ms", strconv.Itoa(args.TimeLimitMs*2 + 1000),
		"--memory-mb", strconv.Itoa(args.MemoryLimitKb),
		"--pid-limit", "10",
		"--inodes", "1000",
		"--blocks", strconv.Itoa(1_000_000_000 / 4096),
	}
	sandboxArgs = append(sandboxArgs,
		mountArgs(
			append(args.ExtraReadPaths, filepath.Dir(args.InputPath)),
			append(args.ExtraWritePaths, filepath.Dir(args.OutputPath), filepath.Dir(args.ErrorPath)))...)
	cmd := exec.Command("/usr/bin/omogenexec", sandboxArgs...)
	sandbox := &Sandbox{
		cmd: cmd,
	}
	cmd.Env = []string{
		"PATH=/bin:/usr/bin",
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

func (s *Sandbox) Start() error {
	return s.cmd.Start()
}

func (s *Sandbox) Run(cmd string, args []string) (ExecResult, error) {
	fields := 1 + len(args)
	if fields > 255 {
		logger.Fatal("Too many args to sandbox (%v)", args)
	}
	msg := []byte{byte(fields)}
	msg = append(msg, []byte(cmd)...)
	msg = append(msg, 0x0)
	for _, arg := range args {
		msg = append(msg, []byte(arg)...)
		msg = append(msg, 0x0)
	}
	// Pipes internally write all the data at once
	if _, err := s.sandboxIn.Write(msg); err != nil {
		logger.Fatalf("Failed writing to the sandbox: %v", err)
	}
	res := ExecResult{}
	for {
		tok := s.sandboxToken()
		if tok == "done" {
			break
		} else if tok == "killed" {
			killReason := s.sandboxToken()
			if killReason == "tle" {
				res.ExitType = TimedOut
			} else if killReason == "setup" {
				return res, fmt.Errorf("sandbox run died during setup (misconfigured?)")
			} else {
				logger.Fatalf("Unrecognized output from sandbox (killed %s)", killReason)
			}
		} else if tok == "code" {
			res.ExitType = Exited
			codeStr := s.sandboxToken()
			exitCode, err := strconv.Atoi(codeStr)
			if err != nil {
				logger.Fatalf("Unrecognized output from sandbox (code %s)", codeStr)
			}
			res.ExitCode = exitCode
		} else if tok == "signal" {
			res.ExitType = Signaled
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
			cpu, err := strconv.Atoi(cpuStr)
			if err != nil {
				logger.Fatalf("Unrecognized output from sandbox (cpu %s)", cpuStr)
			}
			res.TimeUsageMs = cpu
		}
	}
	return res, nil
}

func (s *Sandbox) sandboxToken() string {
	if !s.sandboxOut.Scan() {
		logger.Fatalf("Failed reading to the sandbox: %v", s.sandboxOut.Err())
	}
	return s.sandboxOut.Text()
}

func (s *Sandbox) Finish() error {
	return s.sandboxIn.Close()
}

func (s *Sandbox) Logs() string {
	return s.sandboxErr.String()
}

func mountArgs(readable, writable []string) []string {
	seen := make(map[string]bool)
	var args []string
	// Writable first, in case a path exists in both
	for _, path := range writable {
		if seen[path] {
			continue
		}
		seen[path] = true
		args = append(args, "--writable", path)
	}
	for _, path := range readable {
		if seen[path] {
			continue
		}
		seen[path] = true
		args = append(args, "--readable", path)
	}
	return args
}
