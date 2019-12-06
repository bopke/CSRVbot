package CsrvbotSubcommands

import (
	"csrvbot/Giveaways"
	"csrvbot/Utils"
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleDeleteSubcommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	member, err := session.GuildMember(m.GuildID, m.Message.Author.ID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleDeleteCommand Unable to get guild member! ", err)
		return
	}
	if !Utils.HasAdminPermissions(session, member, m.GuildID) {
		_, err = session.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleDeleteCommand Unable to send channel message! ", err)
		}
		return
	}
	if len(args) == 1 {
		_, err := session.ChannelMessageSend(m.ChannelID, "Musisz podać ID użytkownika!")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleDeleteCommand Unable to send channel message! ", err)
		}
		return
	}
	guild, err := session.Guild(m.GuildID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleDeleteCommand Unable to get guild! ", err)
		return
	}
	if len(m.Mentions) < 1 {
		log.Println(m.Author.Username + " deleted ID " + args[1] + " from giveaway on " + guild.Name)
		Giveaways.DeleteParticipantFromGiveaway(session, m.GuildID, args[1])
		_, err = session.ChannelMessageSend(m.ChannelID, "Usunięto z giveawaya.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleDeleteCommand Unable to send channel message! ", err)
		}
		return
	}
	log.Println(m.Author.Username + " deleted " + m.Mentions[0].Username + " from giveaway on " + guild.Name)
	Giveaways.DeleteParticipantFromGiveaway(session, m.GuildID, m.Mentions[0].ID)
	_, err = session.ChannelMessageSend(m.ChannelID, "Usunięto z giveawaya.")
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleDeleteCommand Unable to send channel message! ", err)
	}
}
