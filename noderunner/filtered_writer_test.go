package noderunner

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilteredWriter(t *testing.T) {
	examples := []struct {
		expr   string
		input  []string
		output string
	}{
		{
			expr: "foo",
			input: []string{
				"foo",
				"bar",
			},
			output: "bar",
		},
		{
			expr: `\smodule=(p2p|pex|consensus|x\/bank)`,
			input: []string{
				`ERR error while stopping peer error="already stopped" module=p2p`,
				`INF Stopping BlockPool service impl={"Logger":{}} module=pex`,
				`Doing something`,
				`ERR error while stopping peer error="already stopped" module=consensus`,
			},
			output: "Doing something",
		},
	}

	for _, ex := range examples {
		out := bytes.NewBuffer(nil)

		writer, err := NewFilteredWriter(out, ex.expr)
		assert.NoError(t, err)

		for _, line := range ex.input {
			_, err := writer.Write([]byte(line))
			assert.NoError(t, err)
		}

		assert.Equal(t, ex.output, out.String())
	}
}
