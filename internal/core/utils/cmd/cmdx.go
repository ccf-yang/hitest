package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"hitest/internal/core/app/task"
	"hitest/internal/core/buserr"
	"hitest/internal/shared/constant"
)

const maxStreamOutputCapture = 64 * 1024

type CommandHelper struct {
	context            context.Context
	workDir            string
	outputFile         string
	env                []string
	timeout            time.Duration
	taskItem           *task.Task
	logger             *log.Logger
	IgnoreExist1       bool
	preserveErrorCause bool
}

type Option func(*CommandHelper)

type PipeCommand struct {
	Name  string
	Args  []string
	Env   []string
	Dir   string
	Stdin io.Reader
}

type lockedBuffer struct {
	mu        sync.Mutex
	buf       bytes.Buffer
	limit     int
	truncated int
}

func NewCommandMgr(opts ...Option) *CommandHelper {
	s := &CommandHelper{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (c *CommandHelper) Run(name string, arg ...string) error {
	_, err := c.run(name, arg...)
	return err
}

func (c *CommandHelper) RunWithStdout(name string, arg ...string) (string, error) {
	return c.run(name, arg...)
}

func (c *CommandHelper) pipeResultErr(ctx context.Context, runErr error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return buserr.New("ErrCmdTimeout")
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return buserr.New("ErrShutDown")
	}
	return runErr
}

func (c *CommandHelper) run(name string, arg ...string) (string, error) {
	var cmd *exec.Cmd
	var newContext context.Context
	var cancel context.CancelFunc
	var outputFile *os.File

	if c.timeout != 0 {
		if c.context == nil {
			newContext, cancel = context.WithTimeout(context.Background(), c.timeout)
		} else {
			newContext, cancel = context.WithTimeout(c.context, c.timeout)
		}
		defer cancel()
	} else if c.context != nil {
		newContext = c.context
	}

	if newContext != nil {
		cmd = exec.CommandContext(newContext, name, arg...)
	} else {
		cmd = exec.Command(name, arg...)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	if newContext != nil {
		cmd.Cancel = func() error {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			return cmd.Process.Kill()
		}
	}

	customWriter := &CustomWriter{taskItem: c.taskItem}
	var stdout, stderr bytes.Buffer
	var loggerCloser io.Closer
	if c.taskItem != nil {
		cmd.Stdout = io.MultiWriter(&stdout, customWriter)
		cmd.Stderr = io.MultiWriter(&stderr, customWriter)
	} else if c.logger != nil {
		streamWriter := c.logger.Writer()
		if closer, ok := streamWriter.(io.Closer); ok {
			loggerCloser = closer
		}
		cmd.Stdout = io.MultiWriter(&stdout, streamWriter)
		cmd.Stderr = io.MultiWriter(&stderr, streamWriter)
	} else if len(c.outputFile) != 0 {
		file, err := os.OpenFile(c.outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
		if err != nil {
			return "", err
		}
		outputFile = file
		cmd.Stdout = io.MultiWriter(&stdout, outputFile)
		cmd.Stderr = io.MultiWriter(&stderr, outputFile)
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}
	cmd.Env = append(os.Environ(), c.env...)
	if len(c.workDir) != 0 {
		cmd.Dir = c.workDir
	}
	defer func() {
		if loggerCloser != nil {
			_ = loggerCloser.Close()
		}
		if outputFile != nil {
			_ = outputFile.Close()
		}
	}()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("cmd start failed: %w", err)
	}
	if c.taskItem != nil {
		defer customWriter.Flush()
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case err := <-done:
		if newContext != nil && newContext.Err() != nil {
			resultErr := c.pipeResultErr(newContext, err)
			if c.preserveErrorCause {
				resultErr = &commandError{message: resultErr.Error(), cause: newContext.Err()}
			}
			return "", resultErr
		}
		if err != nil {
			out, resultErr := handleErr(&stdout, &stderr, c.IgnoreExist1, err)
			if c.preserveErrorCause && resultErr != nil {
				resultErr = &commandError{message: resultErr.Error(), cause: err}
			}
			return out, resultErr
		}
		return stdout.String(), nil
	case <-contextDone(newContext):
		if cmd.Process != nil && cmd.Process.Pid > 0 {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		var err error
		switch newContext.Err() {
		case context.DeadlineExceeded:
			err = buserr.New("ErrCmdTimeout")
		case context.Canceled:
			err = buserr.New("ErrShutDown")
		default:
			err = newContext.Err()
		}
		<-done
		if c.preserveErrorCause {
			err = &commandError{message: err.Error(), cause: newContext.Err()}
		}
		return "", err
	}
}

func contextDone(ctx context.Context) <-chan struct{} {
	if ctx == nil {
		return nil
	}
	return ctx.Done()
}

type commandError struct {
	message string
	cause   error
}

func (e *commandError) Error() string { return e.message }
func (e *commandError) Unwrap() error { return e.cause }

type CustomWriter struct {
	mu       sync.Mutex
	taskItem *task.Task
	buffer   bytes.Buffer
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.buffer.Write(p)
	lines := strings.Split(cw.buffer.String(), "\n")

	for i := 0; i < len(lines)-1; i++ {
		cw.taskItem.Log(lines[i])
	}
	cw.buffer.Reset()
	cw.buffer.WriteString(lines[len(lines)-1])

	return len(p), nil
}

func (cw *CustomWriter) Flush() {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	if cw.buffer.Len() > 0 {
		cw.taskItem.Log(cw.buffer.String())
		cw.buffer.Reset()
	}
}

func handleErr(stdout, stderr fmt.Stringer, ignoreExist1 bool, err error) (string, error) {
	return handleErrString(stdout.String(), stderr.String(), ignoreExist1, err)
}

func handleErrString(stdout, stderr string, ignoreExist1 bool, err error) (string, error) {
	var exitError *exec.ExitError
	if ignoreExist1 && errors.As(err, &exitError) {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if status.ExitStatus() == 1 {
				return "", nil
			}
		}
	}
	outItem := stdout
	errItem := stderr
	if len(errItem) != 0 && len(outItem) != 0 {
		return outItem, fmt.Errorf("stdout: %s; stderr: %s, err: %v", outItem, errItem, err)
	}
	if len(errItem) != 0 {
		return outItem, fmt.Errorf("stderr: %s, err: %v", errItem, err)
	}
	if len(outItem) != 0 {
		return outItem, fmt.Errorf("stdout: %s, err: %v", outItem, err)
	}
	return "", err
}
