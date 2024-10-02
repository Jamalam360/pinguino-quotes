package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func HandlePing(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx *Context) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})
}

func HandleSettings(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx *Context) {
	options := interaction.ApplicationCommandData().Options
	text := ""

	switch options[0].Name {
	case "show":
		settings, err := GetGuildSettings(interaction.GuildID, ctx)

		if err == sql.ErrNoRows {
			text = "Pinguino has not been set up yet"
			break
		} else if CheckOrRespond(err, interaction, session) {
			return
		}

		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title: "Guild Settings",
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:  "Guild ID",
								Value: interaction.GuildID,
							},
							{
								Name:  "Quotes Channel",
								Value: ChannelAsMention(settings.QuotesChannelId),
							},
							{
								Name:  "Logs Channel",
								Value: ChannelAsMention(settings.LogsChannelId),
							},
						},
					},
				},
			},
		})

		return
	case "set-quotes-channel":
		settings, err := GetGuildSettings(interaction.GuildID, ctx)
		if err == sql.ErrNoRows {
			guildId, err := strconv.ParseInt(interaction.GuildID, 10, 64)
			if CheckOrRespond(err, interaction, session) {
				return
			}

			settings = GuildSettings{
				GuildID: guildId,
			}
		} else if CheckOrRespond(err, interaction, session) {
			return
		}

		channel := interaction.ApplicationCommandData().Options[0].Options[0].ChannelValue(session)
		id, err := strconv.ParseInt(channel.ID, 10, 64)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		settings.QuotesChannelId = id
		err = UpdateGuildSettings(settings, ctx)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		text = fmt.Sprintf("Quotes channel set to %s", ChannelAsMention(id))
		SendLog("Settings Updated - Quotes Channel", []*discordgo.MessageEmbedField{
			{
				Name:  "New Quotes Channel",
				Value: ChannelAsMention(id),
			},
			{
				Name: "Actor",
				Value: UserAsMention(interaction.Member.User.ID),
			},
		}, session, interaction, ctx)
	case "set-logs-channel":
		settings, err := GetGuildSettings(interaction.GuildID, ctx)
		if err == sql.ErrNoRows {
			guildId, err := strconv.ParseInt(interaction.GuildID, 10, 64)
			if CheckOrRespond(err, interaction, session) {
				return
			}

			settings = GuildSettings{
				GuildID: guildId,
			}
		} else if CheckOrRespond(err, interaction, session) {
			return
		}

		channel := interaction.ApplicationCommandData().Options[0].Options[0].ChannelValue(session)
		id, err := strconv.ParseInt(channel.ID, 10, 64)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		settings.LogsChannelId = id
		err = UpdateGuildSettings(settings, ctx)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		text = fmt.Sprintf("Logs channel set to %s", ChannelAsMention(id))
		SendLog("Settings Updated - Logs Channel", []*discordgo.MessageEmbedField{
			{
				Name:  "New Logs Channel",
				Value: ChannelAsMention(id),
			},
			{
				Name: "Actor",
				Value: UserAsMention(interaction.Member.User.ID),
			},
		}, session, interaction, ctx)
	}

	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	})
}

func HandleQuote(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx *Context) {
	options := interaction.ApplicationCommandData().Options
	text := ""
	switch options[0].Name {
	case "discord":
		quote := options[0].Options[0].StringValue()
		author := options[0].Options[1].UserValue(session)
		err := SendQuote(author.Username, author.AvatarURL(""), quote, session, interaction, ctx)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		userId, err := strconv.ParseInt(author.ID, 10, 64)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		guildId, err := strconv.ParseInt(interaction.GuildID, 10, 64)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		err = InsertQuote(Quote{
			GuildID:       guildId,
			Author:        author.Username,
			Quote:         quote,
			DiscordUserId: userId,
		}, ctx)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		text = "Quote recorded"
	case "external":
		quote := options[0].Options[0].StringValue()
		author := options[0].Options[1].StringValue()
		err := SendQuote(author, "", quote, session, interaction, ctx)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		guildId, err := strconv.ParseInt(interaction.GuildID, 10, 64)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		err = InsertQuote(Quote{
			GuildID:       guildId,
			Author:        author,
			Quote:         quote,
			DiscordUserId: 0,
		}, ctx)
		if CheckOrRespond(err, interaction, session) {
			return
		}

		text = "Quote recorded"
	}

	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	})
}

