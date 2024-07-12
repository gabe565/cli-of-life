package pattern

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Format string

const (
	FormatAuto      Format = "auto"
	FormatRLE       Format = "rle"
	FormatPlaintext Format = "plaintext"
)

func FormatStrings() []string {
	return []string{string(FormatAuto), string(FormatRLE), string(FormatPlaintext)}
}

var (
	ErrInvalidHeader     = errors.New("invalid header")
	ErrMissingTerminator = errors.New("missing terminator")
	ErrOverflow          = errors.New("overflow")
	ErrUnknownExtension  = errors.New("unknown pattern extension")
	ErrUnsupportedRule   = errors.New("unsupported rule designation")
)

const (
	ExtRLE       = ".rle"
	ExtPlaintext = ".cells"
)

func UnmarshalFile(path string, format Format) ([][]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	ext := filepath.Ext(path)
	switch {
	case format == FormatRLE, ext == ExtRLE:
		return UnmarshalRLE(f)
	case format == FormatPlaintext, ext == ExtPlaintext:
		return UnmarshalPlaintext(f)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownExtension, filepath.Ext(path))
	}
}
