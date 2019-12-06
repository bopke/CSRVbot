package Commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleSetwinnerCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 1 {
		_, err := s.ChannelMessageSend(m.ChannelID, "Na kogo ustawiamy?")
		if err != nil {
			log.Println("Commands HandleSetwinnerCommand Unable to send channel message! ", err)
		}
		return
	}
	_, err := s.ChannelMessageSend(m.ChannelID, ":ok_hand:")
	if err != nil {
		log.Println("Commands HandleSetwinnerCommand Unable to send channel message! ", err)
	}
}
