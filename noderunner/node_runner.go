package noderunner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const (
	defaultKillTimeout = 10 * time.Second
	defaultBufferSize  = 10 * 1024 * 1024
)

type NodeRunner struct {
	bin               string
	dir               string
	args              []string
	env               map[string]string
	stderr            bool
	lineReaderFunc    func(string)
	bufferSize        int
	forcedKillTimeout time.Duration
	logger            *zap.Logger
	done              chan struct{}
	logFilter         bool
	logFilterExpr     string
}

func New(bin string, args []string, stderr bool) *NodeRunner {
	return &NodeRunner{
		bin:               bin,
		args:              args,
		stderr:            stderr,
		forcedKillTimeout: defaultKillTimeout,
		bufferSize:        defaultBufferSize,
		done:              make(chan struct{}),
		logger:            zap.NewNop(),
	}
}

func (runner *NodeRunner) Done() <-chan struct{} {
	return runner.done
}

func (runner *NodeRunner) SetEnv(env map[string]string) {
	runner.env = env
}

func (runner *NodeRunner) SetLogger(logger *zap.Logger) {
	runner.logger = logger
}

func (runner *NodeRunner) SetLineReader(fn func(string)) {
	runner.lineReaderFunc = fn
}

func (runner *NodeRunner) SetDir(dir string) {
	runner.dir = dir
}

func (runner *NodeRunner) SetLogFiltering(expr string) {
	if expr != "" {
		runner.logFilterExpr = expr
	}
}

func (runner *NodeRunner) Start(ctx context.Context) error {
	if runner.bin == "" {
		return errors.New("binary path is not provided")
	}

	runner.logger.Info("starting subprocess", zap.Any("bin", runner.bin), zap.Any("args", runner.args))
	return runner.startProcess(ctx)
}

func (runner *NodeRunner) startProcess(ctx context.Context) error {
	cmd := exec.Command(runner.bin, runner.args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if len(runner.env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range runner.env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if runner.dir != "" {
		cmd.Dir = runner.dir
	}

	if runner.stderr {
		cmd.Stderr = os.Stderr

		if runner.logFilterExpr != "" {
			filterWriter, err := NewFilteredWriter(cmd.Stderr, runner.logFilterExpr)
			if err != nil {
				return err
			}
			cmd.Stderr = filterWriter
		}
	}

	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		runner.logger.Debug("cant initialize stdout pipe", zap.Error(err))
		return err
	}
	defer cmdStdout.Close()

	runner.logger.Debug("starting the subprocess")
	if err := cmd.Start(); err != nil {
		runner.logger.Debug("cant start the process", zap.Error(err))
		return err
	}
	runner.logger.Debug("subprocess started", zap.Any("pid", cmd.Process.Pid))

	// Start line reader in the background
	readerDone := make(chan error)
	go func() {
		readerDone <- runner.startLineReader(cmdStdout)
	}()

	// Wait for command execution.
	// If command context is cancelled, first try to gracefully stop the process.
	// Forceful termination will kick in after the timeout.
	err = runner.waitWithTimeout(ctx, cmd, runner.forcedKillTimeout)
	runner.logger.Debug("runner finished", zap.Error(err))

	// We need to wait until reader can process all lines after the subprocess
	// has been terminated. Returning prematurely will cause the contents to be lost.
	if readerErr := <-readerDone; readerErr != nil {
		runner.logger.Debug("reader finished with error", zap.Error(readerErr))
	}

	return err
}

func (runner *NodeRunner) startLineReader(input io.Reader) error {
	return StartLineReader(input, runner.lineReaderFunc, runner.logger)
}

func (runner *NodeRunner) waitWithTimeout(cmdCtx context.Context, cmd *exec.Cmd, waitTimeout time.Duration) error {
	var (
		killTimer     *time.Timer
		killTimerLock sync.Mutex
	)

	defer func() {
		killTimerLock.Lock()
		if killTimer != nil {
			killTimer.Stop()
		}
		killTimerLock.Unlock()
	}()

	go func() {
		<-cmdCtx.Done()
		runner.logger.Debug("runner context is cancelled")

		// Process is gone, so we can't really do much at this point.
		// Wait until cmd.Wait populates errChan.
		if cmd.Process == nil {
			return
		}

		// Send SIGINT to allow process to terminate gracefully.
		runner.logger.Debug("gracefully stopping the subprocess", zap.Any("pid", cmd.Process.Pid))
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			runner.logger.Debug("cant send INT signal to process", zap.Any("pid", cmd.Process.Pid), zap.Error(err))
			if errors.Is(err, os.ErrProcessDone) {
				return
			}
		}

		// Start the forceful termination timer.
		killTimerLock.Lock()
		killTimer = time.AfterFunc(waitTimeout, func() {
			if cmd != nil && cmd.Process != nil {
				runner.logger.Debug("forcefully stopping the subprocess", zap.Any("pid", cmd.Process.Pid))
				if err := cmd.Process.Kill(); err != nil {
					runner.logger.Debug("cant kill the process", zap.Any("pid", cmd.Process.Pid), zap.Error(err))
				}
			}
		})
		killTimerLock.Unlock()
	}()

	return cmd.Wait()
}
