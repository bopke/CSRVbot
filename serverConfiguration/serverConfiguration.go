package serverConfiguration

import (
	"database/sql"
	"log"
)

type ServerConfig struct {
	Id          int    `db:"id,primarykey,autoincrement"`
	GuildId     string `db:"guild_id,size:255"`
	AdminRole   string `db:"admin_role,size:255"`
	MainChannel string `db:"main_channel,size:255"`
}

func createConfigurationIfNotExists(guildID string) {
	var serverConfig ServerConfig
	err := database.DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig WHERE guild_id=?", guildID)
	if err == sql.ErrNoRows {
		serverConfig.GuildId = guildID
		serverConfig.MainChannel = "giveaway"
		serverConfig.AdminRole = "CraftserveBotAdmin"
		err = database.DbMap.Insert(&serverConfig)
	}
	if err != nil {
		log.Panicln("createConfigurationIfNotExists DbMap.SelectOne ", err)
	}
}

func getServerConfigForGuildId(guildID string) (serverConfig ServerConfig) {
	err := database.DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig WHERE guild_id=?", serverConfig)
	if err != nil {
		log.Panicln("createConfigurationIfNotExists DbMap.SelectOne ", err)
	}
	return
}
