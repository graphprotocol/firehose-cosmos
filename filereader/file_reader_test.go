package filereader

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type FileReaderTest struct {
	suite.Suite

	count   int
	dirName string
	lines   string

	files []*os.File
	lock  sync.Mutex
}

func (t *FileReaderTest) SetupTest() {
	var err error

	t.dirName, err = ioutil.TempDir("", "test_dir")
	t.Require().Nil(err)

	t.lines = ""
}

func (t *FileReaderTest) addFile(str string) (fileName string) {
	fileCount := t.count
	t.count++

	file, err := ioutil.TempFile(t.dirName, fmt.Sprintf("file-%d", fileCount))
	t.Require().Nil(err)

	_, err = file.WriteString(str)
	t.files = append(t.files, file)

	return file.Name()
}

func (t *FileReaderTest) addLineToFileByName(fileName, line string) {
	for i, file := range t.files {
		if file.Name() == fileName {
			_, err := t.files[i].WriteString(line)
			t.Require().Nil(err)
		}
	}
}

func (t *FileReaderTest) readLine(line string) {
	t.lines += line
}

func (t *FileReaderTest) AfterTest() {
	for _, file := range t.files {
		t.Require().Nil(file.Close())
	}

	t.Require().Nil(os.RemoveAll(t.dirName))
}

func (t *FileReaderTest) TestReadFile_WithoutWatch_OK() {
	fileName := t.addFile("block 1\nblock 2\n")
	fileReader, err := NewFileReader(time.Second, fileName, 0)
	t.Require().Nil(err)
	defer fileReader.Close()

	position, err := fileReader.ReadFile(t.readLine, true)
	t.Require().Nil(err)

	t.Equal(int64(16), position)
	t.Equal("block 1block 2", t.lines)
}

func (t *FileReaderTest) TestReadFile_WithWatch_OK() {
	fileName := t.addFile("block 1\nblock 2\n")
	go t.addLineToFileByName(fileName, "block 3\n")
	go t.addLineToFileByName(fileName, "block 4\n")

	fileReader, err := NewFileReader(time.Second, fileName, 0)
	t.Require().Nil(err)
	defer fileReader.Close()

	position, err := fileReader.ReadFile(t.readLine, true)
	t.Require().Nil(err)

	t.Equal(int64(32), position)
	t.Equal("block 1block 2block 3block 4", t.lines)
}

func TestFileReader(t *testing.T) {
	suite.Run(t, new(FileReaderTest))
}
