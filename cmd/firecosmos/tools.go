package main

import (
	"encoding/hex"
	"fmt"
	"io"

	pbcosmos "github.com/graphprotocol/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/streamingfast/bstream"
)

func printBlock(blk *bstream.Block, alsoPrintTransactions bool, out io.Writer) error {
	block := blk.ToProtocol().(*pbcosmos.Block)

	if _, err := fmt.Fprintf(out, "Block #%d (%s) (prev: %s): %d transactions\n",
		block.GetFirehoseBlockNumber(),
		block.GetFirehoseBlockID(),
		block.GetFirehoseBlockParentID()[0:7],
		len(block.Transactions),
	); err != nil {
		return err
	}

	if alsoPrintTransactions {
		for _, trx := range block.Transactions {
			if _, err := fmt.Fprintf(out, "  - Transaction %s\n", hex.EncodeToString(trx.Hash)); err != nil {
				return err
			}
		}
	}

	return nil
}
