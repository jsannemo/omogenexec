package eval

// An exitType describes why a program exited.
type exitType int

const (
	// exited means the program exited normally with an exit code.
	exited exitType = iota
	// signaled means the program was killed by a signal.
	signaled
	// timedOut means the program was killed due to exceeding its Time limit.
	timedOut
)

// An execResult describes the EvalResult of a single execution.
type execResult struct {
	// How the program exited.
	ExitType exitType
	// The exit code. Only set if the program exited with a code.
	ExitCode int
	// The termination signal. Only set if the program exited with a signal.
	Signal int
	// The Time the execution used.
	TimeUsageMs int64
	// The memory the execution used.
	MemoryUsageKb int
}

// CrashedWith checks whether the program exited normally with the given code.
func (res execResult) CrashedWith(code int) bool {
	return res.ExitType == exited && res.ExitCode == code
}

// Crashed checks whether the program exited abnormally.
func (res execResult) Crashed() bool {
	return (res.ExitType == exited && res.ExitCode != 0) || res.ExitType == signaled
}

// TimedOut checks whether the program exceeded its Time limit or not.
func (res execResult) TimedOut() bool {
	return res.ExitType == timedOut
}
