package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type GuildSettings struct {
	GuildID         int64
	QuotesChannelId int64
	LogsChannelId   int64
}

type Quote struct {
	GuildID       int64
	Author        string
	Quote         string
	DiscordUserId int64
}

func ConnectDatabase(config Config) *sql.DB {
	db, err := sql.Open("postgres", config.Database.ConnectionString)
	CheckFatal(err)
	log.Println("Database connected")
	return db
}

func SetupDatabase(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS guild_settings (
		guild_id BIGINT NOT NULL PRIMARY KEY,
		quotes_channel_id BIGINT,
		logs_channel_id BIGINT
	)`)
	CheckFatal(err)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS quotes (
		guild_id BIGINT NOT NULL,
		author VARCHAR(64) NOT NULL,
		quote TEXT NOT NULL,
		discord_user_id BIGINT NOT NULL,
		PRIMARY KEY (author, quote)
	)`)
	CheckFatal(err)
	log.Println("Database setup")
}

func GetQuotes(guildId string, ctx *Context) ([]Quote, error) {
	rows, err := ctx.Database.Query(`SELECT * FROM quotes WHERE guild_id = $1`, guildId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotes := []Quote{}
	for rows.Next() {
		quote := Quote{}
		err = rows.Scan(&quote.GuildID, &quote.Author, &quote.Quote, &quote.DiscordUserId)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

func GetQuotesByAuthor(guildId string, author string, ctx *Context) ([]Quote, error) {
	rows, err := ctx.Database.Query(`SELECT * FROM quotes WHERE guild_id = $1 AND author = $2`, guildId, author)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotes := []Quote{}
	for rows.Next() {
		quote := Quote{}
		err = rows.Scan(&quote.GuildID, &quote.Author, &quote.Quote, &quote.DiscordUserId)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

func GetQuotesByDiscordUserId(guildId string, discordUserId string, ctx *Context) ([]Quote, error) {
	rows, err := ctx.Database.Query(`SELECT * FROM quotes WHERE guild_id = $1 AND discord_user_id = $2`, guildId, discordUserId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotes := []Quote{}
	for rows.Next() {
		quote := Quote{}
		err = rows.Scan(&quote.GuildID, &quote.Author, &quote.Quote, &quote.DiscordUserId)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

func InsertQuote(quote Quote, ctx *Context) error {
	_, err := ctx.Database.Exec(`INSERT INTO quotes (guild_id, author, quote, discord_user_id) VALUES ($1, $2, $3, $4)`, quote.GuildID, quote.Author, quote.Quote, quote.DiscordUserId)
	log.Printf("Inserted quote: %+v", quote)
	return err
}

func GetGuildSettings(guildId string, ctx *Context) (GuildSettings, error) {
	var settings GuildSettings
	err := ctx.Database.QueryRow(`SELECT * FROM guild_settings WHERE guild_id = $1`, guildId).Scan(&settings.GuildID, &settings.QuotesChannelId, &settings.LogsChannelId)
	return settings, err	
}

func UpdateGuildSettings(settings GuildSettings, ctx *Context) error {
	_, err := ctx.Database.Exec(`INSERT INTO guild_settings (guild_id, quotes_channel_id, logs_channel_id) VALUES ($1, $2, $3) ON CONFLICT (guild_id) DO UPDATE SET quotes_channel_id = $2, logs_channel_id = $3`, settings.GuildID, settings.QuotesChannelId, settings.LogsChannelId)
	log.Printf("Updated settings for guild %d: %+v", settings.GuildID, settings)
	return err
}
