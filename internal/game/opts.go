package game

import (
	"context"
	"image"

	"github.com/gabe565/cli-of-life/internal/pattern"
)

type Option func(game *Game)

func WithPattern(pat pattern.Pattern) Option {
	return func(game *Game) {
		game.pattern = pat
	}
}

func WithDimensions(width, height int) Option {
	return func(game *Game) {
		newW := max(width, game.BoardW()+100)
		newH := max(height, game.BoardH()+100)
		game.Resize(newW, newH, image.Pt(0, 0))
	}
}

func WithPlay(play bool) Option {
	return func(game *Game) {
		if play {
			game.ctx, game.cancel = context.WithCancel(context.Background())
			game.keymap.playPause.SetHelp(game.keymap.playPause.Help().Key, "pause")
		} else if game.ctx != nil {
			game.cancel()
			game.ctx, game.cancel = nil, nil
			game.keymap.playPause.SetHelp(game.keymap.playPause.Help().Key, "play")
		}
	}
}
