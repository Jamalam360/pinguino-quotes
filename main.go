package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v3"
)

type Context struct {
	Config   *Config
	Database *sql.DB
}

type Config struct {
	Token    string
	GuildId  string `yaml:"guild_id"`
	Database struct {
		DatabasePath string `yaml:"database_path"`
	}
}

func ReadConfig() Config {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		configPath = "config.yaml"
	}

	log.Printf("Reading config from %s", configPath)
	configText, err := os.ReadFile(configPath)
	CheckFatal(err)

	config := Config{}
	if err := yaml.Unmarshal(configText, &config); err != nil {
		CheckFatal(err)
	}

	config.Token = os.ExpandEnv(config.Token)
	config.GuildId = os.ExpandEnv(config.GuildId)
	config.Database.DatabasePath = os.ExpandEnv(config.Database.DatabasePath)

	if strings.Contains(config.Token, "$") {
		log.Printf("Token might be suspicious: %s", config.Token);
	}

	return config
}

func main() {
	config := ReadConfig()
	database := ConnectDatabase(config)
	SetupDatabase(database)
	ctx := Context{
		Config:   &config,
		Database: database,
	}
	bot, err := discordgo.New("Bot " + config.Token)
	CheckFatal(err)

	bot.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			log.Printf("Interaction received: %s", interaction.ApplicationCommandData().Name)
			if handler, ok := handlers[interaction.ApplicationCommandData().Name]; ok {
				handler(session, interaction, &ctx)
			}
		case discordgo.InteractionMessageComponent:
			log.Printf("Interaction received: %s", interaction.MessageComponentData().CustomID)
			HandlePaginationButton(session, interaction, &ctx)
		}

	})

	bot.AddHandler(func(session *discordgo.Session, ready *discordgo.Ready) {
		log.Printf("Received READY event")
	})

	CheckFatal(bot.Open())
	log.Printf("Registering commands")

	for idx, command := range commands {
		_, err := bot.ApplicationCommandCreate(bot.State.User.ID, config.GuildId, command)
		CheckFatal(err)
		log.Printf("Command %d (%s) registered", idx, command.Name)
	}

	defer bot.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
