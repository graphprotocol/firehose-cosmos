package codec

import (
	"fmt"
	"io"
	"strings"

	"github.com/figment-networks/extractor-cosmos"
	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ConsoleReader struct {
	lines  chan string
	logger *zap.Logger
	done   chan interface{}

	height uint64
	block  *pbcosmos.Block
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
			return cr.block, nil
		case extractor.MsgBlock:
			cr.block = pl.Data.(*pbcosmos.Block)
		case extractor.MsgTx:
			cr.block.Transactions = append(cr.block.Transactions, pl.Data.(*pbcosmos.TxResult))
		case extractor.MsgValidatorSetUpdate:
			cr.block.ValidatorUpdates = pl.Data.(*pbcosmos.ValidatorSetUpdates).ValidatorUpdates
		}
	}

	cr.logger.Info("lines channel has been closed")

	return out, io.EOF
}

func (cr *ConsoleReader) startHeight(height uint64) error {
	if height <= cr.height {
		return fmt.Errorf("unexpected start height %d", height)
	}

	cr.height = height
	return nil
}

func FromProto(b interface{}) (*bstream.Block, error) {
	block, ok := b.(*pbcosmos.Block)
	if !ok {
		return nil, fmt.Errorf("unsupported type")
	}

	payload, err := proto.Marshal(block)
	if err != nil {
		return nil, err
	}

	blk := &bstream.Block{
		Id:             hex2string(block.Header.Hash),
		PreviousId:     hex2string(block.Header.LastBlockId.Hash),
		Number:         uint64(block.Header.Height),
		LibNum:         uint64(block.Header.Height - 1),
		Timestamp:      parseTimestamp(block.Header.Time),
		PayloadKind:    pbbstream.Protocol_COSMOS,
		PayloadVersion: 1,
	}

	if block.Header.Height == bstream.GetProtocolFirstStreamableBlock {
		blk.LibNum = bstream.GetProtocolFirstStreamableBlock
		blk.PreviousId = ""
	}

	return bstream.GetBlockPayloadSetter(blk, payload)
}

func hex2string(src []byte) string {
	return fmt.Sprintf("%X", src)
}
