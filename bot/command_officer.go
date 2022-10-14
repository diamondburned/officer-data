package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/diamondburned/officer-data/acmcsuf"
	"github.com/diamondburned/officer-data/internal/gitwork"
	"github.com/pkg/errors"
)

// initUserWorkspace gives each Discord user a workspace in the git repo.
func (h *Handler) initUserWorkspace(ctx context.Context, guildID discord.GuildID, user *discord.User) (*gitwork.PooledRepository, error) {
	dirPath := filepath.Join(guildID.String(), user.ID.String())

	repo, err := h.gits.Clone(ctx, true, dirPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to clone git repo")
	}

	return repo, nil
}

type commit struct {
	Title string
	Body  string
}

func (h *Handler) updateOfficers(ctx context.Context, command cmdroute.CommandData, updateFn func(officers *acmcsuf.Officers) (commit, error)) *api.InteractionResponseData {
	repo, err := h.initUserWorkspace(ctx, command.Event.GuildID, command.Event.User)
	if err != nil {
		return errorResponse(err)
	}

	f, err := repo.OpenFile(acmcsuf.OfficersJSONPath, os.O_RDWR|gitwork.LockFile)
	if err != nil {
		return errorResponse(errors.Wrap(err, "failed to open officers.json"))
	}
	defer f.Close()

	var officers acmcsuf.Officers
	if err := json.NewDecoder(f).Decode(&officers); err != nil {
		return errorResponse(errors.Wrap(err, "failed to decode officers.json"))
	}

	commit, err := updateFn(&officers)
	if err != nil {
		return errorResponse(err)
	}

	if err := f.Wipe(); err != nil {
		return errorResponse(errors.Wrap(err, "failed to override officers.json"))
	}

	if err := json.NewEncoder(f).Encode(officers); err != nil {
		return errorResponse(errors.Wrap(err, "failed to encode officers.json"))
	}

	commitHash, err := repo.Commit(commit.Title, commit.Body)
	if err != nil {
		return errorResponse(errors.Wrap(err, "failed to commit"))
	}

	if err := f.Close(); err != nil {
		return errorResponse(errors.Wrap(err, "failed to close officers.json"))
	}

	shortHash := commitHash.String()[:7]
	return &api.InteractionResponseData{
		Content: option.NewNullableString(fmt.Sprintf(
			"`[%s]` %s\n\n%s",
			shortHash, commit.Title, commit.Body,
		)),
	}
}

func (h *Handler) forUser(command cmdroute.CommandData, forUser discord.UserID) (*discord.Member, error) {
	if (!forUser.IsValid() || forUser == command.Event.User.ID) && command.Event.Member != nil {
		return command.Event.Member, nil
	}

	member, err := h.state.Member(command.Event.GuildID, forUser)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get requested user")
	}

	return member, nil
}

func (h *Handler) handleLink(ctx context.Context, command cmdroute.CommandData) *api.InteractionResponseData {
	var data struct {
		ForUser  discord.UserID `discord:"for_user?"`
		FullName string         `discord:"full_name"`
	}

	if err := command.Options.Unmarshal(&data); err != nil {
		return errorResponse(err)
	}

	member, err := h.forUser(command, data.ForUser)
	if err != nil {
		return errorResponse(err)
	}

	return h.updateOfficers(ctx, command, func(officers *acmcsuf.Officers) (commit, error) {
		if officer := officers.Find(func(officer *acmcsuf.Officer) bool {
			return officer.FullName == data.FullName
		}); officer != nil {
			officer.Socials.Discord = member.User.Tag()
		} else {
			*officers = append(*officers, acmcsuf.Officer{
				FullName: data.FullName,
				Socials: acmcsuf.Socials{
					Discord: member.User.Tag(),
				},
			})
		}

		return commit{
			Title: fmt.Sprintf("Update officer %s", data.FullName),
			Body:  fmt.Sprintf("Update officer %s's Discord tag to %q.", data.FullName, member.User.Tag()),
		}, nil
	})
}

func (h *Handler) handleSet(ctx context.Context, command cmdroute.CommandData) *api.InteractionResponseData {
	var data struct {
		ForUser   discord.UserID `discord:"for_user?"`
		GitHub    string         `discord:"github?"`
		LinkedIn  string         `discord:"linkedin?"`
		Instagram string         `discord:"instagram?"`
		Website   string         `discord:"website?"`
	}

	if err := command.Options.Unmarshal(&data); err != nil {
		return errorResponse(err)
	}

	member, err := h.forUser(command, data.ForUser)
	if err != nil {
		return errorResponse(err)
	}

	return h.updateOfficers(ctx, command, func(officers *acmcsuf.Officers) (commit, error) {
		officer := officers.Find(func(officer *acmcsuf.Officer) bool {
			return officer.Socials.Discord == member.User.Tag()
		})

		if officer == nil {
			return commit{}, errors.New("officer not found (have you done /officer link?)")
		}

		var updated []string
		if data.GitHub != "" {
			officer.Socials.GitHub = data.GitHub
			updated = append(updated, "GitHub")
		}
		if data.LinkedIn != "" {
			officer.Socials.LinkedIn = data.LinkedIn
			updated = append(updated, "LinkedIn")
		}
		if data.Instagram != "" {
			officer.Socials.Instagram = data.Instagram
			updated = append(updated, "Instagram")
		}
		if data.Website != "" {
			officer.Socials.Website = data.Website
			updated = append(updated, "Website")
		}

		return commit{
			Title: fmt.Sprintf("Update officer %s", member.User.Username),
			Body: fmt.Sprintf(
				"Update officer %s's socials (%s).",
				member.User.Username, strings.Join(updated, ", "),
			),
		}, nil
	})
	panic("TODO")
}

func (h *Handler) handleAddTerm(ctx context.Context, command cmdroute.CommandData) *api.InteractionResponseData {
	panic("TODO")
}

func (h *Handler) handlePR(ctx context.Context, command cmdroute.CommandData) *api.InteractionResponseData {
	panic("TODO")
}
