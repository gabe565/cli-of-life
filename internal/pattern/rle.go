package pattern

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

func UnmarshalRLEFile(path string) ([][]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	return UnmarshalRLE(f)
}

func UnmarshalRLE(r io.Reader) ([][]int, error) {
	var tiles [][]int
	scanner := bufio.NewScanner(r)
	var x, y int
scan:
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.HasPrefix(line, []byte("#")) {
			continue
		}
		if bytes.HasPrefix(line, []byte("x = ")) {
			rleHeaderRe := regexp.MustCompile(`^x *= *(\d+), *y *= *(\d+)`)
			matches := rleHeaderRe.FindAllStringSubmatch(scanner.Text(), -1)
			if len(matches) == 0 {
				return nil, fmt.Errorf("rle: %w", ErrInvalidHeader)
			}
			w, err := strconv.Atoi(matches[0][1])
			if err != nil {
				return nil, err
			}
			h, err := strconv.Atoi(matches[0][2])
			if err != nil {
				return nil, err
			}
			tiles = make([][]int, h)
			for i := range tiles {
				tiles[i] = make([]int, w)
			}
			continue
		}

		var i int
		for {
			var runCount int
			for line[i] >= '0' && line[i] <= '9' {
				runCount *= 10
				runCount += int(line[i] - '0')
				if i++; i > len(line)-1 {
					continue scan
				}
			}
			if runCount == 0 {
				runCount = 1
			}

			switch line[i] {
			case 'b':
				for range runCount {
					if y > len(tiles)-1 || x > len(tiles[0])-1 {
						return nil, fmt.Errorf("rle: %w", ErrOverflow)
					}
					tiles[y][x] = 0
					x++
				}
			case 'o':
				for range runCount {
					if y > len(tiles)-1 || x > len(tiles[0])-1 {
						return nil, fmt.Errorf("rle: %w", ErrOverflow)
					}
					tiles[y][x] = 1
					x++
				}
			case '$':
				y += runCount
				x = 0
			case '!':
				return tiles, nil
			}
			if i++; i > len(line)-1 {
				continue scan
			}
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return nil, fmt.Errorf("rle: %w", ErrMissingTerminator)
}
