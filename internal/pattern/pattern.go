package pattern

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/pattern/embedded"
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

func (p Pattern) NameAuthor() string {
	val := p.Name
	if p.Author != "" {
		val += " by " + p.Author
	}
	return val
}

func Default() *Pattern {
	return &Pattern{
		Tree: quadtree.New(),
		Rule: rule.GameOfLife(),
	}
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

func Extensions() []string {
	return []string{ExtRLE, ExtPlaintext}
}

func UnmarshalFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
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

var ErrResponse = errors.New("HTTP error")

func UnmarshalURL(ctx context.Context, url string) (*Pattern, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", ErrResponse, resp.Status)
	}

	ext := path.Ext(url)
	switch {
	case strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html"):
		urls, err := FindHrefPatterns(resp)
		if err != nil {
			return nil, err
		}

		if len(urls) > 1 {
			return nil, MultiplePatternsError{urls}
		}

		return UnmarshalURL(ctx, urls[0])
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

func Unmarshal(r io.Reader) (*Pattern, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	firstLine, _, _ := bytes.Cut(bytes.TrimSpace(buf), []byte("\n"))
	switch {
	case bytes.HasPrefix(firstLine, []byte("#")), RLEHeaderRegexp().Match(firstLine):
		return UnmarshalRLE(bytes.NewReader(buf))
	case bytes.HasPrefix(firstLine, []byte("!")), bytes.HasPrefix(firstLine, []byte(".")), bytes.HasPrefix(firstLine, []byte("O")):
		return UnmarshalPlaintext(bytes.NewReader(buf))
	default:
		return nil, ErrInferFailed
	}
}

func New(conf *config.Config) (*Pattern, error) {
	var r rule.Rule
	if err := r.UnmarshalText([]byte(conf.RuleString)); err != nil {
		return nil, err
	}

	quadtree.ClearCache()

	var p *Pattern
	switch {
	case conf.Pattern != "":
		u, err := url.Parse(conf.Pattern)
		if err != nil {
			return nil, err
		}

		switch u.Scheme {
		case "embedded":
			slog.Info("Loading embedded pattern", "path", conf.Pattern)
			f, err := embedded.Embedded.Open(strings.TrimPrefix(conf.Pattern, "embedded://"))
			if err != nil {
				return nil, err
			}
			p, err = Unmarshal(f)
			if err != nil {
				return nil, err
			}
		case "http", "https":
			slog.Info("Loading pattern URL", "url", conf.Pattern)
			p, err = UnmarshalURL(context.Background(), conf.Pattern)
			if err != nil {
				return nil, err
			}
		default:
			slog.Info("Loading pattern file", "path", conf.Pattern)
			p, err = UnmarshalFile(conf.Pattern)
			if err != nil {
				return nil, err
			}
		}

		if p.Name == "" {
			p.Name = filepath.Base(conf.Pattern)
		}
		slog.Info("Loaded pattern", "pattern", p)
	default:
		p = Default()
	}

	return p, nil
}
