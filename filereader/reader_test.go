package filereader

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const CLOSE_READER_DURATION time.Duration = 500 * time.Millisecond

var fileCount int = 0

type ReaderTest struct {
	suite.Suite

	dirName string
	lines   string

	ctx    context.Context
	files  []*os.File
	lock   sync.Mutex
	reader *Reader
}

func (t *ReaderTest) SetupTest() {
	var err error

	t.ctx = context.Background()
	t.lines = ""

	t.dirName, err = ioutil.TempDir("", "test_dir")
	t.Require().Nil(err)

	t.addFile("block 1\nblock 2\n")

	lLev := zap.NewAtomicLevelAt(zap.DebugLevel)
	logConfig := zap.Config{
		OutputPaths: []string{"stderr"},
		Encoding:    "json",
		Level:       lLev,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "level",
			TimeKey:        "time",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
	}

	log, err := logConfig.Build()
	t.Require().Nil(err)

	t.reader, err = NewReader(t.ctx, log, 300*time.Microsecond, 200*time.Millisecond, ".*\\.log\\.\\d*", t.dirName)
	t.Require().Nil(err)

	t.closeReader()
}

func (t *ReaderTest) addFile(str string) {
	file, err := ioutil.TempFile(t.dirName, fmt.Sprintf("file.log.%d", fileCount))
	t.Require().Nil(err)
	fileCount++

	_, err = file.WriteString(str)
	t.files = append(t.files, file)
}

func (t *ReaderTest) addLineToFile(line string, i int) {
	_, err := t.files[i].WriteString(line)
	t.Require().Nil(err)
}

func (t *ReaderTest) readLine(line string) {
	t.lines += line
}

func (t *ReaderTest) AfterTest() {
	for _, file := range t.files {
		t.Require().Nil(file.Close())
		fmt.Println("file name: ", file.Name())
		t.Require().Nil(os.Remove(file.Name()))
	}

	t.Require().Nil(os.RemoveAll(t.dirName))
	fmt.Println("after test lines empty")
}

func (t *ReaderTest) closeReader() {
	go func() {
		time.Sleep(CLOSE_READER_DURATION)
		t.reader.Close()
	}()
}

func (t *ReaderTest) Test_Read_TwoFiles_OK() {
	t.addLineToFile("block 3\n", 0)
	t.addLineToFile("block 4\n", 0)

	t.addFile("BLOCK 1\nBLOCK 2\nBLOCK 3\n")
	t.addLineToFile("BLOCK 4\n", 1)

	t.Require().Nil(t.reader.StartSendingFilesInQueue(t.readLine))
	t.Equal("block 1block 2block 3block 4BLOCK 1BLOCK 2BLOCK 3BLOCK 4", t.lines)
}

func (t *ReaderTest) Test_Terminates_On_Context_Cancel() {
	now := time.Now().Unix()
	closeReaderTime := now + CLOSE_READER_DURATION.Milliseconds()
	t.ctx.Done()

	t.Require().Nil(t.reader.StartSendingFilesInQueue(t.readLine))

	t.Require().Condition(func() (success bool) {
		currentTime := time.Now().Unix()
		return currentTime < closeReaderTime

	}, fmt.Sprintf("context should close reader before timeout (close reader duration in ms: %d)", CLOSE_READER_DURATION))

	t.Equal("block 1block 2", t.lines)
}

func TestReader(t *testing.T) {
	suite.Run(t, new(ReaderTest))
}
