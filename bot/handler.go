package bot

import (
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/officer-data/internal/gitwork"
	"github.com/pkg/errors"
)

// Intents is the set of intents that the bot needs.
var Intents = 0 |
	gateway.IntentGuilds

// New creates a new bot instance.
func New(state *state.State, gitPool *gitwork.Pool) *Handler {
	h := Handler{
		state:  state,
		router: cmdroute.NewRouter(),
		gits:   gitPool,
	}

	h.router.Use(cmdroute.UseContext(state.Context()))
	h.router.Use(cmdroute.Deferrable(state.Client, cmdroute.DeferOpts{
		Timeout: 2 * time.Second,
	}))
	h.router.Sub("officer", func(r *cmdroute.Router) {
		r.AddFunc("link", h.handleLink)
		r.AddFunc("set", h.handleSet)
		r.AddFunc("add-term", h.handleAddTerm)
		r.AddFunc("pr", h.handlePR)
	})

	return &h
}

type Handler struct {
	state  *state.State
	router *cmdroute.Router
	gits   *gitwork.Pool
}

func (h *Handler) HandleInteraction(ev *discord.InteractionEvent) *api.InteractionResponse {
	resp := h.router.HandleInteraction(ev)
	if resp != nil {
		return resp
	}
	return errorResponse(errors.New("unknown interaction"))
}

// OverwriteCommands overwrites the commands to the ones defined in Commands.
func (h *Handler) OverwriteCommands() error {
	app, err := h.state.CurrentApplication()
	if err != nil {
		return errors.Wrap(err, "cannot get current app")
	}

	_, err = h.state.BulkOverwriteCommands(app.ID, commands)
	if err != nil {
		return errors.Wrap(err, "cannot overwrite old commands")
	}

	return nil
}

func errorResponse(err error) *api.InteractionResponseData {
	return &api.InteractionResponseData{
		Content:         option.NewNullableString("**Error:** " + err.Error()),
		Flags:           discord.EphemeralMessage,
		AllowedMentions: &api.AllowedMentions{ /* none */ },
	}
}
