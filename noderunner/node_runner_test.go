package noderunner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNodeRunner(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("basic", func(t *testing.T) {
		runner := New("true", []string{}, true)
		runner.SetLogger(logger)

		err := runner.Start(context.Background())
		assert.Nil(t, err)
	})

	t.Run("invalid command", func(t *testing.T) {
		runner := New("foo", []string{}, true)
		runner.SetLogger(logger)

		err := runner.Start(context.Background())
		assert.Error(t, err)
		assert.Equal(t, `exec: "foo": executable file not found in $PATH`, err.Error())
	})

	t.Run("with cancel", func(t *testing.T) {
		runner := New("sleep", []string{"10"}, true)
		runner.SetLogger(logger)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := runner.Start(ctx)
		assert.Error(t, err)
		assert.Equal(t, "signal: interrupt", err.Error())
	})

	t.Run("line reader", func(t *testing.T) {
		lines := []string{}

		runner := New("../test/script.sh", []string{}, true)
		runner.SetLogger(logger)

		runner.SetLineReader(func(line string) {
			lines = append(lines, line)
		})

		err := runner.Start(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []string{"Line 1", "Line 2", "Line 3"}, lines)
	})
}
