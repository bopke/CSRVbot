package CsrvbotSubcommands

import (
	"csrvbot/Giveaways"
	"csrvbot/Utils"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleStartCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	member, err := session.GuildMember(m.GuildID, m.Message.Author.ID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands HandleStartCommand Unable to get guild member! ", err)
		return
	}
	if !Utils.HasAdminPermissions(session, member, m.GuildID) {
		_, err = session.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
		if err != nil {
			log.Println("Commands CsrvbotSubcommands HandleStartCommand Unable to send channel message! ", err)
		}
		return
	}
	Giveaways.FinishGiveaway(session, m.GuildID)
	Giveaways.CreateMissingGiveaways(session)
}
