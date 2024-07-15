package pattern

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRule_UnmarshalText(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Rule
		wantErr require.ErrorAssertionFunc
	}{
		{"B3/S23", args{b: []byte("B3/S23")}, GameOfLife(), require.NoError},
		{"23/3", args{b: []byte("23/3")}, GameOfLife(), require.NoError},
		{"Life edge case", args{b: []byte("Life")}, GameOfLife(), require.NoError},
		{"B36/S23", args{b: []byte("B36/S23")}, HighLife(), require.NoError},
		{"23/36", args{b: []byte("23/36")}, HighLife(), require.NoError},
		{"no slash", args{b: []byte("abc")}, Rule{}, require.Error},
		{"invalid num", args{b: []byte("B3/S2A")}, Rule{}, require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r Rule
			tt.wantErr(t, r.UnmarshalText(tt.args.b))
			assert.Equal(t, tt.want, r)
		})
	}
}

func TestRule_String(t *testing.T) {
	tests := []struct {
		name   string
		fields Rule
		want   string
	}{
		{"B3/S23", GameOfLife(), "B3/S23"},
		{"B36/S23", HighLife(), "B36/S23"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{
				Born:    tt.fields.Born,
				Survive: tt.fields.Survive,
			}
			assert.Equal(t, tt.want, r.String())
		})
	}
}
