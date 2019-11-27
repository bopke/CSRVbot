package giveaways

import "log"

func BlacklistUser(guildID, userID, blacklisterID string) error {
	blacklist := &Blacklist{GuildId: guildID,
		UserId:        userID,
		BlacklisterId: blacklisterID}
	err := database.DbMap.Insert(blacklist)
	if err != nil {
		log.Panicln("BlacklistUser DbMap.Isert(blacklist)", err)
	}
	return err
}

func UnblacklistUser(guildID, userID string) error {
	_, err := database.DbMap.Exec("DELETE FROM Blacklists WHERE guild_id = ? AND user_id = ?", guildID, userID)
	if err != nil {
		log.Panicln("UnblacklistUser DbMap.Exec", err)
	}
	return err
}

func IsBlacklisted(guildID, userID string) bool {
	ret, err := database.DbMap.SelectInt("SELECT count(*) FROM Blacklists WHERE guild_id = ? AND user_id = ?", guildID, userID)
	if err != nil {
		log.Panicln("IsBlacklisted DbMap.SelectInt ", err)
	}
	if ret == 1 {
		return true
	}
	return false
}
