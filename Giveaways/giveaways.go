package Giveaways

import (
	"csrvbot/Config"
	"csrvbot/Database"
	"csrvbot/Models"
	"csrvbot/Utils"
	"database/sql"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetGiveawayForGuild(guildId string) *Models.Giveaway {
	var giveaway Models.Giveaway
	err := Database.DbMap.SelectOne(&giveaway, "SELECT * FROM Giveaways WHERE guild_id = ? AND end_time IS NULL", guildId)
	if err != nil && err != sql.ErrNoRows {
		log.Panicln("Giveaways GetGiveawayForGuild Unable to select from database! ", err)
	}
	if err == sql.ErrNoRows {
		return nil
	}
	return &giveaway
}

func GetAllUnfinishedGiveaways() []Models.Giveaway {
	var res []Models.Giveaway
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
					giveaway = &Models.Giveaway{
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
	var serverConfig Models.ServerConfig
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
	var participants []Models.Participant
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

func PrintGiveawayInfo(session *discordgo.Session, channelID, guildID string) *discordgo.Message {
	splittedCronString := strings.Split(Config.GiveawayCronString, " ")
	giveawayTimeString := splittedCronString[1] + ":" + splittedCronString[2]
	info := "**Ten bot organizuje giveaway kodów na serwery Diamond.**\n" +
		"**Każdy kod przedłuża serwer o 7 dni.**\n" +
		"Aby wziąć udział pomagaj innym użytkownikom. Jeżeli komuś pomożesz, to poproś tą osobę aby napisala `!thx @TwojNick` - w ten sposób dostaniesz się do loterii. To jest nasza metoda na rozruszanie tego Discorda, tak, aby każdy mógł liczyć na pomoc. Każde podziękowanie to jeden los, więc warto pomagać!\n\n" +
		"**Sponsorem tego bota jest https://craftserve.pl/ - hosting serwerów Minecraft.**\n\n" +
		"Pomoc musi odbywać się na tym serwerze na tekstowych kanałach publicznych.\n\n" +
		"Uczestnicy: " +
		GetParticipantsNamesString(GetGiveawayForGuild(guildID).Id) +
		"\n\nNagrody rozdajemy o " + giveawayTimeString + ", Powodzenia!"
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Informacje o Giveawayu",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: info,
		Color:       0x234d20,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	m, err := session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Println("Utils PrintGiveawayInfo Unable to send channel message embed ", err)
	}
	return m
}
