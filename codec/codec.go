package codec

import (
	"errors"
	"fmt"
	"io"
	"time"

	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream"
	pbbstream "github.com/streamingfast/pbgo/sf/bstream/v1"
	"google.golang.org/protobuf/proto"
)

func init() {
	bstream.GetBlockReaderFactory = bstream.BlockReaderFactoryFunc(blockReaderFactory)
	bstream.GetBlockDecoder = bstream.BlockDecoderFunc(blockDecoder)
	bstream.GetBlockWriterFactory = bstream.BlockWriterFactoryFunc(blockWriterFactory)
	bstream.GetBlockWriterHeaderLen = 10
	bstream.GetBlockPayloadSetter = bstream.MemoryBlockPayloadSetter
	bstream.GetMemoizeMaxAge = 200 * 15 * time.Second

	// We want to panic in here to enforce validation in any component that uses this package,
	// instead of running validation in multiple places.
	if err := bstream.ValidateRegistry(); err != nil {
		panic(err)
	}
}

func Validate() error {
	if err := bstream.ValidateRegistry(); err != nil {
		return err
	}

	if bstream.GetProtocolFirstStreamableBlock == 0 {
		return errors.New("protocol first streamable block must be set")
	}

	return nil
}

// SetFirstStreamableBlock sets first block height available for streaming
func SetFirstStreamableBlock(height uint64) {
	bstream.GetProtocolFirstStreamableBlock = height
}

func blockReaderFactory(reader io.Reader) (bstream.BlockReader, error) {
	return NewBlockReader(reader)
}

func blockWriterFactory(writer io.Writer) (bstream.BlockWriter, error) {
	return NewBlockWriter(writer)
}

func blockDecoder(blk *bstream.Block) (interface{}, error) {
	if blk.Kind() != pbbstream.Protocol_COSMOS {
		return nil, fmt.Errorf("expected kind %s, got %s", pbbstream.Protocol_COSMOS, blk.Kind())
	}

	if blk.Version() != 1 {
		return nil, fmt.Errorf("this decoder only knows about version 1, got %d", blk.Version())
	}

	payload, err := blk.Payload.Get()
	if err != nil {
		return nil, fmt.Errorf("cant get block payload: %v", err)
	}

	sp := &pbcosmos.Block{}

	return sp, proto.Unmarshal(payload, sp)
}
