package noderunner

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestStartLineReader(t *testing.T) {
	examples := []struct {
		input  func() io.Reader
		output []string
		err    error
	}{
		{
			input:  func() io.Reader { return strings.NewReader("") },
			output: []string{},
		},
		{
			input:  func() io.Reader { return strings.NewReader("line1\nline2\nline3") },
			output: []string{"line1", "line2", "line3"},
		},
		{
			input:  func() io.Reader { return strings.NewReader("line       \n") },
			output: []string{"line"},
		},
		{
			input: func() io.Reader {
				return testReader{err: errors.New("boom")}
			},
			output: []string{},
			err:    errors.New("boom"),
		},
	}

	for _, ex := range examples {
		output := []string{}
		handler := func(line string) { output = append(output, line) }

		err := StartLineReader(ex.input(), handler, zap.NewNop())

		assert.Equal(t, ex.err, err)
		assert.Equal(t, ex.output, output)
	}
}

type testReader struct {
	err error
}

func (r testReader) Read(buf []byte) (int, error) {
	return 0, r.err
}
