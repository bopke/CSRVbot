package CsrvbotSubcommands

import (
	"csrvbot/Database"
	"csrvbot/Models"
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleResendCommand(session *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	embed := generateResendEmbed(m.Message.Author.ID)
	dm, err := session.UserChannelCreate(m.Message.Author.ID)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands generateResendEmbed Unable to create DM channel! ", err)
		return
	}
	_, err = session.ChannelMessageSendEmbed(dm.ID, embed)
	if err != nil {
		log.Println("Commands CsrvbotSubcommands generateResendEmbed Unable to send channel embed! ", err)
		return
	}
	_, err = session.ChannelMessageSend(m.ChannelID, "Kody zostaly ponownie wyslane :innocent:")
	if err != nil {
		log.Println("Commands CsrvbotSubcommands generateResendEmbed Unable to send channel message! ", err)
		return
	}
	log.Println("Sent codes to " + m.Author.Username + "#" + m.Author.Discriminator)
}

func generateResendEmbed(userId string) (embed *discordgo.MessageEmbed) {
	var givs []Models.Giveaway
	_, err := Database.DbMap.Select(&givs, "SELECT code FROM Giveaways WHERE winner_id=? ORDER BY id DESC LIMIT 10", userId)
	if err != nil {
		log.Panicln("Commands CsrvbotSubcommands generateResendEmbed Unable to select from database! ", err)
	}
	embed = &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Twoje ostatnie wygrane kody",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
	}
	embed.Description = ""
	for _, giveaway := range givs {
		embed.Description += giveaway.Code.String + "\n"
	}
	return
}
