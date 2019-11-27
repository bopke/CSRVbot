package giveaways

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

func GetParticipantsNames(giveawayId int) []string {
	var participants []Participant
	_, err := database.DbMap.Select(&participants, "SELECT user_name FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveawayId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Panicln("GetParticipantsNames DbMap.Select ", err)
	}
	names := make([]string, len(participants))
	for i := range participants {
		names[i] = participants[i].UserName
	}
	return names
}

func GetParticipantByMessageId(messageId string) *Participant {
	var participant Participant
	err := database.DbMap.SelectOne(&participant, "SELECT * FROM Participants WHERE message_id = ?", messageId)
	if err != nil && err != sql.ErrNoRows {
		log.Panicln("GetParticipantByMessageId DbMap.SelectOne ", err)
	}
	if err == sql.ErrNoRows {
		return nil
	}
	return &participant
}

func GetParticipantsByGiveawayId(giveawayId int) []Participant {
	var participants []Participant
	_, err := database.DbMap.Select(&participants, "SELECT * FROM Participants WHERE giveaway_id = ? AND is_accepted = true", giveawayId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Panicln("GetParticipantsByGiveawayId DbMap.Select ", err)
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

func notifyWinner(session *discordgo.Session, guildID, channelID, winnerID, code string) string {
	guild, err := session.Guild(guildID)
	var guildName string
	if err != nil {
		log.Println("notifyWinner session.Guild(" + guildID + ") " + err.Error())
		guildName = guildID
	} else {
		guildName = guild.Name
	}
	if winnerID == nil {
		log.Println("Giveaway na " + guildName + " zakończył się bez uczestników.")
		message, _ := session.ChannelMessageSend(channelID, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie pomagał ;(")
		return message.ID
	}
	winner, err := session.GuildMember(guildID, *winnerID)
	if err != nil {
		log.Println("notifyWinner session.GuildMember(" + guildID + ", " + *winnerID + ") " + err.Error())
		return ""
	}
	log.Println(winner.User.Username + " wygrał giveaway na " + guild.Name + ". Kod: " + code)
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
	dm, _ := session.UserChannelCreate(*winnerID)
	_, _ = session.ChannelMessageSendEmbed(dm.ID, &embed)
	embed = discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wyniki giveaway!",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: winner.User.Username + " wygrał kod. Moje gratulacje ;)",
	}
	message, _ := session.ChannelMessageSendEmbed(channelID, &embed)
	return message.ID
}

func DeleteParticipantFromGiveaway(session *discordgo.Session, guildID, userID string) {
	giveawayId := GetGiveawayForGuild(guildID).Id
	participants := GetParticipantsByGiveawayId(giveawayId)
	for _, participant := range participants {
		if participant.UserId == userID {
			participant.IsAccepted.Valid = true
			participant.IsAccepted.Bool = false
			_, err := database.DbMap.Update(&participant)
			if err != nil {
				log.Panicln(err)
			}
		}
	}
	for _, participant := range participants {
		updateThxInfoMessage(session, &participant.MessageId, participant.ChannelId, participant.UserId, participant.GiveawayId, nil, reject)
	}
	return
}
