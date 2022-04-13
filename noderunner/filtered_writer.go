package noderunner

import (
	"io"
	"regexp"
)

type FilteredWriter struct {
	re  *regexp.Regexp
	dst io.Writer
}

var (
	decolorizeRe = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
)

func NewFilteredWriter(dst io.Writer, expr string) (io.Writer, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	return FilteredWriter{re: re, dst: dst}, nil
}

func (w FilteredWriter) Write(data []byte) (int, error) {
	clean := decolorizeRe.ReplaceAll(data, []byte(""))
	if w.re.Match(clean) {
		return len(data), nil
	}

	return w.dst.Write(data)
}
