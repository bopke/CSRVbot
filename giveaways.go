package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func getGiveawayForGuild(guildId string) *Giveaway {
	var giveaway Giveaway
	err := DbMap.SelectOne(&giveaway, "SELECT * FROM Giveaways WHERE guild_id = ? AND end_time IS NULL", guildId)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &giveaway
}

func getAllUnfinishedGiveaways() []Giveaway {
	var res []Giveaway
	_, err := DbMap.Select(&res, "SELECT * FROM giveaways WHERE end_time IS NULL")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return res
}

func createMissingGiveaways() {
	for i := 0; i < len(session.State.Guilds); i++ {
		// Jak się tak dziwnie nie wyciągnie gildii to nie działa
		guild, _ := session.Guild(session.State.Guilds[i].ID)
		for _, channel := range guild.Channels {
			if channel.Name == config.MainChannel {
				giveaway := getGiveawayForGuild(guild.ID)
				if giveaway == nil {
					giveaway = &Giveaway{
						StartTime: time.Now(),
						GuildId:   guild.ID,
						GuildName: guild.Name,
					}
					err := DbMap.Insert(giveaway)
					if err != nil {
						fmt.Print(err)
					}
				}
				break
			}
		}
	}
}

func finishGiveaways() {
	giveaways := getAllUnfinishedGiveaways()
	for _, giveaway := range giveaways {
		guild, _ := session.Guild(giveaway.GuildId)
		var giveawayChannelId string
		for _, channel := range guild.Channels {
			if channel.Name == config.MainChannel {
				giveawayChannelId = channel.ID
				break
			}
		}
		var participants []Participant
		_, err := DbMap.Select(&participants, "SELECT * FROM participants WHERE giveaway_id = ?", giveaway.Id)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if participants == nil || len(participants) == 0 {
			notifyWinner(giveaway.GuildId, giveawayChannelId, nil, "")
			continue
		}
		code := getCSRVCode()
		rand.Seed(time.Now().UnixNano())
		winner := participants[rand.Int()%len(participants)]
		giveaway.InfoMessageId.String = notifyWinner(giveaway.GuildId, giveawayChannelId, &winner.UserId, code)
		giveaway.InfoMessageId.Valid = true
		giveaway.EndTime.Time = time.Now()
		giveaway.EndTime.Valid = true
		giveaway.Code.String = code
		giveaway.Code.Valid = true
		giveaway.WinnerId.String = winner.UserId
		giveaway.WinnerId.Valid = true
		giveaway.WinnerName.String = winner.UserName
		giveaway.WinnerName.Valid = true
		giveaway.update()
	}
	createMissingGiveaways()
}

func getParticipantsNames(giveawayId int) ([]string, error) {
	var participants []Participant
	_, err := DbMap.Select(&participants, "SELECT user_name FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveawayId)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(participants))
	for i := range participants {
		names[i] = participants[i].UserName
	}
	return names, nil
}

func getParticipantByMessageId(messageId string) *Participant {
	var participant Participant
	err := DbMap.SelectOne(&participant, "SELECT * FROM participants WHERE message_id = ?", messageId)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &participant
}

func getParticipantsNamesString(giveawayId int) string {
	participants, err := getParticipantsNames(giveawayId)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return strings.Join(participants, ", ")
}

func notifyWinner(guildID, channelID string, winnerID *string, code string) string {

	if winnerID == nil {
		message, err := session.ChannelMessageSend(channelID, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie pomagał ;(")
		if err != nil {
			fmt.Println(err)
		}
		return message.ID
	}
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wygrałeś kod na serwer diamond!",
			IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
		},
		Description: "Gratulacje! W loterii wygrałeś darmowy kod na serwer w CraftServe!",
	}
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "KOD:", Value: code})
	winner, err := session.GuildMember(guildID, *winnerID)
	if err != nil {
		fmt.Println(err)
	}
	dm, err := session.UserChannelCreate(*winnerID)
	if err != nil {
		fmt.Println(err)
	}
	_, err = session.ChannelMessageSendEmbed(dm.ID, &embed)
	if err != nil {
		fmt.Println(err)
	}
	embed = discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wyniki giveaway!",
			IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
		},
		Description: winner.User.Username + " wygrał kod. Moje gratulacje ;)",
	}
	message, err := session.ChannelMessageSendEmbed(channelID, &embed)
	if err != nil {
		fmt.Println(err)
	}
	return message.ID
}

func deleteFromGiveaway(userID, guildID string) {
	//TODO: PRZYTUL BAZE
	return
}

func blacklistUser(userID, guildID string) {
	//TODO: PRZYTUL BAZE
	return
}

func isBlacklisted(userID, guildID string) bool {
	//TODO: PRZYTUL BAZE
	return false
}
