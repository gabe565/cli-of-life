package pattern

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"io"

	"github.com/gabe565/cli-of-life/internal/quadtree"
	"github.com/gabe565/cli-of-life/internal/rule"
)

func UnmarshalPlaintext(r io.Reader) (*Pattern, error) {
	pattern := &Pattern{
		Rule: rule.GameOfLife(),
		Tree: quadtree.New(),
	}
	scanner := bufio.NewScanner(r)
	var p image.Point
	for scanner.Scan() {
		line := scanner.Bytes()
		switch {
		case bytes.HasPrefix(line, []byte("!")):
			if name, found := bytes.CutPrefix(line, []byte("!Name: ")); found {
				pattern.Name = string(bytes.TrimSpace(name))
			} else if author, found := bytes.CutPrefix(line, []byte("!Author: ")); found {
				pattern.Author = string(bytes.TrimSpace(author))
			} else if comment := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("!"))); len(line) != 0 {
				if len(pattern.Comment) != 0 {
					pattern.Comment += "\n"
				}
				pattern.Comment += string(comment)
			}
		default:
			pattern.Tree.GrowToFit(p.Add(image.Pt(0, len(line))))
			for _, b := range line {
				switch b {
				case '.':
					pattern.Tree.Set(p, 0)
					p.X++
				case 'O', '*':
					pattern.Tree.Set(p, 1)
					p.X++
				default:
					return nil, fmt.Errorf("plaintext: %w: %q in line %q", ErrUnexpectedCharacter, string(b), line)
				}
			}
			p.X = 0
			p.Y++
		}
	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("plaintext: %w", scanner.Err())
	}
	pattern.Tree.SetReset()
	return pattern, nil
}
