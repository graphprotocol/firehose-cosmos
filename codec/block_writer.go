package codec

import (
	"fmt"
	"io"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/dbin"
	"google.golang.org/protobuf/proto"
)

const (
	dbinContentType = "CSM"
)

type BlockWriter struct {
	src *dbin.Writer
}

func NewBlockWriter(writer io.Writer) (*BlockWriter, error) {
	dbinWriter := dbin.NewWriter(writer)

	err := dbinWriter.WriteHeader(dbinContentType, 1)
	if err != nil {
		return nil, fmt.Errorf("unable to write file header: %s", err)
	}

	return &BlockWriter{
		src: dbinWriter,
	}, nil
}

func (w *BlockWriter) Write(block *bstream.Block) error {
	pbBlock, err := block.ToProto()
	if err != nil {
		return err
	}

	bytes, err := proto.Marshal(pbBlock)
	if err != nil {
		return fmt.Errorf("unable to marshal proto block: %s", err)
	}

	return w.src.WriteMessage(bytes)
}
