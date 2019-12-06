package Giveaways

import (
	"csrvbot/Database"
	"csrvbot/Utils"
	"database/sql"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func GetParticipantsNames(giveawayId int) []string {
	var participants []Participant
	_, err := Database.DbMap.Select(&participants, "SELECT user_name FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveawayId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Panicln("Giveaways GetParticipantsNames Unable to select from database! ", err)
	}
	names := make([]string, len(participants))
	for i := range participants {
		names[i] = participants[i].UserName
	}
	return names
}

func GetParticipantByMessageId(messageId string) *Participant {
	var participant Participant
	err := Database.DbMap.SelectOne(&participant, "SELECT * FROM Participants WHERE message_id = ?", messageId)
	if err != nil && err != sql.ErrNoRows {
		log.Panicln("Giveaways GetParticipantByMessageId Unable to select from database ", err)
	}
	if err == sql.ErrNoRows {
		return nil
	}
	return &participant
}

func GetParticipantsByGiveawayId(giveawayId int) []Participant {
	var participants []Participant
	_, err := Database.DbMap.Select(&participants, "SELECT * FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveawayId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Panicln("Giveaways GetParticipantsByGiveawayId Unable to select from database! ", err)
	}
	return participants
}

func GetParticipantsNamesString(giveawayId int) string {
	participants := GetParticipantsNames(giveawayId)
	if participants == nil {
		return ""
	}
	return strings.Join(participants, ", ")
}

func notifyWinner(session *discordgo.Session, guildID, channelID string, winnerID *string, code string) string {
	guild, err := session.Guild(guildID)
	var guildName string
	if err != nil {
		log.Println("Giveaways notifyWinner Unable to get guild! ", err)
		guildName = guildID
	} else {
		guildName = guild.Name
	}
	if winnerID == nil {
		log.Println("Giveaway on " + guildName + " finished without participants.")
		message, err := session.ChannelMessageSend(channelID, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie pomagał ;(")
		if err != nil {
			log.Println("Giveaways notifyWinner Unable to send channel message! ", err)
		}
		return message.ID
	}
	winner, err := session.GuildMember(guildID, *winnerID)
	if err != nil {
		log.Println("Giveaways notifyWinner Unable to get guild member! ", err)
		return ""
	}
	log.Println(winner.User.Username + " have won giveaway on " + guild.Name + ". Code: " + code)
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wygrałeś kod na serwer diamond!",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: "Gratulacje! W loterii wygrałeś darmowy kod na serwer w CraftServe!",
	}
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "KOD:", Value: code})
	dm, err := session.UserChannelCreate(*winnerID)
	if err != nil {
		log.Println("Giveaways notifyWinner Unable to create user channel! ", err)
		return ""
	}
	_, err = session.ChannelMessageSendEmbed(dm.ID, &embed)
	if err != nil {
		log.Println("Giveaways notifyWinner Unable to send embed channel message! ", err)
	}
	embed = discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wyniki giveaway!",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: winner.User.Username + " wygrał kod. Moje gratulacje ;)",
	}
	message, err := session.ChannelMessageSendEmbed(channelID, &embed)
	if err != nil {
		log.Println("Giveaways notifyWinner Unable to send embed channel message! ", err)
		return ""
	}
	return message.ID
}

func DeleteParticipantFromGiveaway(session *discordgo.Session, guildID, userID string) {
	giveawayId := GetGiveawayForGuild(guildID).Id
	participants := GetParticipantsByGiveawayId(giveawayId)
	for _, participant := range participants {
		if participant.UserId == userID {
			participant.IsAccepted.Valid = true
			participant.IsAccepted.Bool = false
			_, err := Database.DbMap.Update(&participant)
			if err != nil {
				log.Panicln(err)
			}
		}
	}
	for _, participant := range participants {
		Utils.UpdateThxInfoMessage(session, &participant.MessageId, participant.ChannelId, participant.UserId, participant.GiveawayId, nil, Utils.Reject)
	}
	return
}
