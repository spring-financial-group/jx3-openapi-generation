package domain

type CommandRunner interface {
	// Execute executes the given command in the given directory returning the output and an error if any
	Execute(dir, name string, args ...string) (string, error)
	// ExecuteAndLog executes the given command in the given directory and logs the output if no error is returned
	ExecuteAndLog(dir, name string, args ...string) error
}

type CommandFailedError struct {
}

func (e *CommandFailedError) Error() string {
	return "stderr indicates command failed"
}
