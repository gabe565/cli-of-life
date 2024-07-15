package pattern

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

func UnmarshalRLE(r io.Reader) (Pattern, error) {
	var pattern Pattern
	scanner := bufio.NewScanner(r)
	var x, y int
scan:
	for scanner.Scan() {
		line := scanner.Bytes()
		switch {
		case bytes.HasPrefix(line, []byte("#")):
			continue
		case len(pattern.Grid) == 0 && bytes.HasPrefix(line, []byte("x")):
			rleHeaderRe := regexp.MustCompile(`^x *= *(?P<x>[^,]+), *y *= *(?P<y>[^,]+)(?:, *rule *= *(?P<rule>.+))?$`)
			matches := rleHeaderRe.FindStringSubmatch(scanner.Text())

			if len(matches) == 0 {
				return pattern, fmt.Errorf("rle: %w: %s", ErrInvalidHeader, line)
			}

			var w, h int
			var err error
			for i, name := range rleHeaderRe.SubexpNames() {
				switch name {
				case "x":
					if w, err = strconv.Atoi(matches[i]); err != nil {
						return pattern, fmt.Errorf("rle: parsing header x: %w", err)
					}
				case "y":
					if h, err = strconv.Atoi(matches[i]); err != nil {
						return pattern, fmt.Errorf("rle: parsing header y: %w", err)
					}
				case "rule":
					switch matches[i] {
					case "":
						pattern.Rule = GameOfLife()
					default:
						if err := pattern.Rule.UnmarshalText([]byte(matches[i])); err != nil {
							return pattern, fmt.Errorf("rle: %w", err)
						}
					}
				}
			}

			pattern.Grid = make([][]int, h)
			for i := range pattern.Grid {
				pattern.Grid[i] = make([]int, w)
			}
			continue
		default:
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
						if y > len(pattern.Grid)-1 || x > len(pattern.Grid[0])-1 {
							return pattern, fmt.Errorf("rle: %w", ErrOverflow)
						}
						pattern.Grid[y][x] = 0
						x++
					}
				case 'o':
					for range runCount {
						if y > len(pattern.Grid)-1 || x > len(pattern.Grid[0])-1 {
							return pattern, fmt.Errorf("rle: %w", ErrOverflow)
						}
						pattern.Grid[y][x] = 1
						x++
					}
				case '$':
					y += runCount
					x = 0
				case '!':
					return pattern, nil
				}
				if i++; i > len(line)-1 {
					continue scan
				}
			}
		}
	}
	if scanner.Err() != nil {
		return pattern, fmt.Errorf("rle: %w", scanner.Err())
	}
	return pattern, fmt.Errorf("rle: %w", ErrMissingTerminator)
}
