package main

import "github.com/bwmarrin/discordgo"

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Ping!",
		},
		{
			Name:        "settings",
			Description: "Settings for Pinguino",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "show",
					Description: "Show current settings",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set-quotes-channel",
					Description: "Set the channel to send quotes to",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "channel",
							Description: "The channel to send logs to",
							Required:    true,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildText,
							},
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set-logs-channel",
					Description: "Set the channel to send logs to",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "channel",
							Description: "The channel to send logs to",
							Required:    true,
							ChannelTypes: []discordgo.ChannelType{
								discordgo.ChannelTypeGuildText,
							},
						},
					},
				},
			},
		},
		{
			Name:        "quote",
			Description: "Record a quote",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "discord",
					Description: "Record a quote said by someone on Discord",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "quote",
							Description: "The quote",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "author",
							Description: "The user who said the quote",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "external",
					Description: "Record a quote said by someone not on the Discord",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "quote",
							Description: "The quote",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "author",
							Description: "The person who said the quote",
							Required:    true,
						},
					},
				},
			},
		},
		{
			Name:        "quotes",
			Description: "Commands to browse quotes",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "all",
					Description: "Browse all quotes",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "by-external",
					Description: "Browse quotes by someone not on the Discord",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "author",
							Description: "The author to search by",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "by-discord",
					Description: "Browse quotes by someone on the Discord",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "author",
							Description: "The author to search by",
							Required:    true,
						},
					},
				},
			},
		},
	}

	handlers = map[string]func(*discordgo.Session, *discordgo.InteractionCreate, *Context){
		"ping":     HandlePing,
		"settings": HandleSettings,
		"quote":    HandleQuote,
		"quotes":   HandleQuotes,
	}
)
