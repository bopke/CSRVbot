package CsrvbotSubcommands

import (
	"csrvbot/Database"
	"csrvbot/ServerConfiguration"
	"csrvbot/Utils"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleSetBotAdminRoleNameCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	member, err := session.GuildMember(m.GuildID, m.Message.Author.ID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleSetBotAdminRoleNameCommand Unable to get guild member! ", err)
		return
	}
	if !Utils.HasAdminPermissions(session, member, m.GuildID) {
		_, err = session.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleSetBotAdminRoleNameCommand Unable to send channel message! ", err)
		}
		return
	}
	if len(args) == 1 {
		_, err := session.ChannelMessageSend(m.ChannelID, "Musisz podać nazwę roli!")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleSetBotAdminRoleNameCommand Unable to send channel message! ", err)
		}
		return
	}
	serverConfig := ServerConfiguration.GetServerConfigForGuildId(m.GuildID)
	serverConfig.AdminRole = args[1]
	_, err = Database.DbMap.Update(&serverConfig)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleSetBotAdminRoleNameCommand Unable to update in database! ", err)
	}
	_, err = session.ChannelMessageSend(m.ChannelID, "Ustawiono.")
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleSetBotAdminRoleNameCommand Unable to send channel message! ", err)
	}
}
