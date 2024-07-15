package pattern

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
)

func UnmarshalPlaintext(r io.Reader) (Pattern, error) {
	pattern := Pattern{Rule: GameOfLife()}
	scanner := bufio.NewScanner(r)
	var largest int
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
			if len(line)*(len(pattern.Grid)+1) > MaxTiles {
				return pattern, fmt.Errorf("rle: %w: w=%d, h=%d", ErrPatternTooBig, len(line), len(pattern.Grid))
			}
			tileLine := make([]int, len(line))
			var x int
			for _, b := range line {
				switch b {
				case '.':
					tileLine[x] = 0
					x++
				case 'O', '*':
					tileLine[x] = 1
					x++
				default:
					return pattern, fmt.Errorf("plaintext: %w: %q in line %q", ErrUnexpectedCharacter, string(b), line)
				}
			}
			if len(tileLine) > largest {
				largest = len(tileLine)
			}
			pattern.Grid = append(pattern.Grid, tileLine)
		}
	}
	for i := range pattern.Grid {
		if diff := largest - len(pattern.Grid[i]); diff > 0 {
			pattern.Grid[i] = append(pattern.Grid[i], make([]int, diff)...)
		}
	}
	if scanner.Err() != nil {
		return pattern, fmt.Errorf("plaintext: %w", scanner.Err())
	}
	pattern.Grid = slices.Clip(pattern.Grid)
	return pattern, nil
}
