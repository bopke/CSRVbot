package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type State byte

const (
	wait State = iota
	confirm
	reject
)

func isThxMessage(messageID string) bool {
	ret, err := DbMap.SelectInt("SELECT count(*) FROM participants WHERE message_id = ?", messageID)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if ret == 1 {
		return true
	}
	if ret > 1 {
		fmt.Println("Co tu sie?")
	}
	return false
}

func printThxInfoMessage(channelId, participantId string, giveawayId int, state State) *string {
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Giveaway info",
			IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
		},
		Description: "**Ten bot organizuje giveaway kodów na serwery Diamond. Każdy kod przedłuża serwer o 7 dni.**\n" +
			"Aby wziąć udział pomagaj innym użytkownikom. " +
			"Jeżeli komuś pomożesz, to poproś tą osobę aby napisala `!thx @TwojNick` - w ten sposób dostaniesz się do loterii. " +
			"To jest nasza metoda na rozruszanie tego Discorda, tak, aby każdy mógł liczyć na pomoc. " +
			"Każde podziękowanie to jeden los, więc warto pomagać!\n\n" +
			"**Pomoc musi odbywać się na tym serwerze na tekstowych kanałach publicznych.**\n\n" +
			"W aktualnym giveawayu są: " + getParticipantsNamesString(giveawayId) + "\n\n" +
			"Nagrody rozdajemy o 20:00, Powodzenia!",
	}
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Added", Value: "<@" + participantId + ">", Inline: true})
	var status string
	if state == wait {
		status = "Oczekiwanie"
	} else if state == confirm {
		status = "Potwierdzono"
	} else if state == reject {
		status = "Odrzucono"
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: status, Inline: true})
	message, err := session.ChannelMessageSendEmbed(channelId, &embed)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &message.ID
}
