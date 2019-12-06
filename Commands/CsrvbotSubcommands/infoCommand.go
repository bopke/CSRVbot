package CsrvbotSubcommands

import (
	"csrvbot/Utils"
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleInfoCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	member, err := session.GuildMember(m.GuildID, m.Message.Author.ID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleInfoCommand Unable to get guild member! ", err)
		return
	}
	if !Utils.HasAdminPermissions(session, member, m.GuildID) {
		_, err = session.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleInfoCommand Unable to send channel message! ", err)
		}
		return
	}
	Utils.PrintServerInfo(session, m.ChannelID, m.GuildID)

}
