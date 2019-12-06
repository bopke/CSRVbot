package ServerConfiguration

import (
	"csrvbot/Database"
	"database/sql"
	"log"
)

type ServerConfig struct {
	Id          int    `db:"id,primarykey,autoincrement"`
	GuildId     string `db:"guild_id,size:255"`
	AdminRole   string `db:"admin_role,size:255"`
	MainChannel string `db:"main_channel,size:255"`
}

func CreateConfigurationIfNotExists(guildID string) {
	var serverConfig ServerConfig
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

func GetServerConfigForGuildId(guildID string) (serverConfig ServerConfig) {
	err := Database.DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig WHERE guild_id = ?", guildID)
	if err != nil {
		log.Panicln("ServerConfiguration CreateConfigurationIfNotExists Unable to select from database! ", err)
	}
	return
}
