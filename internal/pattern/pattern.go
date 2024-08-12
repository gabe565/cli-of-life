package pattern

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gabe565/cli-of-life/internal/quadtree"
	"github.com/gabe565/cli-of-life/internal/rule"
)

type Pattern struct {
	Name    string
	Comment string
	Author  string
	Tree    *quadtree.Gosper
	Rule    rule.Rule
}

func (p Pattern) Step(steps uint) {
	p.Tree.Step(&p.Rule, steps)
}

var _ slog.LogValuer = Pattern{}

func (p Pattern) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, 4)
	if p.Name != "" {
		attrs = append(attrs, slog.String("name", p.Name))
	}
	if p.Author != "" {
		attrs = append(attrs, slog.String("author", p.Author))
	}
	attrs = append(attrs,
		slog.String("rule", p.Rule.String()),
		slog.String("size", p.Tree.FilledCoords().Size().String()),
	)
	return slog.GroupValue(attrs...)
}

var (
	ErrInvalidHeader       = errors.New("invalid header")
	ErrUnexpectedCharacter = errors.New("unexpected character")
	ErrInferFailed         = errors.New("unable to infer pattern file type")
)

const (
	ExtRLE       = ".rle"
	ExtPlaintext = ".cells"
)

func UnmarshalFile(path string) (Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return Pattern{}, err
	}
	defer func() {
		_ = f.Close()
	}()

	ext := filepath.Ext(path)
	switch {
	case ext == ExtRLE:
		return UnmarshalRLE(f)
	case ext == ExtPlaintext:
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

func UnmarshalURL(ctx context.Context, url string) (Pattern, error) {
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
	case ext == ExtRLE:
		return UnmarshalRLE(resp.Body)
	case ext == ExtPlaintext:
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
