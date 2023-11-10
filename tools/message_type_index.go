package tools

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/graphprotocol/firehose-cosmos/transform"
	pbcosmos "github.com/graphprotocol/proto-cosmos/pb/sf/cosmos/type/v1"

	"github.com/spf13/cobra"
	"github.com/streamingfast/bstream"
	"github.com/streamingfast/bstream/stream"
	bstransform "github.com/streamingfast/bstream/transform"
	"github.com/streamingfast/dstore"
	"github.com/streamingfast/firehose"
	pbfirehose "github.com/streamingfast/pbgo/sf/firehose/v2"
	"go.uber.org/zap"
)

var generateMessageTypeIdxCmd = &cobra.Command{
	Use:   "generate-message-type-index {index-url} {irr-index-url} {source-blocks-url} {start-block-num} {stop-block-num}",
	Short: "Generate index files for message types present in blocks",
	Args:  cobra.RangeArgs(4, 5),
	RunE:  generateMessageTypeIdxE,
}

func init() {
	generateMessageTypeIdxCmd.Flags().Uint64("first-streamable-block", 0, "first streamable block of this chain")
	generateMessageTypeIdxCmd.Flags().Uint64("indexes-size", 10000, "size of index bundles that will be created")
	generateMessageTypeIdxCmd.Flags().IntSlice("lookup-indexes-sizes", []int{1000000, 100000, 10000, 1000}, "index bundle sizes that we will look for on start to find first unindexed block (should include indexes-size)")
	Cmd.AddCommand(generateMessageTypeIdxCmd)
}

func generateMessageTypeIdxE(cmd *cobra.Command, args []string) error {
	var err error
	bstream.GetProtocolFirstStreamableBlock, err = cmd.Flags().GetUint64("first-streamable-block")
	if err != nil {
		return err
	}
	idxSize, err := cmd.Flags().GetUint64("indexes-size")
	if err != nil {
		return err
	}
	lais, err := cmd.Flags().GetIntSlice("lookup-indexes-sizes")
	if err != nil {
		return err
	}
	var lookupIdxSizes []uint64
	for _, size := range lais {
		if size < 0 {
			return fmt.Errorf("invalid negative size for bundle-sizes: %d", size)
		}
		lookupIdxSizes = append(lookupIdxSizes, uint64(size))
	}

	indexStoreURL := args[0]
	blocksStoreURL := args[1]
	startBlockNum, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse block number %q: %w", args[0], err)
	}
	var stopBlockNum uint64
	if len(args) == 5 {
		stopBlockNum, err = strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse block number %q: %w", args[0], err)
		}
	}

	mergedBlocksStore, err := dstore.NewDBinStore(blocksStoreURL)
	if err != nil {
		return fmt.Errorf("failed setting up block store from url %q: %w", blocksStoreURL, err)
	}

	indexStore, err := dstore.NewStore(indexStoreURL, "", "", false)
	if err != nil {
		return fmt.Errorf("failed setting up an index store from url %q: %w", indexStoreURL, err)
	}

	streamFactory := firehose.NewStreamFactory(
		mergedBlocksStore,
		nil,
		nil,
		nil,
	)
	cmd.SilenceUsage = true

	ctx := context.Background()

	startBlockNum = bstransform.FindNextUnindexed(ctx, uint64(startBlockNum), lookupIdxSizes, transform.EventOriginIndexShortName, indexStore)

	zlog.Info("resolved next unindexed regions", zap.Uint64("resolved_start", startBlockNum))

	t := transform.NewMessageTypeIndexer(indexStore, idxSize, startBlockNum)

	handler := bstream.HandlerFunc(func(blk *bstream.Block, obj interface{}) error {
		t.ProcessBlock(blk.ToProtocol().(*pbcosmos.Block))
		return nil
	})

	req := &pbfirehose.Request{
		StartBlockNum:   int64(startBlockNum),
		StopBlockNum:    stopBlockNum,
		FinalBlocksOnly: true,
	}

	s, err := streamFactory.New(
		ctx,
		handler,
		req,
		true,
		zlog,
	)
	if err != nil {
		return fmt.Errorf("getting firehose stream: %w", err)
	}

	if err := s.Run(ctx); err != nil {
		if !errors.Is(err, stream.ErrStopBlockReached) {
			return err
		}
	}
	zlog.Info("complete")
	return nil
}
