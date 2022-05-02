package tools

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/figment-networks/firehose-cosmos/codec"
	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/spf13/cobra"
	"github.com/streamingfast/bstream"
	sftools "github.com/streamingfast/sf-tools"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func init() {
	Cmd.AddCommand(DownloadFromFirehoseCmd)
	DownloadFromFirehoseCmd.Flags().StringP("api-token-env-var", "a", "FIREHOSE_API_TOKEN", "Look for a JWT in this environment variable to authenticate against endpoint")
	DownloadFromFirehoseCmd.Flags().BoolP("plaintext", "p", false, "Use plaintext connection to firehose")
	DownloadFromFirehoseCmd.Flags().BoolP("insecure", "k", false, "Skip SSL certificate validation when connecting to firehose")
}

var DownloadFromFirehoseCmd = &cobra.Command{
	Use:     "download-from-firehose",
	Short:   "download blocks from firehose and save them to merged-blocks",
	Args:    cobra.ExactArgs(4),
	RunE:    downloadFromFirehoseE,
	PreRunE: initFirstStreamable,
	Example: "firehose-cosmos tools download-from-firehose f.q.d.n:443 1000 2000 ./outputdir",
}

func downloadFromFirehoseE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	endpoint := args[0]
	start, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("parsing start block num: %w", err)
	}
	stop, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("parsing stop block num: %w", err)
	}
	destFolder := args[3]

	apiTokenEnvVar := mustGetString(cmd, "api-token-env-var")
	apiToken := os.Getenv(apiTokenEnvVar)

	plaintext := mustGetBool(cmd, "plaintext")
	insecure := mustGetBool(cmd, "insecure")

	return sftools.DownloadFirehoseBlocks(
		ctx,
		endpoint,
		apiToken,
		insecure,
		plaintext,
		start,
		stop,
		destFolder,
		decodeAnyPB,
		zlog,
	)
}

func decodeAnyPB(in *anypb.Any) (*bstream.Block, error) {
	block := &pbcosmos.Block{}
	if err := anypb.UnmarshalTo(in, block, proto.UnmarshalOptions{}); err != nil {
		return nil, fmt.Errorf("unmarshal anypb: %w", err)
	}

	return codec.FromProto(block)
}
