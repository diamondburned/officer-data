package bot

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

var commands = []api.CreateCommandData{
	{
		Name:        "officer",
		Description: "Obtain or modify information about an officer.",
		// /officer link name:"Diamond"              // match name, add a Discord username
		// /officer set instagram="<instagram name>" // set the instagram name
		// /officer pr                               // commit to a new PR or update an existing PR
		Options: []discord.CommandOption{
			&discord.SubcommandOption{
				OptionName: "link",
				Description: "Link a Discord user to an officer. " +
					"The officer that's linked to will have their Discord tag updated.",
				Options: []discord.CommandOptionValue{
					&discord.UserOption{
						OptionName: "for_user",
						Description: "The Discord user to link with. " +
							"If not specified, then the current user is used.",
					},
					&discord.StringOption{
						OptionName:  "full_name",
						Description: "Your full name or the other user's full name.",
						Required:    true,
					},
				},
			},
			&discord.SubcommandOption{
				OptionName:  "set",
				Description: "Set a field for an officer.",
				Options: []discord.CommandOptionValue{
					&discord.UserOption{
						OptionName: "for_user",
						Description: "The Discord user to link with. " +
							"If not specified, then the current user is used.",
					},
					&discord.StringOption{
						OptionName:  "github",
						Description: "The GitHub username of the officer.",
					},
					&discord.StringOption{
						OptionName:  "linkedin",
						Description: "The LinkedIn username of the officer.",
					},
					&discord.StringOption{
						OptionName:  "instagram",
						Description: "The Instagram username of the officer.",
					},
					&discord.StringOption{
						OptionName:  "website",
						Description: "The website of the officer.",
					},
				},
			},
			&discord.SubcommandOption{
				OptionName:  "add-term",
				Description: "Add a new term of an officer.",
				Options: []discord.CommandOptionValue{
					&discord.UserOption{
						OptionName: "for_user",
					},
					&discord.StringOption{
						OptionName: "semester",
						Description: "The semester of the term. If not specified, then the " +
							"current semester is used.",
						Choices: []discord.StringChoice{
							{Value: "F", Name: "Fall"},
							{Value: "S", Name: "Spring"},
						},
					},
					&discord.NumberOption{
						OptionName: "year",
						Description: "The year of the term. If not specified, then the " +
							"current year is used.",
					},
				},
			},
			&discord.SubcommandOption{
				OptionName: "pr",
				Description: "Create a new PR or update the existing one containing " +
					"all the changes made previously. One user can have one ongoing PR.",
			},
		},
	},
}
