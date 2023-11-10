package codec

import (
	"fmt"
	"io"
	"strings"

	"github.com/graphprotocol/extractor-cosmos"
	pbcosmos "github.com/graphprotocol/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream"
	firecore "github.com/streamingfast/firehose-core"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/node-manager/mindreader"
	"go.uber.org/zap"
)

type ConsoleReader struct {
	lines   chan string
	logger  *zap.Logger
	done    chan interface{}
	encoder firecore.BlockEncoder

	height uint64
	block  *pbcosmos.Block
}

func NewConsoleReader(lines chan string, blockEncoder firecore.BlockEncoder, logger *zap.Logger, tracer logging.Tracer) (mindreader.ConsolerReader, error) {
	return &ConsoleReader{
		lines:   lines,
		logger:  logger,
		done:    make(chan interface{}),
		encoder: blockEncoder,
	}, nil
}

func (cr *ConsoleReader) Done() <-chan interface{} {
	return cr.done
}

func (cr *ConsoleReader) Close() {}

func (cr *ConsoleReader) ReadBlock() (out *bstream.Block, err error) {
	v, err := cr.next()
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, fmt.Errorf("console reader read a nil *bstream.Block, this is invalid")
	}

	pbBlock := v.(*pbcosmos.Block)
	blk, err := cr.encoder.Encode(pbBlock)
	if err != nil {
		return nil, err
	}
	return blk, nil

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
