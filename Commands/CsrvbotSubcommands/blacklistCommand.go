package CsrvbotSubcommands

import (
	"csrvbot/Giveaways"
	"csrvbot/Utils"
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleBlacklistCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	member, err := session.GuildMember(m.GuildID, m.Message.Author.ID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleBlacklistCommand Unable to get guild member! ", err)
		return
	}
	if !Utils.HasAdminPermissions(session, member, m.GuildID) {
		_, err = session.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleBlacklistCommand Unable to send channel message! ", err)
		}
		return
	}
	if len(args) == 1 {
		_, err = session.ChannelMessageSend(m.ChannelID, "Musisz podać użytkownika!")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleBlacklistCommand Unable to send channel message! ", err)
		}
		return
	}
	guild, err := session.Guild(m.GuildID)
	if err != nil {
		_, err = session.ChannelMessageSend(m.ChannelID, "Błąd pobierania informacji o gildii.")
		log.Println("Commands CsrvbotSubcommands HandleBlacklistCommand Unable to get guild! ", err)
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleBlacklistCommand Unable to send channel message! ", err)
		}
		return
	}
	if len(m.Mentions) < 1 {
		log.Println(m.Author.Username + " blackilisted ID " + args[1] + " on " + guild.Name)
		Giveaways.BlacklistUser(m.GuildID, args[2], m.Author.ID)
		_, err = session.ChannelMessageSend(m.ChannelID, "Użytkownik został zablokowany z możliwości udziału w giveawayu.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleBlacklistCommand Unable to send channel message! ", err)
		}
		return
	}
	log.Println(m.Author.Username + " blacklisted " + m.Mentions[0].Username + " on " + guild.Name)
	Giveaways.BlacklistUser(m.GuildID, m.Mentions[0].ID, m.Author.ID)
	_, err = session.ChannelMessageSend(m.ChannelID, "Użytkownik został zablokowany z możliwości udziału w giveawayu.")
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleBlacklistCommand Unable to send channel message! ", err)

	}

}
