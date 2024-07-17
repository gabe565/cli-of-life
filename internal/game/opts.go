package game

import (
	"context"

	"github.com/gabe565/cli-of-life/internal/config"
	"github.com/gabe565/cli-of-life/internal/pattern"
)

type Option func(game *Game)

func WithPattern(pat pattern.Pattern) Option {
	return func(game *Game) {
		game.startPattern = pat
		game.pattern = game.startPattern
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
		WithPlay(c.Play)(game)
	}
}
