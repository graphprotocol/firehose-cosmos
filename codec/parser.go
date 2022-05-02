package codec

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/figment-networks/extractor-cosmos"
	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"

	"google.golang.org/protobuf/proto"
)

const (
	dmLogPrefix = "DMLOG "
)

var (
	errInvalidFormat   = errors.New("invalid format")
	errInvalidData     = errors.New("invalid data")
	errUnsupportedKind = errors.New("unsupported kind")
)

type ParsedLine struct {
	Kind string
	Data interface{}
}

// DMLOG BEGIN <HEIGHT>
// DMLOG BLOCK <DATA>
// DMLOG TX <DATA>
// DMLOG END <HEIGHT>
func parseLine(line string) (*ParsedLine, error) {
	if !strings.HasPrefix(line, dmLogPrefix) {
		return nil, nil
	}

	tokens := strings.Split(line[6:], " ")
	if len(tokens) < 2 {
		return nil, errInvalidFormat
	}

	kind := tokens[0]

	data, err := parseData(kind, tokens[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %s", errInvalidData, err)
	}

	return &ParsedLine{
		Kind: kind,
		Data: data,
	}, nil
}

func parseData(kind, data string) (interface{}, error) {
	switch kind {
	case extractor.MsgBegin:
		return parseNumber(data)
	case extractor.MsgEnd:
		return parseNumber(data)
	case extractor.MsgBlock:
		return parseFromProto(data, &pbcosmos.Block{})
	case extractor.MsgTx:
		return parseFromProto(data, &pbcosmos.TxResult{})
	case extractor.MsgValidatorSetUpdate:
		return parseFromProto(data, &pbcosmos.ValidatorSetUpdates{})
	default:
		return nil, fmt.Errorf("%w: %s", errUnsupportedKind, kind)
	}
}

func parseNumber(str string) (uint64, error) {
	return strconv.ParseUint(str, 10, 64)
}

func parseTimestamp(ts *pbcosmos.Timestamp) time.Time {
	return time.Unix(ts.Seconds, int64(ts.Nanos)).UTC()
}

func parseFromProto(data string, message proto.Message) (proto.Message, error) {
	buf, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	err = proto.Unmarshal(buf, message)
	return message, err
}
