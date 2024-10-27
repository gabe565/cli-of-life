package pattern

import (
	"bytes"
	_ "embed"
	"io"
	"testing"

	"gabe565.com/cli-of-life/internal/rule"
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
		name     string
		args     args
		want     *Pattern
		wantGrid [][]int
		wantErr  require.ErrorAssertionFunc
	}{
		{
			"glider",
			args{bytes.NewReader(gliderPlaintext)},
			&Pattern{
				Name:    "Glider",
				Comment: "The smallest, most common, and first discovered spaceship.\nwww.conwaylife.com/wiki/index.php?title=Glider",
				Author:  "Richard K. Guy",
				Rule:    rule.GameOfLife(),
			},
			[][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}},
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalPlaintext(tt.args.r)
			tt.wantErr(t, err)
			if len(tt.wantGrid) != 0 {
				assert.Equal(t, tt.wantGrid, got.Tree.ToSlice())
			}
			if got != nil {
				got.Tree = nil
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}
