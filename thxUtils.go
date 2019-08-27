package main

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

type State byte

const (
	wait State = iota
	confirm
	reject
)

func isThxMessage(messageID string) bool {
	ret, err := DbMap.SelectInt("SELECT count(*) FROM Participants WHERE message_id = ?", messageID)
	if err != nil {
		log.Panicln("isThxMessage DbMap.SelectInt " + err.Error())
	}
	if ret == 1 {
		return true
	}
	return false
}

func updateThxInfoMessage(messageId *string, channelId, participantId string, giveawayId int, confirmerId *string, state State) *string {
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Giveaway info",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: "**Ten bot organizuje giveaway kodów na serwery Diamond. Każdy kod przedłuża serwer o 7 dni.**\n" +
			"Aby wziąć udział pomagaj innym użytkownikom. " +
			"Jeżeli komuś pomożesz, to poproś tą osobę aby napisala `!thx @TwojNick` - w ten sposób dostaniesz się do loterii. " +
			"To jest nasza metoda na rozruszanie tego Discorda, tak, aby każdy mógł liczyć na pomoc. " +
			"Każde podziękowanie to jeden los, więc warto pomagać!\n\n" +
			"**Pomoc musi odbywać się na tym serwerze na tekstowych kanałach publicznych.**\n\n" +
			"W aktualnym giveawayu są: " + getParticipantsNamesString(giveawayId) + "\n\n" +
			"Nagrody rozdajemy o " + config.GiveawayTimeS + ", Powodzenia!",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	embed.Color = 0x234d20
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Dodany", Value: "<@" + participantId + ">", Inline: true})
	var status string
	if state == wait {
		status = "Oczekiwanie"
	} else if state == confirm {
		if confirmerId != nil {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Potwierdzający", Value: "<@" + *confirmerId + ">", Inline: true})
		}
		status = "Potwierdzono"
	} else if state == reject {
		if confirmerId != nil {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Odrzucający", Value: "<@" + *confirmerId + ">", Inline: true})
		}
		status = "Odrzucono"
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: status, Inline: true})
	var message *discordgo.Message
	var err error
	if messageId != nil {
		message, err = session.ChannelMessageEditEmbed(channelId, *messageId, &embed)
		if err != nil {
			log.Println("updateThxInfoMessage session.ChannelMessageEditEmbed(" + channelId + ", " + *messageId + ", embed) " + err.Error())
			return nil
		}
	} else {
		message, err = session.ChannelMessageSendEmbed(channelId, &embed)
		if err != nil {
			log.Println("updateThxInfoMessage session.ChannelMessageEditEmbed(" + channelId + ", nil, embed) " + err.Error())
			return nil
		}
	}
	return &message.ID
}
