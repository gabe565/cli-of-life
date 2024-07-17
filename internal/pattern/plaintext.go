package pattern

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/gabe565/cli-of-life/internal/quadtree"
	"github.com/gabe565/cli-of-life/internal/rule"
)

func UnmarshalPlaintext(r io.Reader) (Pattern, error) {
	pattern := Pattern{
		Rule: rule.GameOfLife(),
		Tree: quadtree.Empty(quadtree.DefaultTreeSize),
	}
	scanner := bufio.NewScanner(r)
	var y int
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
			var x int
			pattern.Tree = pattern.Tree.GrowToFit(x, len(line))
			for _, b := range line {
				switch b {
				case '.':
					pattern.Tree = pattern.Tree.Set(x, y, 0)
					x++
				case 'O', '*':
					pattern.Tree = pattern.Tree.Set(x, y, 1)
					x++
				default:
					return pattern, fmt.Errorf("plaintext: %w: %q in line %q", ErrUnexpectedCharacter, string(b), line)
				}
			}
			y++
		}
	}
	if scanner.Err() != nil {
		return pattern, fmt.Errorf("plaintext: %w", scanner.Err())
	}
	return pattern, nil
}
