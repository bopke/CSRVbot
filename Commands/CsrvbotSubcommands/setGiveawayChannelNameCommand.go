package CsrvbotSubcommands

import (
	"csrvbot/Database"
	"csrvbot/ServerConfiguration"
	"csrvbot/Utils"
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleSetGiveawayChannelNameCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	member, err := session.GuildMember(m.GuildID, m.Message.Author.ID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleSetGiveawayChannelNameCommand Unable to get guild member! ", err)
		return
	}
	if !Utils.HasAdminPermissions(session, member, m.GuildID) {
		_, err = session.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleSetGiveawayChannelNameCommand Unable to send channel message! ", err)
		}
		return
	}
	if len(args) == 1 {
		_, err := session.ChannelMessageSend(m.ChannelID, "Musisz podać nazwę kanału!")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleSetGiveawayChannelNameCommand Unable to send channel message! ", err)
		}
		return
	}
	serverConfig := ServerConfiguration.GetServerConfigForGuildId(m.GuildID)
	serverConfig.MainChannel = args[2]
	_, err = Database.DbMap.Update(&serverConfig)
	if err != nil {
		log.Panicln("Commands CsrvbotSubcommands HandleSetGiveawayChannelNameCommand Unable to update in database! ", err)
	}
	_, err = session.ChannelMessageSend(m.ChannelID, "Ustawiono.")
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleSetGiveawayChannelNameCommand Unable to send channel message! ", err)
	}
}
