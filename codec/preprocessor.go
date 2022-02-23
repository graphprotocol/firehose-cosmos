package codec

import (
	"github.com/streamingfast/bstream"
)

func FilterPreprocessorFactory() func(includeExpr, excludeExpr string) (bstream.PreprocessFunc, error) {
	return func(includeExpr, excludeExpr string) (bstream.PreprocessFunc, error) {
		preproc := NOOPFilteringPreprocessor{}
		return preproc.PreprocessBlock, nil
	}
}

type NOOPFilteringPreprocessor struct {
}

func (f *NOOPFilteringPreprocessor) PreprocessBlock(blk *bstream.Block) (interface{}, error) {
	return blk, nil
}
