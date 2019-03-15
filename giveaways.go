package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var forceStart chan string

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

func getCurrentGiveawayTime() time.Time {
	//TODO: przytul baze
	return time.Now().Add(5 * time.Minute)
}

func waitForGiveaway() {
	giveawayTime := getCurrentGiveawayTime()
	select {
	case <-forceStart:
	case <-time.After(time.Until(giveawayTime)):
	}
	finishGiveaways()
}

func finishGiveaways() {
	//TODO: przytul baze ogłoś wygranko
	// notifyWinner()
	// printGiveawayInfo()
	return
}

func getParticipants(guildID *string) (participants []string) {
	//TODO: przytul baze
	return
}

func notifyWinner(s *discordgo.Session, guildID *string, channelID *string, winnerID *string) {

	if winnerID == nil {
		_, err := s.ChannelMessageSend(*channelID, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie pomagał ;(")
		if err != nil {
			fmt.Println(err)
		}
	}
	embed := discordgo.MessageEmbed{}
	embed.Author = &discordgo.MessageEmbedAuthor{
		URL:     "https://craftserve.pl",
		Name:    "Wygrałeś kod na serwer diamond!",
		IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
	}
	embed.Description = "Gratulacje! W loterii wygrałeś darmowy kod na serwer w CraftServe!"
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "KOD:", Value: getCSRVCode()})
	winner, err := s.GuildMember(*guildID, *winnerID)
	if err != nil {
		fmt.Println(err)
	}
	dm, err := s.UserChannelCreate(*winnerID)
	if err != nil {
		fmt.Println(err)
	}
	_, err = s.ChannelMessageSendEmbed(dm.ID, &embed)
	if err != nil {
		fmt.Println(err)
	}
	embed = discordgo.MessageEmbed{}
	embed.Author = &discordgo.MessageEmbedAuthor{
		URL:     "https://craftserve.pl",
		Name:    "Wyniki giveaway!",
		IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
	}
	embed.Description = winner.User.Username + " wygrał kod. Moje gratulacje ;)"
	_, err = s.ChannelMessageSendEmbed(*channelID, &embed)
	if err != nil {
		fmt.Println(err)
	}
}

func deleteFromGiveaway(userID, guildID *string) {
	//TODO: PRZYTUL BAZE
	return
}

func blacklistUser(userID, guildID *string) {
	//TODO: PRZYTUL BAZE
	return
}
