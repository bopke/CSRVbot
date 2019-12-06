package Giveaways

import (
	"csrvbot/Database"
	"csrvbot/Models"
	"log"
)

func BlacklistUser(guildID, userID, blacklisterID string) {
	blacklist := &Models.Blacklist{GuildId: guildID,
		UserId:        userID,
		BlacklisterId: blacklisterID}
	err := Database.DbMap.Insert(blacklist)
	if err != nil {
		log.Panicln("Giveaways BlacklistUser Unable to insert to database! ", err)
	}
	return
}

func UnblacklistUser(guildID, userID string) {
	_, err := Database.DbMap.Exec("DELETE FROM Blacklists WHERE guild_id = ? AND user_id = ?", guildID, userID)
	if err != nil {
		log.Panicln("Giveaways UnblacklistUser Unable to delete from database! ", err)
	}
	return
}

func IsBlacklisted(guildID, userID string) bool {
	ret, err := Database.DbMap.SelectInt("SELECT count(*) FROM Blacklists WHERE guild_id = ? AND user_id = ?", guildID, userID)
	if err != nil {
		log.Panicln("Giveaways IsBlacklisted Unable to select from database! ", err)
	}
	if ret == 1 {
		return true
	}
	return false
}