func SendQuote(author string, iconUrl string, quote string, session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx *Context) error {
	settings := GuildSettings{}
	err := ctx.Database.QueryRow(`SELECT guild_id, quotes_channel_id FROM guild_settings WHERE guild_id = $1`, interaction.GuildID).Scan(&settings.GuildID, &settings.QuotesChannelId)
	if err != nil {
		return err
	}

	if settings.QuotesChannelId == 0 {
		return fmt.Errorf("pinguino has not been set up yet")
	}

	channel, err := session.Channel(fmt.Sprintf("%v", settings.QuotesChannelId))
	if err != nil {
		return err
	}

	_, err = session.ChannelMessageSendEmbed(channel.ID, &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    author,
			IconURL: iconUrl,
		},
		Title: quote,
	})
	if err != nil {
		return err
	}

	SendLog("Quote Sent", []*discordgo.MessageEmbedField{
		{
			Name:  "Quote",
			Value: quote,
		},
		{
			Name:  "Author",
			Value: author,
		},
		{
			Name: "Actor",
			Value: UserAsMention(interaction.Member.User.ID),
		},
	}, session, interaction, ctx)

	return nil
}

func HandleQuotes(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx *Context) {
	options := interaction.ApplicationCommandData().Options
	quotes := []Quote{}
	switch options[0].Name {
	case "all":
		var err error
		quotes, err = GetQuotes(interaction.GuildID, ctx)
		if CheckOrRespond(err, interaction, session) {
			return
		}
	case "by-external":
		author := options[0].Options[0].StringValue()
		var err error
		quotes, err = GetQuotesByAuthor(interaction.GuildID, author, ctx)
		if CheckOrRespond(err, interaction, session) {
			return
		}
	case "by-discord":
		author := options[0].Options[0].UserValue(session)
		var err error
		quotes, err = GetQuotesByDiscordUserId(interaction.GuildID, author.ID, ctx)
		if CheckOrRespond(err, interaction, session) {
			return
		}
	}

	if len(quotes) == 0 {
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No quotes found",
			},
		})
	} else {
		SendPaginatedQuotes(quotes, session, interaction, ctx)
	}
}

type PaginationState struct {
	Quotes []Quote
	PageIndex int
}

var paginationStates = map[string]PaginationState{}

func SendPaginatedQuotes(quotes []Quote, session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx *Context) {
	paginationStates[interaction.ID] = PaginationState{
		Quotes: quotes,
		PageIndex: 0,
	}
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Components: CreatePaginationButtons(interaction.ID),
			Embeds: []*discordgo.MessageEmbed{
				CreatePaginationEmbed(quotes, 0),
			},
		},
	})
}

func HandlePaginationButton(session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx *Context) {
	data := interaction.MessageComponentData()
	parts := strings.SplitN(data.CustomID, "_", 3)
	if len(parts) != 3 {
		return
	}

	action := parts[1]
	interactionId := parts[2]

	state, ok := paginationStates[interactionId]
	if !ok {
		return
	}

	switch action {
	case "previous":
		if state.PageIndex > 0 {
			state.PageIndex--
		}
	case "next":
		if state.PageIndex < len(state.Quotes) / 10 {
			state.PageIndex++
		}
	}

	paginationStates[interactionId] = state
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Components: CreatePaginationButtons(interactionId),
			Embeds: []*discordgo.MessageEmbed{
				CreatePaginationEmbed(state.Quotes, state.PageIndex),
			},
		},
	})
}

func CreatePaginationButtons(interactionId string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label: 	"Previous",
					Style: 	discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("quotes_previous_%s", interactionId),
				},
				discordgo.Button{
					Label: 	"Next",
					Style: 	discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("quotes_next_%s", interactionId),
				},
			},
		},
	}
}

func CreatePaginationEmbed(quotes []Quote, pageIndex int) *discordgo.MessageEmbed {
	pageSize := 10
	start := pageIndex * pageSize
	end := (pageIndex + 1) * pageSize
	if end > len(quotes) {
		end = len(quotes)
	}

	fields := []*discordgo.MessageEmbedField{}
	for i := start; i < end; i++ {
		quote := quotes[i]
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: quote.Author,
			Value: quote.Quote,
		})
	}

	return &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Quotes (%d-%d of %d)", start + 1, end, len(quotes)),
		Fields: fields,
	}
}

func SendLog(action string, fields []*discordgo.MessageEmbedField, session *discordgo.Session, interaction *discordgo.InteractionCreate, ctx *Context) {
	settings := GuildSettings{}
	err := ctx.Database.QueryRow(`SELECT guild_id, logs_channel_id FROM guild_settings WHERE guild_id = $1`, interaction.GuildID).Scan(&settings.GuildID, &settings.LogsChannelId)
	if err != nil {
		return
	}

	if settings.LogsChannelId == 0 {
		return
	}

	channel, err := session.Channel(fmt.Sprintf("%v", settings.LogsChannelId))
	if err != nil {
		return
	}

	_, err = session.ChannelMessageSendEmbed(channel.ID, &discordgo.MessageEmbed{
		Title: action,
		Fields: fields,
	})
	if err != nil {
		return
	}
}