package pattern

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
)

func UnmarshalRLE(r io.Reader) (Pattern, error) {
	var pattern Pattern
	scanner := bufio.NewScanner(r)
	var x, y int
	var needsClip bool
scan:
	for scanner.Scan() {
		line := scanner.Bytes()
		switch {
		case bytes.HasPrefix(line, []byte("#")):
			if name, found := bytes.CutPrefix(line, []byte("#N ")); found {
				pattern.Name = string(bytes.TrimSpace(name))
			} else if author, found := bytes.CutPrefix(line, []byte("#O ")); found {
				pattern.Author = string(bytes.TrimSpace(author))
			} else if len(line) > 3 && bytes.EqualFold(line[:3], []byte("#C ")) {
				if comment := bytes.TrimSpace(line[2:]); len(comment) != 0 {
					if len(pattern.Comment) != 0 {
						pattern.Comment += "\n"
					}
					pattern.Comment += string(comment)
				}
			}
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

			if w*h > MaxTiles {
				return pattern, fmt.Errorf("rle: %w: w=%d, h=%d", ErrPatternTooBig, w, h)
			}

			pattern.Grid = make([][]int, h)
			for i := range pattern.Grid {
				pattern.Grid[i] = make([]int, w)
			}
			continue
		default:
			if len(line) == 0 {
				continue
			}

			var runCount int
			for _, b := range line {
				switch {
				case b >= '0' && b <= '9':
					runCount *= 10
					runCount += int(b - '0')
				case b == '$':
					runCount = max(runCount, 1)
					if x != 0 || y != 0 {
						y += runCount
						x = 0
					}
					runCount = 0
				case b == '!':
					break scan
				default:
					runCount = max(runCount, 1)

					if y > len(pattern.Grid)-1 {
						diff := max(y-len(pattern.Grid)+1, 1)
						needsClip = true
						pattern.Grid = slices.Grow(pattern.Grid, diff)
						for range diff {
							var w int
							if len(pattern.Grid) == 0 {
								w = runCount
							} else {
								w = len(pattern.Grid[0])
							}
							if w*(y+1) > MaxTiles {
								return pattern, fmt.Errorf("rle: %w: w=%d, h=%d", ErrPatternTooBig, x, y)
							}
							pattern.Grid = append(pattern.Grid, make([]int, w))
						}
					}

					if x+runCount-1 > len(pattern.Grid[y])-1 {
						if (x+runCount)*y > MaxTiles {
							return pattern, fmt.Errorf("rle: %w: w=%d, h=%d", ErrPatternTooBig, x, y)
						}
						needsClip = true
						for i, row := range pattern.Grid {
							pattern.Grid[i] = append(row, make([]int, runCount)...)
						}
					}

					switch b {
					case 'b':
						for range runCount {
							pattern.Grid[y][x] = 0
							x++
						}
					case ' ':
					default:
						for range runCount {
							pattern.Grid[y][x] = 1
							x++
						}
					}
					runCount = 0
				}
			}
		}
	}
	if scanner.Err() != nil {
		return pattern, fmt.Errorf("rle: %w", scanner.Err())
	}

	if needsClip {
		pattern.Grid = slices.Clip(pattern.Grid)
		for i, row := range pattern.Grid {
			pattern.Grid[i] = slices.Clip(row)
		}
	}

	return pattern, nil
}
