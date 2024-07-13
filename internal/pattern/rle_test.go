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
		want    [][]int
		wantErr require.ErrorAssertionFunc
	}{
		{"rule B23/S23", args{strings.NewReader("x = 3, y = 3, rule = B3/S23\n!")}, [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, require.NoError},
		{"rule b23/s23", args{strings.NewReader("x = 3, y = 3, rule = b3/s23\n!")}, [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, require.NoError},
		{"rule 23/3", args{strings.NewReader("x = 3, y = 3, rule = 23/3\n!")}, [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, require.NoError},
		{"no rule", args{strings.NewReader("x = 3, y = 3\n!")}, [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, require.NoError},
		{"no header spacing", args{strings.NewReader("x=3,y=3,rule=B3/S23\n!")}, [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}, require.NoError},
		{"unsupported rule", args{strings.NewReader("x = 3, y = 3, rule = B36/S23")}, nil, require.Error},
		{"invalid header", args{strings.NewReader("x = 3, a = 5\n!")}, nil, require.Error},
		{"glider", args{bytes.NewReader(gliderRLE)}, [][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}}, require.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalRLE(tt.args.r)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
