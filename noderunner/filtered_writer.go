package noderunner

import (
	"io"
	"regexp"
)

type FilteredWriter struct {
	re  *regexp.Regexp
	dst io.Writer
}

const (
	// Regular expression to filter out unwanted logs
	// NOTE: these wont work when node process runs without "--log_format=json" CLI arg
	exampleFilterExpr = `"module":"(p2p|pex|consensus|x\/bank)"`
)

func NewFilteredWriter(dst io.Writer, expr string) (io.Writer, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	return FilteredWriter{re: re, dst: dst}, nil
}

func (w FilteredWriter) Write(data []byte) (int, error) {
	if w.re.Match(data) {
		return len(data), nil
	}
	return w.dst.Write(data)
}
