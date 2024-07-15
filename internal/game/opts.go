package game

import (
	"context"
	"image"

	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/pattern"
)

type Option func(game *Game)

func WithPattern(pat pattern.Pattern) Option {
	return func(game *Game) {
		game.startPattern = pat
		game.Reset()
	}
}

func WithDimensions(width, height uint) Option {
	return func(game *Game) {
		newW := max(int(width), game.BoardW()+200)
		newH := max(int(height), game.BoardH()+200)
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

func WithConfig(c *config.Config) Option {
	return func(game *Game) {
		game.conf = c
		WithDimensions(c.Width, c.Height)(game)
		WithPlay(c.Play)(game)
	}
}
