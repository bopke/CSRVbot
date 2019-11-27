package giveaways

import (
	"csrvbot/serverConfiguration"
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"time"
)

func GetGiveawayForGuild(guildId string) *Giveaway {
	var giveaway Giveaway
	err := database.DbMap.SelectOne(&giveaway, "SELECT * FROM Giveaways WHERE guild_id = ? AND end_time IS NULL", guildId)
	if err != nil && err != sql.ErrNoRows {
		log.Panicln("getGiveawayForGuild DbMap.SelectOne ", err)
	}
	if err == sql.ErrNoRows {
		return nil
	}
	return &giveaway
}

func GetAllUnfinishedGiveaways() []Giveaway {
	var res []Giveaway
	_, err := database.DbMap.Select(&res, "SELECT * FROM Giveaways WHERE end_time IS NULL")
	if err != nil {
		log.Panicln("GetAllUnfinishedGiveaways DbMap.Select ", err)
		return nil
	}
	return res
}

func CreateMissingGiveaways(session *discordgo.Session) {
	for i := 0; i < len(session.State.Guilds); i++ {
		guild, _ := session.Guild(session.State.Guilds[i].ID)
		for _, channel := range guild.Channels {
			if channel.Name == GetGiveawayChannelNameForGuild(guild.ID) {
				giveaway := GetGiveawayForGuild(guild.ID)
				if giveaway == nil {
					giveaway = &Giveaway{
						StartTime: time.Now(),
						GuildId:   guild.ID,
						GuildName: guild.Name,
					}
					err := database.DbMap.Insert(giveaway)
					if err != nil {
						log.Panicln("CreateMissingGiveaways DbMap.Insert ", err)
					}
				}
				break
			}
		}
	}
}

func GetGiveawayChannelNameForGuild(guildID string) string {
	var serverConfig serverConfiguration.ServerConfig
	err := database.DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig")
	if err != nil {
		log.Println("GetGiveawayChannelNameForGuild("+guildID+") ", err)
		return ""
	}
	return serverConfig.MainChannel
}

func FinishGiveaway(session *discordgo.Session, guildID string) {
	giveaway := GetGiveawayForGuild(guildID)
	guild, err := session.Guild(giveaway.GuildId)
	if err != nil {
		log.Println("Nie mogę się dobrać do gildii o ID " + guildID + ", pomijam.")
		return
	}
	var giveawayChannelId string
	for _, channel := range guild.Channels {
		if channel.Name == GetGiveawayChannelNameForGuild(guildID) {
			giveawayChannelId = channel.ID
			break
		}
	}
	var participants []Participant
	_, err = database.DbMap.Select(&participants, "SELECT * FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveaway.Id)
	if err != nil {
		log.Panicln("FinishGiveaway DbMap.Select " + err.Error())
	}
	if participants == nil || len(participants) == 0 {
		giveaway.EndTime.Time = time.Now()
		giveaway.EndTime.Valid = true
		_, err := database.DbMap.Update(giveaway)
		if err != nil {
			log.Panicln("FinishGiveaway DbMap.Select ", err)
		}
		notifyWinner(session, giveaway.GuildId, giveawayChannelId, nil, "")
		return
	}
	code, err := utils.GetCSRVCode()
	if err != nil {
		log.Println("FinishGiveaway getCSRVCode ", err)
		_, _ = session.ChannelMessageSend(giveawayChannelId, "Błąd API Craftserve, nie udało się pobrać kodu!")
		return
	}
	rand.Seed(time.Now().UnixNano())
	winner := participants[rand.Int()%len(participants)]
	giveaway.InfoMessageId.String = notifyWinner(session, giveaway.GuildId, giveawayChannelId, &winner.UserId, code)
	giveaway.InfoMessageId.Valid = true
	giveaway.EndTime.Time = time.Now()
	giveaway.EndTime.Valid = true
	giveaway.Code.String = code
	giveaway.Code.Valid = true
	giveaway.WinnerId.String = winner.UserId
	giveaway.WinnerId.Valid = true
	giveaway.WinnerName.String = winner.UserName
	giveaway.WinnerName.Valid = true
	_, err = database.DbMap.Update(giveaway)
	if err != nil {
		log.Panicln("FinishGiveaway DbMap.Update ", err)
	}
}
