package ServerConfiguration

import (
	"csrvbot/Database"
	"csrvbot/Models"
	"database/sql"
	"log"
)

func CreateConfigurationIfNotExists(guildID string) {
	var serverConfig Models.ServerConfig
	err := Database.DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig WHERE guild_id = ?", guildID)
	if err == sql.ErrNoRows {
		serverConfig.GuildId = guildID
		serverConfig.MainChannel = "giveaway"
		serverConfig.AdminRole = "CraftserveBotAdmin"
		err = Database.DbMap.Insert(&serverConfig)
		if err != nil {
			log.Panicln("ServerConfiguration CreateConfigurationIfNotExists Unable to insert to database! ", err)
		}
	}
	if err != nil {
		log.Panicln("ServerConfiguration CreateConfigurationIfNotExists Unable to select from database! ", err)
	}
}

func GetServerConfigForGuildId(guildID string) (serverConfig Models.ServerConfig) {
	err := Database.DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig WHERE guild_id = ?", guildID)
	if err != nil {
		log.Panicln("ServerConfiguration CreateConfigurationIfNotExists Unable to select from database! ", err)
	}
	return
}
