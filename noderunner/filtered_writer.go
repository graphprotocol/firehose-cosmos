package noderunner

import (
	"io"
	"regexp"
)

type FilteredWriter struct {
	Writer io.Writer
}

var (
	// Regular expression to filter out unwanted logs
	// NOTE: these wont work when node process runs without "--log_format=json" CLI arg
	ignoreRegex = regexp.MustCompile(`"module":"(p2p|pex|consensus|x\/bank)"`)
)

func (w FilteredWriter) Write(data []byte) (int, error) {
	if ignoreRegex.Match(data) {
		return len(data), nil
	}
	return w.Writer.Write(data)
}
