package pattern

import (
	"bytes"
	_ "embed"
	"io"
	"strings"
	"testing"

	"gabe565.com/cli-of-life/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed embedded/glider.rle
var gliderRLE []byte

func TestUnmarshalRLE(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name     string
		args     args
		want     *Pattern
		wantGrid [][]int
		wantErr  require.ErrorAssertionFunc
	}{
		{
			"rule B3/S23",
			args{strings.NewReader("x = 3, y = 3, rule = B3/S23\n!")},
			&Pattern{Rule: rule.GameOfLife()},
			nil,
			require.NoError,
		},
		{
			"rule b3/s23",
			args{strings.NewReader("x = 3, y = 3, rule = b3/s23\n!")},
			&Pattern{Rule: rule.GameOfLife()},
			nil,
			require.NoError,
		},
		{
			"rule 23/3",
			args{strings.NewReader("x = 3, y = 3, rule = 23/3\n!")},
			&Pattern{Rule: rule.GameOfLife()},
			nil,
			require.NoError,
		},
		{
			"high life",
			args{strings.NewReader("x = 3, y = 3, rule = B36/S23\n!")},
			&Pattern{Rule: rule.HighLife()},
			nil,
			require.NoError,
		},
		{
			"no rule",
			args{strings.NewReader("x = 3, y = 3\n!")},
			&Pattern{Rule: rule.GameOfLife()},
			nil,
			require.NoError,
		},
		{
			"no header spacing",
			args{strings.NewReader("x=3,y=3,rule=B3/S23\n!")},
			&Pattern{Rule: rule.GameOfLife()},
			nil,
			require.NoError,
		},
		{
			"unsupported rule",
			args{strings.NewReader("x = 3, y = 3, rule = abc\n!")},
			nil,
			nil,
			require.Error,
		},
		{
			"invalid header",
			args{strings.NewReader("x = 3, a = 5\n!")},
			nil,
			nil,
			require.Error,
		},
		{
			"glider",
			args{bytes.NewReader(gliderRLE)},
			&Pattern{
				Name:    "Glider",
				Comment: "The smallest, most common, and first discovered spaceship. Diagonal, has period 4 and speed c/4.\nwww.conwaylife.com/wiki/index.php?title=Glider",
				Author:  "Richard K. Guy",
				Rule:    rule.GameOfLife(),
			},
			[][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}},
			require.NoError,
		},
		{
			"edge case first is $",
			args{strings.NewReader("x = 1, y = 2\n$o$o!")},
			&Pattern{Rule: rule.GameOfLife()},
			[][]int{{1}, {1}},
			require.NoError,
		},
		{
			"edge case missing !",
			args{strings.NewReader("x = 1, y = 1\no")},
			&Pattern{Rule: rule.GameOfLife()},
			[][]int{{1}},
			require.NoError,
		},
		{
			"edge case incorrect x",
			args{strings.NewReader("x = 0, y = 1\no!")},
			&Pattern{Rule: rule.GameOfLife()},
			[][]int{{1}},
			require.NoError,
		},
		{
			"edge case incorrect y",
			args{strings.NewReader("x = 1, y = 0\no!")},
			&Pattern{Rule: rule.GameOfLife()},
			[][]int{{1}},
			require.NoError,
		},
		{
			"blank lines",
			args{strings.NewReader("x = 1, y = 1\n\n\no!")},
			&Pattern{Rule: rule.GameOfLife()},
			[][]int{{1}},
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalRLE(tt.args.r)
			tt.wantErr(t, err)
			if len(tt.wantGrid) != 0 {
				assert.Equal(t, tt.wantGrid, got.Tree.ToSlice())
			}
			if got != nil {
				got.Tree = nil
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
