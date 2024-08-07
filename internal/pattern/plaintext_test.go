package pattern

import (
	"bytes"
	_ "embed"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed glider.cells
var gliderPlaintext []byte

func TestUnmarshalPlaintext(t *testing.T) {
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
			"glider",
			args{bytes.NewReader(gliderPlaintext)},
			Pattern{
				Name:    "Glider",
				Comment: "The smallest, most common, and first discovered spaceship.\nwww.conwaylife.com/wiki/index.php?title=Glider",
				Author:  "Richard K. Guy",
				Grid:    [][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}},
				Rule:    GameOfLife(),
			},
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalPlaintext(tt.args.r)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
