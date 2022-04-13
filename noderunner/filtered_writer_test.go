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
		{
			expr: "deepmind",
			input: []string{
				"\x1b[90m1:01PM\x1b[0m \x1b[32mINF\x1b[0m Version info \x1b[36mblock=\x1b[0m11 \x1b[36mp2p=\x1b[0m8 \x1b[36msoftware=\x1b[0mv0.34.9-deepmind\n",
			},
			output: "",
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
