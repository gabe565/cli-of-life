package pattern

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Pattern struct {
	Name    string
	Comment string
	Author  string
	Grid    [][]int
	Rule    Rule
}

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
	ErrInvalidHeader    = errors.New("invalid header")
	ErrOverflow         = errors.New("overflow")
	ErrUnknownExtension = errors.New("unknown pattern extension")
)

const (
	ExtRLE       = ".rle"
	ExtPlaintext = ".cells"
)

func UnmarshalFile(path string, format Format) (Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return Pattern{}, err
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
		return Pattern{}, fmt.Errorf("%w: %s", ErrUnknownExtension, filepath.Ext(path))
	}
}
