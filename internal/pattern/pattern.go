package pattern

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

const MaxTiles = 33_554_432

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
	ErrInvalidHeader       = errors.New("invalid header")
	ErrPatternTooBig       = errors.New("pattern too big")
	ErrUnexpectedCharacter = errors.New("unexpected character")
	ErrInferFailed         = errors.New("unable to infer pattern file type")
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
		pattern, err := Unmarshal(f)
		if err != nil {
			err = fmt.Errorf("%w: %s", err, path)
		}
		return pattern, err
	}
}

var ErrResponse = errors.New("received an error response")

func UnmarshalURL(ctx context.Context, url string, format Format) (Pattern, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Pattern{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Pattern{}, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return Pattern{}, ErrResponse
	}

	ext := path.Ext(url)
	switch {
	case format == FormatRLE, ext == ExtRLE:
		return UnmarshalRLE(resp.Body)
	case format == FormatPlaintext, ext == ExtPlaintext:
		return UnmarshalPlaintext(resp.Body)
	default:
		pattern, err := Unmarshal(resp.Body)
		if err != nil {
			err = fmt.Errorf("%w: %s", err, url)
		}
		return pattern, err
	}
}

func Unmarshal(r io.Reader) (Pattern, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Pattern{}, err
	}

	firstLine, _, _ := bytes.Cut(bytes.TrimSpace(buf), []byte("\n"))
	switch {
	case bytes.HasPrefix(firstLine, []byte("#")), RLEHeaderRegexp().Match(firstLine):
		return UnmarshalRLE(bytes.NewReader(buf))
	case bytes.HasPrefix(firstLine, []byte("!")), bytes.HasPrefix(firstLine, []byte(".")), bytes.HasPrefix(firstLine, []byte("O")):
		return UnmarshalPlaintext(bytes.NewReader(buf))
	default:
		return Pattern{}, ErrInferFailed
	}
}
