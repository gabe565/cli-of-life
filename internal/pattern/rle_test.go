package pattern

import (
	"bytes"
	_ "embed"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed glider.rle
var gliderRLE []byte

func TestUnmarshalRLE(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    Pattern
		wantErr require.ErrorAssertionFunc
	}{
		{
			"rule B3/S23",
			args{strings.NewReader("x = 3, y = 3, rule = B3/S23\n!")},
			Pattern{[][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, GameOfLife()},
			require.NoError,
		},
		{
			"rule b3/s23",
			args{strings.NewReader("x = 3, y = 3, rule = b3/s23\n!")},
			Pattern{[][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, GameOfLife()},
			require.NoError,
		},
		{
			"rule 23/3",
			args{strings.NewReader("x = 3, y = 3, rule = 23/3\n!")},
			Pattern{[][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, GameOfLife()},
			require.NoError,
		},
		{
			"high life",
			args{strings.NewReader("x = 3, y = 3, rule = B36/S23\n!")},
			Pattern{[][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, HighLife()},
			require.NoError,
		},
		{
			"no rule",
			args{strings.NewReader("x = 3, y = 3\n!")},
			Pattern{[][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, GameOfLife()},
			require.NoError,
		},
		{
			"no header spacing",
			args{strings.NewReader("x=3,y=3,rule=B3/S23\n!")},
			Pattern{[][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, GameOfLife()},
			require.NoError,
		},
		{
			"unsupported rule",
			args{strings.NewReader("x = 3, y = 3, rule = abc\n!")},
			Pattern{},
			require.Error,
		},
		{
			"invalid header",
			args{strings.NewReader("x = 3, a = 5\n!")},
			Pattern{},
			require.Error,
		},
		{
			"glider",
			args{bytes.NewReader(gliderRLE)},
			Pattern{[][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}}, GameOfLife()},
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalRLE(tt.args.r)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
