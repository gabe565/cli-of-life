package pattern

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"io"
	"regexp"
	"strconv"
	"strings"

	"gabe565.com/cli-of-life/internal/rule"
)

func RLEHeaderRegexp() *regexp.Regexp {
	return regexp.MustCompile(`^x *= *(?P<x>[^,]+), *y *= *(?P<y>[^,]+)(?:, *rule *= *(?P<rule>.+))?$`)
}

func UnmarshalRLE(r io.Reader) (*Pattern, error) {
	pattern := Default()
	scanner := bufio.NewScanner(r)
	var p image.Point
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
		case bytes.HasPrefix(line, []byte("x")):
			headerRe := RLEHeaderRegexp()
			matches := headerRe.FindStringSubmatch(scanner.Text())

			if len(matches) == 0 {
				return nil, fmt.Errorf("rle: %w: %q", ErrInvalidHeader, line)
			}

			var w, h int
			var err error
			for i, name := range headerRe.SubexpNames() {
				switch name {
				case "x":
					if w, err = strconv.Atoi(matches[i]); err != nil {
						return nil, fmt.Errorf("rle: parsing header x: %w", err)
					}
				case "y":
					if h, err = strconv.Atoi(matches[i]); err != nil {
						return nil, fmt.Errorf("rle: parsing header y: %w", err)
					}
				case "rule":
					switch matches[i] {
					case "":
						pattern.Rule = rule.GameOfLife()
					default:
						if err := pattern.Rule.UnmarshalText([]byte(matches[i])); err != nil {
							return nil, fmt.Errorf("rle: %w", err)
						}
					}
				}
			}

			pattern.Tree.GrowToFit(image.Pt(w, h))
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
					if p.X != 0 || p.Y != 0 {
						p.Y += runCount
						p.X = 0
					}
					runCount = 0
				case b == '!':
					break scan
				default:
					runCount = max(runCount, 1)
					switch b {
					case 'b':
						for range runCount {
							pattern.Tree.Set(p, 0)
							p.X++
						}
					case ' ':
					default:
						for range runCount {
							pattern.Tree.Set(p, 1)
							p.X++
						}
					}
					runCount = 0
				}
			}
		}
	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("rle: %w", scanner.Err())
	}
	pattern.Tree.SetReset()
	return pattern, nil
}

const rleLineLength = 70

func MarshalRLE(w io.Writer, p *Pattern) error {
	bw := bufio.NewWriter(w)

	if p.Name != "" {
		if _, err := fmt.Fprintf(bw, "#N %s\n", p.Name); err != nil {
			return err
		}
	}
	if p.Author != "" {
		if _, err := fmt.Fprintf(bw, "#O %s\n", p.Author); err != nil {
			return err
		}
	}
	for _, line := range strings.Split(p.Comment, "\n") {
		if line == "" {
			continue
		}
		if _, err := fmt.Fprintf(bw, "#C %s\n", line); err != nil {
			return err
		}
	}

	grid := p.Tree.ToSlice()
	height := len(grid)
	width := 0
	if height != 0 {
		width = len(grid[0])
	}

	if _, err := fmt.Fprintf(bw, "x = %d, y = %d, rule = %s\n", width, height, p.Rule.String()); err != nil {
		return err
	}

	var tokens []string
	for y, row := range grid {
		runChar := byte(0)
		runCount := 0
		flush := func() {
			if runCount == 0 {
				return
			}
			c := byte('b')
			if runChar != 0 {
				c = 'o'
			}
			token := string(c)
			if runCount > 1 {
				token = strconv.Itoa(runCount) + token
			}
			tokens = append(tokens, token)
			runCount = 0
		}

		for _, v := range row {
			c := byte(0)
			if v != 0 {
				c = 1
			}
			if runCount > 0 && c != runChar {
				flush()
			}
			runChar = c
			runCount++
		}
		if runChar != 0 {
			flush()
		} else {
			runCount = 0
		}

		if y < height-1 {
			tokens = append(tokens, "$")
		}
	}
	tokens = append(tokens, "!")

	var lineLen int
	for i, token := range tokens {
		if lineLen != 0 && lineLen+len(token) > rleLineLength {
			if _, err := bw.WriteString("\n"); err != nil {
				return err
			}
			lineLen = 0
		}
		if _, err := bw.WriteString(token); err != nil {
			return err
		}
		lineLen += len(token)
		if i == len(tokens)-1 {
			if _, err := bw.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	return bw.Flush()
}
