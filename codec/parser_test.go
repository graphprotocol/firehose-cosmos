package codec

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
)

func TestParseLine(t *testing.T) {
	input := "DMLOG BLOCK Cg0Yt7m9AiIGCIXPuYEG"

	want := &ParsedLine{
		Kind: "BLOCK",
		Data: &pbcosmos.Block{
			Header: &pbcosmos.Header{
				Height: 5201079,
				Time:   &pbcosmos.Timestamp{Seconds: 1613653893},
			},
		},
	}

	got, err := parseLine(input)
	assert.NoError(t, err)

	if diff := cmp.Diff(want, got, cmp.Comparer(proto.Equal)); diff != "" {
		t.Errorf("parseLine(%q) mismatch (-want +got):\n%s", input, diff)
	}
}

func TestParseLine_NoPrefix(t *testing.T) {
	line, err := parseLine("BLOCK 5201079 0 Cg8KDRi3ub0CIgYIhc+5gQY=")
	assert.NoError(t, err)
	assert.Nil(t, line)
}

func TestParseLine_Errors(t *testing.T) {
	examples := []struct {
		input string
		err   error
	}{
		{"DMLOG", nil},
		{"DMLOG ", errInvalidFormat},
		{"DMLOG BLOCK", errInvalidFormat},
		{"DMLOG FOO BAR", errors.New("invalid data: unsupported kind: FOO")},
	}

	for _, example := range examples {
		t.Run(example.input, func(t *testing.T) {
			data, err := parseLine(example.input)

			assert.Nil(t, data)
			if example.err != nil {
				assert.Equal(t, example.err.Error(), err.Error())
			}
		})
	}
}

func TestParseData(t *testing.T) {
	input := "Cg0Y37i9AiIGCKWyp7IC"

	want := &pbcosmos.Block{
		Header: &pbcosmos.Header{
			Height: 5200991,
			Time:   &pbcosmos.Timestamp{Seconds: 642373925},
		},
	}

	data, err := parseData("BLOCK", input)
	if err != nil {
		t.Fatal(err)
	}
	got := data.(*pbcosmos.Block)

	if diff := cmp.Diff(want, got, cmp.Comparer(proto.Equal)); diff != "" {
		t.Errorf("parseData(%q) mismatch (-want +got):\n%s", input, diff)
	}
}

func TestParseData_UnsupportedKind(t *testing.T) {
	input := "Cg8KDRjfuL0CIgYIpbKnsgI="
	data, err := parseData("UNSUPPORTED", input)

	assert.Equal(t, nil, data)
	assert.ErrorContains(t, err, "unsupported kind: UNSUPPORTED")
}

func TestParseNumber(t *testing.T) {
	examples := []struct {
		input    string
		expected uint64
		err      string
	}{
		{input: "0", expected: uint64(0)},
		{input: "100", expected: uint64(100)},
		{input: "", err: `strconv.ParseUint: parsing "": invalid syntax`},
		{input: "-1", err: `strconv.ParseUint: parsing "-1": invalid syntax`},
		{input: "foobar", err: `strconv.ParseUint: parsing "foobar": invalid syntax`},
	}

	for _, example := range examples {
		number, err := parseNumber(example.input)
		if err != nil {
			assert.Equal(t, example.err, err.Error())
		}
		assert.Equal(t, example.expected, number)
	}
}

func TestParseTimestamp(t *testing.T) {
	examples := []struct {
		input    *pbcosmos.Timestamp
		expected time.Time
	}{
		{
			input: &pbcosmos.Timestamp{
				Seconds: 1613653893,
				Nanos:   2137,
			},
			expected: time.Date(2021, 2, 18, 13, 11, 33, 2137, time.UTC),
		},
	}

	for _, example := range examples {
		assert.Equal(t, example.expected, parseTimestamp(example.input))
	}
}
