package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

func ChannelAsMention(channelId int64) string {
	if channelId == 0 {
		return "None"
	}

	return fmt.Sprintf("<#%d>", channelId)
}

func UserAsMention(userId string) string {
	if userId == "" {
		return "None"
	}

	return fmt.Sprintf("<@%s>", userId)
}

func CheckFatal(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckOrRespond(err error, interaction *discordgo.InteractionCreate, session *discordgo.Session) bool {
	if err != nil {
		log.Printf("Encountered an error processing interaction %s (%s): %s", interaction.ID, interaction.Interaction.ApplicationCommandData().Name, err)
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "An error occurred while processing this interaction",
			},
		})
	}

	return err != nil
}
