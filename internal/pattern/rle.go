package pattern

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

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
		if len(tiles) == 0 && bytes.HasPrefix(line, []byte("x")) {
			rleHeaderRe := regexp.MustCompile(`^x *= *(?P<x>\d+), *y *= *(?P<y>\d+)(?:, *rule *= *(?P<rule>.+))?$`)
			matches := rleHeaderRe.FindStringSubmatch(scanner.Text())

			if len(matches) == 0 {
				return nil, fmt.Errorf("rle: %w", ErrInvalidHeader)
			}

			var w, h int
			var err error
			for i, name := range rleHeaderRe.SubexpNames() {
				switch name {
				case "x":
					if w, err = strconv.Atoi(matches[i]); err != nil {
						return nil, err
					}
				case "y":
					if h, err = strconv.Atoi(matches[i]); err != nil {
						return nil, err
					}
				case "rule":
					if matches[i] != "" && strings.ToUpper(matches[i]) != "B3/S23" && matches[i] != "23/3" {
						return nil, fmt.Errorf("rle: %w: %s", ErrUnsupportedRule, matches[i])
					}
				}
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
