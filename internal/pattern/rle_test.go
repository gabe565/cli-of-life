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
		{"glider", args{bytes.NewReader(gliderRLE)}, [][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}}, require.NoError},
		{"unsupported", args{strings.NewReader("x = 3, y = 3, rule = B36/S23")}, nil, require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalRLE(tt.args.r)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
