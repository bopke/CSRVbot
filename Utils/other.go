package Utils

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func GetAllMembers(session *discordgo.Session, guildId string) []*discordgo.Member {
	after := ""
	var allMembers []*discordgo.Member
	for {
		members, err := session.GuildMembers(guildId, after, 1000)
		if err != nil {
			log.Println("Utils GetAllMembers Unable to get members! ", err)
			return nil
		}
		allMembers = append(allMembers, members...)
		if len(members) != 1000 {
			break
		}
		after = members[999].User.ID
	}
	return allMembers
}

func PrintServerInfo(session *discordgo.Session, channelID, guildID string) *discordgo.Message {
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Informacje o serwerze",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: "ID:" + guildID,
		Color:       0x234d20,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	guild, err := session.Guild(guildID)
	if err != nil {
		log.Println("Utils PrintServerInfo Unable to get guild! ", err)
		return nil
	}
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Region", Value: guild.Region})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Kanały", Value: fmt.Sprintf("%d kanałów", len(guild.Channels))})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: fmt.Sprintf("Użytkowników [%d]", guild.MemberCount), Value: "Wielu"})
	createTime, _ := guild.JoinedAt.Parse()
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Data utworzenia", Value: createTime.Format(time.RFC1123)})
	msg, err := session.ChannelMessageSendEmbed(channelID, &embed)
	if err != nil {
		log.Println("Utils PrintServerInfo Unable to send channel message embed! ", err)
	}
	return msg
}
