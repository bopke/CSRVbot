package Commands

import (
	"csrvbot/Database"
	"csrvbot/Giveaways"
	"csrvbot/Utils"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func HandleThxCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) != 1 || len(m.Mentions) != 1 {
		Utils.PrintGiveawayInfo(session, m.ChannelID, m.GuildID)
		return
	}
	if m.Author.ID == m.Mentions[0].ID {
		_, err := session.ChannelMessageSend(m.ChannelID, "Nie można dziękować sobie!")
		if err != nil {
			log.Println("Commands HandleThxCommand Unable to send channel message! ", err)
		}
		return
	}
	guild, err := session.Guild(m.GuildID)
	if err != nil {
		log.Println("Command HandleThxCommand Unable to get guild! ", err)
		_, err = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
		if err != nil {
			log.Println("Commands HandleThxCommand Unable to send channel message! ", err)
		}
		return
	}
	log.Println(m.Author.Username + " thxed " + m.Mentions[0].Username + " on " + guild.Name)
	if m.Mentions[0].Bot {
		_, err = session.ChannelMessageSend(m.ChannelID, "Nie można dziękować botom!")
		if err != nil {
			log.Println("Commands HandleThxCommand Unable to send channel message! ", err)
		}
		return
	}
	if Giveaways.IsBlacklisted(m.GuildID, m.Mentions[0].ID) {
		_, err = session.ChannelMessageSend(m.ChannelID, "Ten użytkownik jest na czarnej liście i nie może brać udziału :(")
		if err != nil {
			log.Println("Commands HandleThxCommand Unable to send channel message! ", err)
		}
		return
	}
	participant := &Giveaways.Participant{
		UserId:     m.Mentions[0].ID,
		GiveawayId: Giveaways.GetGiveawayForGuild(m.GuildID).Id,
		CreateTime: time.Now(),
		GuildId:    m.GuildID,
		ChannelId:  m.ChannelID,
	}
	participant.GuildName = guild.Name
	participant.UserName = m.Mentions[0].Username
	participant.MessageId = *Utils.UpdateThxInfoMessage(session, nil, m.ChannelID, m.Mentions[0].ID, participant.GiveawayId, nil, Utils.Wait)
	err = Database.DbMap.Insert(participant)
	if err != nil {
		log.Println("Commands HandleThxCommand Unable to insert to database! ", err)
		_, err = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
		if err != nil {
			log.Println("Commands HandleThxCommand Unable to send channel message! ", err)
		}
	}
	err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "✅")
	if err != nil {
		log.Println("Commands HandleThxCommand Unable to add reaction to message! ", err)
	}
	err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "⛔")
	if err != nil {
		log.Println("Commands HandleThxCommand Unable to add reaction to message! ", err)
	}
}
