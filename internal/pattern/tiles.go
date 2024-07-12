package pattern

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrInvalidHeader     = errors.New("invalid header")
	ErrMissingTerminator = errors.New("missing terminator")
	ErrOverflow          = errors.New("overflow")
	ErrUnknownExtension  = errors.New("unknown pattern extension")
)

func UnmarshalFile(path string) ([][]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	switch filepath.Ext(path) {
	case ".rle":
		return UnmarshalRLE(f)
	case ".cells":
		return UnmarshalPlaintext(f)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownExtension, filepath.Ext(path))
	}
}
