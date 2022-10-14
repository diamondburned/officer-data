package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/officer-data/bot"
	"github.com/diamondburned/officer-data/internal/gitwork"
	"github.com/pkg/errors"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatalln(err)
	}
}

func run(ctx context.Context) error {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("$BOT_TOKEN is not set")
	}

	state := state.New(token)
	state = state.WithContext(ctx)
	state.AddIntents(bot.Intents)

	gitAuthor := gitwork.DefaultAuthor
	gitAuthor.Name = envOr(gitAuthor.Name, "GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME")
	gitAuthor.Email = envOr(gitAuthor.Email, "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL")

	gitworkDir := envOr(os.TempDir(), "GITWORK_DIR")
	gitworkRemote := envOr("https://github.com/EthanThatOneKid/acmcsuf.com.git", "GITWORK_REMOTE")

	gitPool, err := gitwork.NewPool(gitworkDir, gitworkRemote)
	if err != nil {
		return errors.Wrap(err, "cannot create git pool")
	}
	gitPool.Author = gitAuthor

	handler := bot.New(state, gitPool)
	state.AddInteractionHandler(handler)

	if err := handler.OverwriteCommands(); err != nil {
		return errors.Wrap(err, "cannot overwrite commands")
	}

	return state.Connect(ctx)
}

func envOr(def string, keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return def
}
