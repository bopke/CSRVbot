package Utils

import (
	"csrvbot/Config"
	"csrvbot/Giveaways"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"time"
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

func PrintGiveawayInfo(session *discordgo.Session, channelID, guildID string) *discordgo.Message {
	splittedCronString := strings.Split(Config.GiveawayCronString, " ")
	giveawayTimeString := splittedCronString[1] + ":" + splittedCronString[2]
	info := "**Ten bot organizuje giveaway kodów na serwery Diamond.**\n" +
		"**Każdy kod przedłuża serwer o 7 dni.**\n" +
		"Aby wziąć udział pomagaj innym użytkownikom. Jeżeli komuś pomożesz, to poproś tą osobę aby napisala `!thx @TwojNick` - w ten sposób dostaniesz się do loterii. To jest nasza metoda na rozruszanie tego Discorda, tak, aby każdy mógł liczyć na pomoc. Każde podziękowanie to jeden los, więc warto pomagać!\n\n" +
		"**Sponsorem tego bota jest https://craftserve.pl/ - hosting serwerów Minecraft.**\n\n" +
		"Pomoc musi odbywać się na tym serwerze na tekstowych kanałach publicznych.\n\n" +
		"Uczestnicy: " +
		Giveaways.GetParticipantsNamesString(Giveaways.GetGiveawayForGuild(guildID).Id) +
		"\n\nNagrody rozdajemy o " + giveawayTimeString + ", Powodzenia!"
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Informacje o Giveawayu",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: info,
		Color:       0x234d20,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	m, err := session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Println("Utils PrintGiveawayInfo Unable to send channel message embed ", err)
	}
	return m
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
