package Commands

import (
	"csrvbot/Database"
	"csrvbot/Giveaways"
	"csrvbot/Models"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleThxmeCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) != 1 || len(m.Mentions) != 1 {
		Giveaways.PrintGiveawayInfo(session, m.ChannelID, m.GuildID)
		return
	}
	if m.Author.ID == m.Mentions[0].ID {
		_, err := session.ChannelMessageSend(m.ChannelID, "Nie można poprosic o podziękowanie samego siebie!")
		if err != nil {
			log.Println("Commands HandleThxmeCommand Unable to send channel message! ", err)
		}
		return
	}
	guild, err := session.Guild(m.GuildID)
	if err != nil {
		log.Println("Commands HandleThxmeCommand Unable to get guild! ", err)
		_, err = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
		if err != nil {
			log.Println("Commands HandleThxmeCommand Unable to send channel message! ", err)
		}
		return
	}
	log.Println(m.Author.Username + " asked for thx from " + m.Mentions[0].Username + " on " + guild.Name)
	if m.Mentions[0].Bot {
		_, err = session.ChannelMessageSend(m.ChannelID, "Nie można prosic o podziękowanie bota!")
		if err != nil {
			log.Println("Commands HandleThxmeCommand Unable to send channel message! ", err)
		}

		return
	}
	if Giveaways.IsBlacklisted(m.GuildID, m.Author.ID) {
		_, err = session.ChannelMessageSend(m.ChannelID, "Nie mozesz poprosic o thx, gdyż jestes na czarnej liscie!")
		if err != nil {
			log.Println("Commands HandleThxmeCommand Unable to send channel message! ", err)
		}
		return
	}
	candidate := &Models.ParticipantCandidate{
		CandidateName:         m.Author.Username,
		CandidateId:           m.Author.ID,
		CandidateApproverName: m.Mentions[0].Username,
		CandidateApproverId:   m.Mentions[0].ID,
		GuildName:             guild.Name,
		GuildId:               m.GuildID,
		ChannelId:             m.ChannelID,
	}
	messageId, err := session.ChannelMessageSend(m.ChannelID, m.Mentions[0].Mention()+", czy chcesz podziękować użytkownikowi "+m.Author.Mention()+"? Kliknij reakcję")
	if err != nil {
		log.Println("Commands HandleThxmeCommand Unable to send channel message! ", err)
	}
	candidate.MessageId = messageId.ID
	err = Database.DbMap.Insert(candidate)
	if err != nil {
		_, err2 := session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu kandydata do podziekowania :(")
		if err2 != nil {
			log.Println("Commands HandleThxmeCommand Unable to send channel message! ", err)
		}
		log.Panicln("Commands HandleThxmeCommand Unable to insert to database! ", err)
	}
	err = session.MessageReactionAdd(m.ChannelID, candidate.MessageId, "✅")
	if err != nil {
		log.Println("Commands HandleThxmeCommand Unable to add reaction to message! ", err)
	}
	err = session.MessageReactionAdd(m.ChannelID, candidate.MessageId, "⛔")
	if err != nil {
		log.Println("Commands HandleThxmeCommand Unable to add reaction to message! ", err)
	}
}
