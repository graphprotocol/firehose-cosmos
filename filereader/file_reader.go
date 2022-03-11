package filereader

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var errMaxTimeDuration = errors.New("Max time duration for new changes exceeded")

type SendFunc func(string)

type FileReader struct {
	ctx                      context.Context
	fileName                 string
	fileSize                 int64
	position                 int64
	maxDurationForNewChanges time.Duration

	file   *os.File
	reader *bufio.Reader

	lock sync.Mutex
}

func NewFileReader(maxDuration time.Duration, fileName string, position int64) (*FileReader, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	// setting buffer for previous position
	if err = setBufferPosition(file, position); err != nil {
		return nil, err
	}

	fi, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}

	return &FileReader{
		ctx:                      context.Background(),
		fileName:                 fileName,
		fileSize:                 fi.Size(),
		position:                 position,
		maxDurationForNewChanges: maxDuration,
		file:                     file,
		reader:                   bufio.NewReader(file),
	}, nil
}

func (r *FileReader) Close() {
	if r.file != nil {
		r.file.Close()
	}

	r.ctx.Done()
}

func (r *FileReader) GetPosition() (offset int64, err error) {
	if r.file == nil {
		return 0, nil
	}

	offset, err = r.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, nil
	}

	if r.reader == nil {
		return 0, nil
	}

	offset -= int64(r.reader.Buffered())
	return
}

func (r *FileReader) ReadLine() (line string, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	prevPosition := r.position
	line, readerErr := r.reader.ReadString('\n')

	r.position, err = r.GetPosition()
	if err != nil {
		return "", err
	}

	if prevPosition == r.position {
		return "", io.EOF
	}

	if readerErr != nil {
		return line, readerErr
	}

	return strings.TrimRight(line, "\n"), nil
}

func (r *FileReader) WaitForChanges() (err error) {
	timeoutTime := time.Now().Add(r.maxDurationForNewChanges)

	for {
		select {
		case <-r.ctx.Done():
			return nil
		default:
			if time.Now().After(timeoutTime) {
				return errMaxTimeDuration
			}

			fi, err := os.Stat(r.fileName)
			if err != nil {
				return err
			}

			prevSize := r.fileSize
			if prevSize < fi.Size() {
				if err = r.Reopen(); err != nil {
					return err
				}

				return nil
			}
		}
	}
}

func (r *FileReader) Reopen() (err error) {
	if r.file, err = os.Open(r.fileName); err != nil {
		return err
	}
	r.reader = bufio.NewReader(r.file)

	fi, err := os.Stat(r.fileName)
	if err != nil {
		return err
	}
	r.fileSize = fi.Size()

	// setting buffer for previous position
	if err = setBufferPosition(r.file, r.position); err != nil {
		return err
	}

	return nil
}

func setBufferPosition(file *os.File, position int64) error {
	_, err := file.Seek(position, 0)
	return err
}

func (r *FileReader) ReadFile(sendFunc SendFunc, watch bool) (position int64, err error) {
	var line string

	defer r.Close()

	for {
		select {
		case <-r.ctx.Done():
			return 0, nil
		default:
			if r.fileSize == 0 {
				return 0, nil
			}

			line, err = r.ReadLine()

			if line != "" {
				sendFunc(line)
			}

			if err != nil && err == io.EOF {
				if watch {
					if err = r.WaitForChanges(); err != nil && err == errMaxTimeDuration {
						return r.position, nil
					}

				} else {
					return r.position, nil
				}

			}

			if err != nil {
				return r.position, err
			}
		}
	}
}
