package cmd

import (
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_completion(t *testing.T) {
	r, w := io.Pipe()
	_ = r.Close()

	type args struct {
		cmd   *cobra.Command
		shell string
	}
	tests := []struct {
		name    string
		w       io.Writer
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{"bash", io.Discard, args{New(), "bash"}, require.NoError},
		{"bash error", w, args{New(), "bash"}, require.Error},
		{"zsh", io.Discard, args{New(), "zsh"}, require.NoError},
		{"zsh error", w, args{New(), "zsh"}, require.Error},
		{"fish", io.Discard, args{New(), "fish"}, require.NoError},
		{"fish error", w, args{New(), "fish"}, require.Error},
		{"powershell", io.Discard, args{New(), "powershell"}, require.NoError},
		{"powershell error", w, args{New(), "powershell"}, require.Error},
		{"other", io.Discard, args{New(), "other"}, require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.cmd.SetOut(tt.w)

			if err := tt.args.cmd.Flags().Set("completion", tt.args.shell); !assert.NoError(t, err) {
				return
			}
			err := completion(tt.args.cmd, tt.args.shell)
			tt.wantErr(t, err)
		})
	}
}
