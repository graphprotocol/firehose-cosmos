package noderunner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeRunner(t *testing.T) {
	// t.Run("basic", func(t *testing.T) {
	// 	runner := New("true", []string{}, true)

	// 	err := runner.Start(context.Background())
	// 	assert.Nil(t, err)
	// })

	// t.Run("invalid command", func(t *testing.T) {
	// 	runner := New("foo", []string{}, true)

	// 	err := runner.Start(context.Background())
	// 	assert.Error(t, err)
	// 	assert.Equal(t, `exec: "foo": executable file not found in $PATH`, err.Error())
	// })

	// t.Run("with cancel", func(t *testing.T) {
	// 	runner := New("sleep", []string{"10"}, true)

	// 	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	// 	defer cancel()

	// 	err := runner.Start(ctx)
	// 	assert.Error(t, err)
	// 	assert.Equal(t, "signal: interrupt", err.Error())
	// })

	t.Run("line reader", func(t *testing.T) {
		lines := []string{}

		runner := New("echo", []string{"hello", "world"}, true)
		runner.SetLineReader(func(line string) {
			lines = append(lines, line)
		})

		err := runner.Start(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []string{"hello world"}, lines)
	})
}
