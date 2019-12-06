package Commands

import (
	"csrvbot/Commands/CsrvbotSubcommands"
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleCsrvbotCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) >= 2 {
		switch args[0] {
		case "info":
			CsrvbotSubcommands.HandleInfoCommand(session, m, args[1:])
			return
		case "start":
			CsrvbotSubcommands.HandleStartCommand(session, m, args[1:])
			return
		case "delete":
			CsrvbotSubcommands.HandleDeleteSubcommand(session, m, args[1:])
			return
		case "blacklist":
			CsrvbotSubcommands.HandleBlacklistCommand(session, m, args[1:])
			return
		case "unblacklist":
			CsrvbotSubcommands.HandleUnblacklistCommand(session, m, args[1:])
			return
		case "setGiveawayChannelName":
			CsrvbotSubcommands.HandleSetGiveawayChannelNameCommand(session, m, args[1:])
			return
		case "setBotAdminRoleName":
			CsrvbotSubcommands.HandleSetBotAdminRoleNameCommand(session, m, args[1:])
			return
		case "resend":
			CsrvbotSubcommands.HandleResendCommand(session, m, args[1:])
			return
		}
	}
	_, err := session.ChannelMessageSend(m.ChannelID, "!csrvbot <delete|resend|start|blacklist|unblacklist|setGiveawayChannelName|setBotAdminRoleName|info>")
	if err != nil {
		log.Println("Commands HandleCsrvbotCommand Unable to send channel message! ", err)
	}
}
