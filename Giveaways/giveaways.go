package Giveaways

import (
	"csrvbot/Database"
	"csrvbot/ServerConfiguration"
	"csrvbot/Utils"
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"time"
)

func GetGiveawayForGuild(guildId string) *Giveaway {
	var giveaway Giveaway
	err := Database.DbMap.SelectOne(&giveaway, "SELECT * FROM Giveaways WHERE guild_id = ? AND end_time IS NULL", guildId)
	if err != nil && err != sql.ErrNoRows {
		log.Panicln("Giveaways GetGiveawayForGuild Unable to select from database! ", err)
	}
	if err == sql.ErrNoRows {
		return nil
	}
	return &giveaway
}

func GetAllUnfinishedGiveaways() []Giveaway {
	var res []Giveaway
	_, err := Database.DbMap.Select(&res, "SELECT * FROM Giveaways WHERE end_time IS NULL")
	if err != nil {
		log.Panicln("Giveaways GetAllUnfinishedGiveaways Unable to select from database! ", err)
		return nil
	}
	return res
}

func CreateMissingGiveaways(session *discordgo.Session) {
	for i := 0; i < len(session.State.Guilds); i++ {
		guild, err := session.Guild(session.State.Guilds[i].ID)
		if err != nil {
			log.Println("Giveaways CreateMissingGiveaways Unable to get guild! ", err)
		}
		for _, channel := range guild.Channels {
			if channel.Name == GetGiveawayChannelNameForGuild(guild.ID) {
				giveaway := GetGiveawayForGuild(guild.ID)
				if giveaway == nil {
					giveaway = &Giveaway{
						StartTime: time.Now(),
						GuildId:   guild.ID,
						GuildName: guild.Name,
					}
					err := Database.DbMap.Insert(giveaway)
					if err != nil {
						log.Panicln("Giveaways CreateMissingGiveaways Unable to insert to database! ", err)
					}
				}
				break
			}
		}
	}
}

func GetGiveawayChannelNameForGuild(guildID string) string {
	var serverConfig ServerConfiguration.ServerConfig
	err := Database.DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig WHERE guild_id = ?", guildID)
	if err != nil {
		log.Println("Giveaways GetGiveawaysChannelNameForGuild Unable to select from database! ", err)
		return ""
	}
	return serverConfig.MainChannel
}

func FinishGiveaway(session *discordgo.Session, guildID string) {
	giveaway := GetGiveawayForGuild(guildID)
	guild, err := session.Guild(giveaway.GuildId)
	if err != nil {
		log.Println("Giveaways FinishGiveaway Unable to get guild! ", err)
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
	_, err = Database.DbMap.Select(&participants, "SELECT * FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveaway.Id)
	if err != nil {
		log.Panicln("Giveaways FinishGiveaway Unable to select from database! ", err)
	}
	if participants == nil || len(participants) == 0 {
		giveaway.EndTime.Time = time.Now()
		giveaway.EndTime.Valid = true
		_, err := Database.DbMap.Update(giveaway)
		if err != nil {
			log.Panicln("Giveaways FinishGiveaway Unable to select from database ", err)
		}
		notifyWinner(session, giveaway.GuildId, giveawayChannelId, nil, "")
		return
	}
	code, err := Utils.GetCSRVCode()
	if err != nil {
		log.Println("Giveaways FinishGiveaway Unable to get CSRV code! ", err)
		_, err = session.ChannelMessageSend(giveawayChannelId, "Błąd API Craftserve, nie udało się pobrać kodu! <@205745502266851329> ")
		if err != nil {
			log.Println("Giveaways FinishGiveaway Unable to send channel message! ", err)
		}
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
	_, err = Database.DbMap.Update(giveaway)
	if err != nil {
		log.Panicln("Giveaways FinishGiveaway Unable to update in database! ", err)
	}
}
