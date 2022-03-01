package codec

import (
	"fmt"
	"io"
	"strings"

	"github.com/figment-networks/extractor-tendermint"
	"github.com/figment-networks/tendermint-protobuf-def/codec"

	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ConsoleReader struct {
	lines  chan string
	logger *zap.Logger
	done   chan interface{}

	height    uint64
	lastFrame *codec.EventList
}

func NewConsoleReader(lines chan string, logger *zap.Logger) (*ConsoleReader, error) {
	return &ConsoleReader{
		lines:  lines,
		logger: logger,
		done:   make(chan interface{}),
	}, nil
}

func (cr *ConsoleReader) Done() <-chan interface{} {
	return cr.done
}

func (cr *ConsoleReader) Close() {}

func (cr *ConsoleReader) Read() (out interface{}, err error) {
	return cr.next()
}

func (cr *ConsoleReader) next() (out interface{}, err error) {
	for line := range cr.lines {
		pl, err := parseLine(strings.TrimSpace(line))
		if err != nil {
			return nil, fmt.Errorf("%s (line %q)", err, line)
		}
		if pl == nil {
			continue
		}

		switch pl.Kind {
		case extractor.MsgBegin:
			if err := cr.startHeight(pl.Data.(uint64)); err != nil {
				return nil, err
			}
		case extractor.MsgEnd:
			height := pl.Data.(uint64)
			if cr.height != height {
				return nil, fmt.Errorf("unexpected end height end: %d", height)
			}
			return cr.lastFrame, nil
		case extractor.MsgTx:
			cr.lastFrame.Transaction = append(cr.lastFrame.Transaction, pl.Data.(*codec.EventTx))
		case extractor.MsgBlock:
			cr.lastFrame = initializeNewFrame(pl.Data.(*codec.EventBlock))
		case extractor.MsgValidatorSetUpdate:
			cr.lastFrame.ValidatorSetUpdates = pl.Data.(*codec.EventValidatorSetUpdates)
		}
	}

	cr.logger.Info("lines channel has been closed")

	return out, io.EOF
}

func (cr *ConsoleReader) startHeight(height uint64) error {
	if height <= cr.height {
		return fmt.Errorf("unexpected height %d", height)
	}

	cr.height = height
	return nil
}

func initializeNewFrame(nblock *codec.EventBlock) *codec.EventList {
	return &codec.EventList{
		NewBlock:    nblock,
		Transaction: []*codec.EventTx{},
	}
}

func FromProto(b interface{}) (*bstream.Block, error) {
	eventList, ok := b.(*codec.EventList)
	if !ok {
		return nil, fmt.Errorf("unsupported type")
	}

	payload, err := proto.Marshal(eventList)
	if err != nil {
		return nil, err
	}

	header := eventList.NewBlock.Block.Header

	block := &bstream.Block{
		Id:             hex2string(eventList.NewBlock.BlockId.Hash),
		PreviousId:     hex2string(header.LastBlockId.Hash),
		Number:         uint64(header.Height),
		LibNum:         uint64(header.Height - 1),
		Timestamp:      parseTimestamp(header.Time),
		PayloadKind:    pbbstream.Protocol_TENDERMINT,
		PayloadVersion: 1,
	}

	if header.Height == bstream.GetProtocolFirstStreamableBlock {
		block.LibNum = bstream.GetProtocolFirstStreamableBlock
		block.PreviousId = ""
	}

	return bstream.GetBlockPayloadSetter(block, payload)
}

func hex2string(src []byte) string {
	return fmt.Sprintf("%X", src)
}
