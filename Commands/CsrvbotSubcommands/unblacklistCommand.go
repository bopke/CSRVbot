package CsrvbotSubcommands

import (
	"csrvbot/Giveaways"
	"csrvbot/Utils"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleUnblacklistCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	member, err := session.GuildMember(m.GuildID, m.Message.Author.ID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleUnblacklistCommand Unable to get guild member! ", err)
		return
	}
	if !Utils.HasAdminPermissions(session, member, m.GuildID) {
		_, err = session.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleUnblacklistCommand Unable to send channel message! ", err)
		}
		return
	}
	if len(args) == 1 {
		_, err = session.ChannelMessageSend(m.ChannelID, "Musisz podać użytkownika!")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleUnblacklistCommand Unable to send channel message! ", err)
		}
		return
	}
	guild, err := session.Guild(m.GuildID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleUnblacklistCommand Unable to get guild! ", err)
		return
	}
	if len(m.Mentions) < 1 {
		log.Println(m.Author.Username + " unblacklisted ID " + args[1] + " on " + guild.Name)
		Giveaways.UnblacklistUser(m.GuildID, args[1])
		_, err = session.ChannelMessageSend(m.ChannelID, "Użytkownik ponownie może brać udział w giveawayach.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleUnblacklistCommand Unable to send channel message! ", err)
		}
		return
	}
	log.Println(m.Author.Username + " unblacklisted " + m.Mentions[0].Username + " on " + guild.Name)
	Giveaways.UnblacklistUser(m.GuildID, m.Mentions[0].ID)
	_, err = session.ChannelMessageSend(m.ChannelID, "Użytkownik ponownie może brać udział w giveawayach.")
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleUnblacklistCommand Unable to send channel message! ", err)
	}
}
