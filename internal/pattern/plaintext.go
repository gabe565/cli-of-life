package pattern

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
)

func UnmarshalPlaintext(r io.Reader) ([][]int, error) {
	var tiles [][]int
	scanner := bufio.NewScanner(r)
	var largest int
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.HasPrefix(line, []byte("!")) {
			continue
		}

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
		tiles = append(tiles, tileLine)
	}
	for i := range tiles {
		diff := largest - len(tiles[i])
		if diff > 0 {
			tiles[i] = append(tiles[i], make([]int, largest-len(tiles[i]))...)
		}
	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("plaintext: %w", scanner.Err())
	}
	return slices.Clip(tiles), nil
}
