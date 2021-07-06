package main

// An exitType describes why a program exited.
type exitType int

const (
	// Exited means the program exited normally with an exit code.
	Exited exitType = iota
	// Signaled means the program was killed by a signal.
	Signaled
	// TimedOut means the program was killed due to exceeding its Time limit.
	TimedOut
)

// An ExecResult describes the EvalResult of a single execution.
type ExecResult struct {
	// How how the program exited.
	ExitType exitType
	// The exit code. Only set if the program exited with a code.
	ExitCode int
	// The termination signal. Only set if the program exited with a signal.
	Signal int
	// The Time the execution used.
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

// TimedOut checks whether the program exceeded its Time limit or not.
func (res ExecResult) TimedOut() bool {
	return res.ExitType == TimedOut
}
