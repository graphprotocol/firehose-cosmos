package filereader

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

var fileNameCheck *regexp.Regexp = regexp.MustCompile(`.*\.log\.\d*`)

type fileInfo struct {
	EOF      bool
	name     string
	position int64
}

type Reader struct {
	fileListName string
	path         string

	ctx                     context.Context
	fileMap                 map[string]fileInfo
	filesToRead             chan fs.DirEntry
	fileNameCheck           *regexp.Regexp
	lastFile                fileInfo
	log                     *zap.Logger
	maxDuration             time.Duration
	waitForNewFilesDuration time.Duration

	file *os.File
	lock *sync.RWMutex
}

func NewReader(ctx context.Context, l *zap.Logger, maxDuration, waitForNewFilesDuration time.Duration, fileNameRegexp, path string) (reader *Reader, err error) {
	reader = &Reader{
		fileListName:            fmt.Sprintf("%s/file_list.txt", path),
		path:                    path,
		ctx:                     ctx,
		fileMap:                 make(map[string]fileInfo),
		filesToRead:             make(chan fs.DirEntry, 100000),
		fileNameCheck:           regexp.MustCompile(fileNameRegexp),
		log:                     l,
		maxDuration:             maxDuration,
		waitForNewFilesDuration: waitForNewFilesDuration,
		lock:                    &sync.RWMutex{},
	}

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", path)
	}

	if reader.file, err = openFileOrCreateIfNotExists(reader.fileListName); err != nil {
		return nil, err
	}

	fileList, err := NewFileReader(maxDuration, reader.fileListName, 0)
	if err != nil {
		return nil, err
	}

	if _, err := fileList.ReadFile(reader.markFileAsRead, false); err != nil {
		return nil, err
	}

	return reader, nil
}

func (r *Reader) markFileAsRead(fileNameSizeAndPosition string) {
	strs := strings.Split(fileNameSizeAndPosition, ";")
	fileName := strs[0]
	position, err := strconv.Atoi(strs[1])
	if err != nil {
		r.log.Error("Could not read file size", zap.Error(err))
		os.Exit(1)
	}

	newFileInfo := fileInfo{
		EOF:      true,
		name:     fileName,
		position: int64(position),
	}

	fi, ok := r.fileMap[fileName]
	if !ok || ok && fi.position < newFileInfo.position {
		r.fileMap[fileName] = newFileInfo
	}

	r.lastFile = newFileInfo
}

func (r *Reader) Close() {
	if r.file != nil {
		r.file.Close()
	}

	r.ctx.Done()
	close(r.filesToRead)
}

func (r *Reader) StartSendingFilesInQueue(sendFunc SendFunc) error {
	for {
		select {
		case <-r.ctx.Done():
			return nil
		case file := <-r.filesToRead:
			defer r.log.Sync()

			if file == nil {
				return nil
			}

			fileName := fmt.Sprintf("%s/%s", r.path, file.Name())
			position := int64(0)
			if fi, ok := r.fileMap[fileName]; ok {
				position = fi.position
			}

			r.log.Debug("Start reading file", zap.String("path", fileName), zap.Int64("position", position))

			fileReader, err := NewFileReader(r.maxDuration, fileName, position)
			if err != nil {
				break
			}

			position, err = fileReader.ReadFile(sendFunc, true)
			if err != nil {
				break
			}

			if err = r.addFileNameToFileList(fileName, position); err != nil {
				break
			}

			fInfo := fileInfo{
				EOF:      true,
				name:     fileName,
				position: position,
			}

			r.fileMap[fileName] = fInfo

			r.lastFile = fInfo

			r.log.Debug("Finished reading file", zap.String("path", fileName), zap.Int64("position", position))

		default:
			r.updateFilesToRead()
		}
	}
}

func (r *Reader) addFileNameToFileList(fileName string, position int64) (err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.file == nil {
		if r.file, err = os.OpenFile(r.fileListName, os.O_RDWR, fs.ModeAppend); err != nil {
			return err
		}
	}

	if _, err := r.file.WriteString(fmt.Sprintf("%s;%d\n", fileName, position)); err != nil {
		return err
	}

	return nil
}

func (r *Reader) updateFilesToRead() {
	filesInfo, err := os.ReadDir(r.path)
	if err != nil {
		r.log.Error("Could not read directory info", zap.String("dir", r.path), zap.Error(err))
		os.Exit(1)
	}

	r.addNewFilesToReader(filesInfo)

	time.Sleep(r.waitForNewFilesDuration)
}

func (r *Reader) addNewFilesToReader(dirEntries []fs.DirEntry) {
	newFilesToRead := make([]fs.DirEntry, 0)

	for _, dirEntry := range dirEntries {
		fileName := dirEntry.Name()

		if !dirEntry.IsDir() && r.fileNameCheck.MatchString(fileName) {
			filePath := fmt.Sprintf("%s/%s", r.path, fileName)
			oldFileInfo, exists := r.fileMap[filePath]
			fi, err := dirEntry.Info()
			if err != nil {
				r.log.Error("Could not get file info", zap.String("path", filePath), zap.Error(err))
				os.Exit(1)
			}

			actualFileSize := fi.Size()
			if !exists || r.isLatestFileAndChangedSize(filePath, actualFileSize) {
				r.fileMap[filePath] = fileInfo{
					EOF:      false,
					name:     filePath,
					position: oldFileInfo.position,
				}
				newFilesToRead = append(newFilesToRead, dirEntry)
			}
		}
	}

	sort.Sort(sortFilesByName(newFilesToRead))

	for _, newFile := range newFilesToRead {
		r.filesToRead <- newFile
	}
}

func (r *Reader) isLatestFileAndChangedSize(fileName string, fileSize int64) bool {
	return r.lastFile.name != "" && r.lastFile.EOF && fileName == r.lastFile.name && r.lastFile.position < fileSize
}

func openFileOrCreateIfNotExists(fileName string) (file *os.File, err error) {
	file, err = os.OpenFile(fileName, os.O_RDWR, fs.ModeAppend)
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		file, err = os.Create(fileName)
	}

	if err != nil {
		return nil, err
	}

	return file, nil
}
