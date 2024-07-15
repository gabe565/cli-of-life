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
	var patternLine int
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
				case 'b', 'o':
					for range runCount {
						if y > len(pattern.Grid)-1 {
							diff := max(y+1-len(pattern.Grid), 1)
							needsClip = true
							for range diff {
								if len(pattern.Grid) == 0 {
									pattern.Grid = append(pattern.Grid, make([]int, runCount))
								} else {
									pattern.Grid = append(pattern.Grid, make([]int, len(pattern.Grid[0])))
								}
							}
						}
						if x > len(pattern.Grid[y])-1 {
							needsClip = true
							for i, row := range pattern.Grid {
								pattern.Grid[i] = append(row, make([]int, runCount)...)
							}
						}

						switch line[i] {
						case 'b':
							pattern.Grid[y][x] = 0
						case 'o':
							pattern.Grid[y][x] = 1
						}
						x++
					}
				case '$':
					if i != 0 || patternLine != 0 {
						y += runCount
						x = 0
					}
				case '!':
					break scan
				}
				patternLine++
				if i++; i > len(line)-1 {
					continue scan
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
