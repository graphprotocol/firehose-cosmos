package noderunner

import (
	"bufio"
	"io"
	"strings"

	"go.uber.org/zap"
)

func StartLineReader(input io.Reader, readerFunc func(string), logger *zap.Logger) error {
	logger.Info("starting line reader")
	reader := bufio.NewReaderSize(input, defaultBufferSize)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// Abnormal termination
			if err != io.EOF {
				logger.Debug("line reader aborted with error", zap.Error(err))
				return err
			}

			// We're done reading content
			if len(line) == 0 {
				logger.Debug("line reader finished reading input")
				return nil
			}
		}

		readerFunc(strings.TrimSpace(line))
	}
}
