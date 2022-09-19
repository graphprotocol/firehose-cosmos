package codec

import (
	"testing"

	pbcosmos "github.com/figment-networks/proto-cosmos/pb/sf/cosmos/type/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func makeLinesChan(lines ...string) chan string {
	ch := make(chan string, len(lines))
	for _, line := range lines {
		ch <- line
	}
	close(ch)
	return ch
}

func TestConsoleReader(t *testing.T) {
	t.Run("normal path", func(t *testing.T) {
		lines := makeLinesChan(
			"DMLOG BEGIN 5201079",
			"DMLOG BLOCK Cg0Yt7m9AiIGCIXPuYEG",
			"DMLOG TX CLe5vQIQAQ==",
			"DMLOG TX CLe5vQIQAQ==",
			"DMLOG VSET_UPDATE CgkKBGFkZHIYuWA=",
			"DMLOG END 5201079",
		)

		reader, err := NewConsoleReader(lines, zap.NewNop())
		assert.NoError(t, err)

		blockObj, err := reader.Read()
		assert.NoError(t, err)
		assert.IsType(t, &pbcosmos.Block{}, blockObj)

		block := blockObj.(*pbcosmos.Block)
		assert.Equal(t, uint64(5201079), block.Header.Height)
		assert.Len(t, block.Transactions, 2)
		assert.Len(t, block.ValidatorUpdates, 1)
	})
}

func TestConsoleReaderValidation(t *testing.T) {
	examples := []struct {
		name  string
		lines []string
		err   string
	}{
		{
			name:  "received nothing",
			lines: []string{},
			err:   "EOF",
		},
		{
			name:  "invalid dmlog marker",
			lines: []string{"DMLOG FOO"},
			err:   `invalid format (line "DMLOG FOO")`,
		},
		{
			name: "unexpected start height marker",
			lines: []string{
				"DMLOG BEGIN 5201079",
				"DMLOG BEGIN 5201079",
			},
			err: "unexpected start height 5201079",
		},
		{
			name: "unexpected end height marker",
			lines: []string{
				"DMLOG BEGIN 5201079",
				"DMLOG END 5201080",
			},
			err: "unexpected end height end: 5201080",
		},
	}

	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			reader, err := NewConsoleReader(makeLinesChan(ex.lines...), zap.NewNop())
			assert.NoError(t, err)

			_, err = reader.Read()
			assert.EqualError(t, err, ex.err)
		})
	}
}
