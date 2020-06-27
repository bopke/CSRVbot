package main

import (
	"database/sql"
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

func isThxmeMessage(messageID string) bool {
	ret, err := DbMap.SelectInt("SELECT count(*) FROM ParticipantCandidates WHERE message_id = ?", messageID)
	if err != nil {
		log.Panicln("isThxmeMessage DbMap.SelectInt " + err.Error())
	}

	return ret == 1
}

func updateThxInfoMessage(messageId *string, guildId, channelId, participantId string, giveawayId int, confirmerId *string, state State) *string {
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
			"Nagrody rozdajemy o " + config.GiveawayTimeString + ", Powodzenia!",
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
	notifyThxOnThxInfoChannel(guildId, channelId, message.ID, participantId, confirmerId, state)
	return &message.ID
}

func notifyThxOnThxInfoChannel(guildId, channelId, messageId, participantId string, confirmerId *string, state State) {
	embed := discordgo.MessageEmbed{
		Timestamp: time.Now().Format(time.RFC3339),
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Nowe podziękowanie",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Color: 0x234d20,
	}
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Dla", Value: "<@" + participantId + ">", Inline: true})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Kanał", Value: "<#" + channelId + ">", Inline: true})
	if state == wait {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Oczekiwanie", Inline: true})
	} else if state == confirm {
		if confirmerId != nil {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Potwierdzono przez <@" + *confirmerId + ">", Inline: true})
		}
	} else if state == reject {
		if confirmerId != nil {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Odrzucono przez <@" + *confirmerId + ">", Inline: true})
		}
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Link", Value: "https://discordapp.com/channels/" + guildId + "/" + channelId + "/" + messageId, Inline: false})

	var serverConfig ServerConfig
	err := DbMap.SelectOne(&serverConfig, "SELECT * from ServerConfig WHERE guild_id=?", guildId)
	if err != nil {
		log.Println("notifyThxOnThxInfoChannel Unable to read from database! ", err)
		return
	}
	if serverConfig.ThxInfoChannel == "" {
		return
	}

	var thxNotification ThxNotification
	err = DbMap.SelectOne(&thxNotification, "SELECT * from ThxNotifications WHERE message_id=?", messageId)
	if err == sql.ErrNoRows {

		message, err := session.ChannelMessageSendEmbed(serverConfig.ThxInfoChannel, &embed)
		if err != nil {
			log.Println("notifyThxOnThxInfoChannel Unable to send thx info! ", err)
			return
		}
		thxNotification = ThxNotification{
			MessageId:                messageId,
			ThxNotificationMessageId: message.ID,
		}
		err = DbMap.Insert(&thxNotification)
		if err != nil {
			log.Println("notifyThxOnThxInfoChannel Unable to insert to database! ", err)
			return
		}
		return
	}
	_, err = session.ChannelMessageEditEmbed(serverConfig.ThxInfoChannel, thxNotification.ThxNotificationMessageId, &embed)
	if err != nil {
		log.Println("notifyThxOnThxInfoChannel Unable to edit embed! ", err)
	}
}
