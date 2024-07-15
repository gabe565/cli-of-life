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
		default:
			tileLine := make([]int, len(line))
			var x int
			for _, b := range line {
				switch b {
				case '.':
					tileLine[x] = 0
					x++
				case 'O':
					tileLine[x] = 1
					x++
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
