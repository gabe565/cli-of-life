package pattern

import (
	"bytes"
	_ "embed"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed glider.rle
var glider []byte

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
		{"glider", args{bytes.NewReader(glider)}, [][]int{{0, 1, 0}, {0, 0, 1}, {1, 1, 1}}, require.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalRLE(tt.args.r)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
