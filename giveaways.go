package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var forceStart = make(chan int, 1)

func getNextGiveawayTime() time.Time {
	now := time.Now()
	if now.Hour() > config.GiveawayTimeH || (now.Hour() == config.GiveawayTimeH && now.Minute() >= config.GiveawayTimeM) {
		now = now.Add(24 * time.Hour)
	}
	return time.Date(now.Year(),
		now.Month(),
		now.Day(),
		config.GiveawayTimeH,
		config.GiveawayTimeM,
		0,
		0,
		now.Location())
}

func getCurrentGiveawayTime(giveawayId int) time.Time {
	var giveaway Giveaway
	err := DbMap.SelectOne(&giveaway, "SELECT start_time FROM giveaways WHERE GiveawayId = ?", giveawayId)
	if err != nil {
		fmt.Println(err)
		// TODO: z rozumkiem to jakoś zrobić
		return time.Now().Add(24 * time.Hour)
	}
	return giveaway.StartTime.Add(24 * time.Hour)
}

func getGiveawayForGuild(guildId *string) *Giveaway {
	var giveaway Giveaway
	err := DbMap.SelectOne(&giveaway, "SELECT * FROM Giveaways WHERE guildId = ? AND EndTime IS NULL", *guildId)
	if err != nil {
		fmt.Println(err)
		// TODO: z rozumkiem to jakoś zrobić
		return nil
	}
	return &giveaway
}

func waitForGiveaway(giveawayID int) {
	giveawayTime := getCurrentGiveawayTime(giveawayID)
thatwasntme:
	select {
	case x := <-forceStart:
		if x != giveawayID {
			goto thatwasntme
		}
	case <-time.After(time.Until(giveawayTime)):
	}
	finishGiveaway(giveawayID)
}

func getAllUnfinishedGiveaways() []Giveaway {
	var res []Giveaway
	_, err := DbMap.Select(&res, "SELECT * FROM giveaways WHERE EndTime IS NULL")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return res
}

func waitForGiveaways() {
	return
}

func finishGiveaway(giveawayId int) {
	var giveaway Giveaway
	err := DbMap.SelectOne(&giveaway, "SELECT * FROM Giveaways WHERE giveawayId = ?", giveawayId)
	if err != nil {
		fmt.Println(err)
		return
	}
	participants, err := getParticipants(giveawayId)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(participants) == 0 {
		guild, err := session.Guild(giveaway.GuildId)
		if err != nil {
			fmt.Println(err)
			return
		}
		for i := range guild.Channels {
			if guild.Channels[i].Name == config.MainChannel {
				_, err := session.ChannelMessageSend(guild.Channels[i].ID, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie pomagał ;(")
				if err != nil {

				}
			}
		}
	}
	// notifyWinner()
	// printGiveawayInfo()
	return
}

func getParticipants(giveawayId int) ([]Participant, error) {
	var res []Participant
	_, err := DbMap.Select(&res, "SELECT * FROM Participants WHERE giveawayId = ? AND is_accepted = true", giveawayId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func getParticipantsNames(giveawayId int) ([]string, error) {
	var participants []Participant
	_, err := DbMap.Select(&participants, "SELECT UserName FROM Participants WHERE giveawayId = ? AND is_accepted = true", giveawayId)
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
	err := DbMap.SelectOne(participant, "SELECT * FROM participants WHERE message_id = ?", messageId)
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

func notifyWinner(guildID, channelID string, winnerID *string) {

	if winnerID == nil {
		_, err := session.ChannelMessageSend(channelID, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie pomagał ;(")
		if err != nil {
			fmt.Println(err)
		}
		return
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
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "KOD:", Value: getCSRVCode()})
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
	_, err = session.ChannelMessageSendEmbed(channelID, &embed)
	if err != nil {
		fmt.Println(err)
	}
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
	return true
}
