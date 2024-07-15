package pattern

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type parseSection uint8

const (
	parseSurvive parseSection = iota
	parseBorn
)

type Rule struct {
	Born    []int
	Survive []int
}

var ErrUnsupportedRule = errors.New("unsupported rule string")

func (r *Rule) UnmarshalText(text []byte) error {
	if !bytes.Contains(text, []byte("/")) {
		return fmt.Errorf("%w: %s", ErrUnsupportedRule, text)
	}

	var born, survive []int
	var section parseSection
	for _, b := range bytes.ToUpper(text) {
		switch b {
		case '/':
			if section == parseSurvive {
				section = parseBorn
			}
		case 'B':
			section = parseBorn
		case 'S':
			section = parseSurvive
		default:
			val, err := strconv.Atoi(string(b))
			if err != nil {
				return fmt.Errorf("%w: %s", ErrUnsupportedRule, text)
			}

			switch section {
			case parseBorn:
				born = append(born, val)
			case parseSurvive:
				survive = append(survive, val)
			default:
				panic("section is invalid")
			}
		}
	}

	r.Born = slices.Clip(born)
	r.Survive = slices.Clip(survive)
	return nil
}

func (r Rule) IsZero() bool {
	return len(r.Born) == 0 && len(r.Survive) == 0
}

func (r Rule) String() string {
	var buf strings.Builder
	buf.Grow(3 + len(r.Born) + len(r.Survive))
	buf.WriteByte('B')
	for _, v := range r.Born {
		buf.WriteByte(byte(v + '0'))
	}
	buf.WriteByte('/')
	buf.WriteByte('S')
	for _, v := range r.Survive {
		buf.WriteByte(byte(v + '0'))
	}
	return buf.String()
}

func GameOfLife() Rule {
	return Rule{
		Born:    []int{3},
		Survive: []int{2, 3},
	}
}

func HighLife() Rule {
	return Rule{
		Born:    []int{3, 6},
		Survive: []int{2, 3},
	}
}
