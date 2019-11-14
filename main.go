package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"csrvbot/config"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
)

var session *discordgo.Session

func InitLog() {
	file, err := os.OpenFile("csrvbot.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(file)
}

func main() {
	//	InitLog()
	err := config.Load()
	if err != nil {
		panic(err)
	}
	InitDB()
	session, err = discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		panic(err)
	}

	session.AddHandler(OnMessageCreate)
	session.AddHandler(HandleGiveawayReactions)
	session.AddHandler(HandleThxmeReactions)
	session.AddHandler(OnGuildCreate)
	session.AddHandler(OnGuildMemberUpdate)
	session.AddHandler(OnGuildMemberAdd)
	err = session.Open()
	if err != nil {
		panic(err)
	}

	c := cron.New()
	_ = c.AddFunc(config.GiveawayCronString, finishGiveaways)
	c.Start()

	log.Println("Wystartowałem")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Println("Przyjąłem polecenie wyłączenia")
	err = session.Close()
	if err != nil {
		panic(err)
	}
}

func printServerInfo(channelID, guildID string) *discordgo.Message {
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
		log.Println("printServerInfo session.Guild(" + guildID + ") " + err.Error())
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
		log.Println("printServerInfo session.ChannelMessageSendEmbed(" + channelID + ", embed) " + err.Error())
	}
	return msg
}

func printGiveawayInfo(channelID, guildID string) *discordgo.Message {
	splittedCronString := strings.Split(config.GiveawayCronString, " ")
	giveawayTimeString := splittedCronString[1] + ":" + splittedCronString[2]
	info := "**Ten bot organizuje giveaway kodów na serwery Diamond.**\n" +
		"**Każdy kod przedłuża serwer o 7 dni.**\n" +
		"Aby wziąć udział pomagaj innym użytkownikom. Jeżeli komuś pomożesz, to poproś tą osobę aby napisala `!thx @TwojNick` - w ten sposób dostaniesz się do loterii. To jest nasza metoda na rozruszanie tego Discorda, tak, aby każdy mógł liczyć na pomoc. Każde podziękowanie to jeden los, więc warto pomagać!\n\n" +
		"**Sponsorem tego bota jest https://craftserve.pl/ - hosting serwerów Minecraft.**\n\n" +
		"Pomoc musi odbywać się na tym serwerze na tekstowych kanałach publicznych.\n\n" +
		"Uczestnicy: " +
		getParticipantsNamesString(getGiveawayForGuild(guildID).Id) +
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
	m, _ := session.ChannelMessageSendEmbed(channelID, embed)
	return m
}

func getCSRVCode() (string, error) {
	req, err := http.NewRequest("POST", "https://craftserve.pl/api/generate_voucher", nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("csrvbot", config.CsrvSecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("getCSRVCode http.DefaultClient.Do(req) " + err.Error())
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Code string `json:"code"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}
	return data.Code, nil
}

func generateResendEmbed(userId string) (embed *discordgo.MessageEmbed, err error) {
	var giveaways []Giveaway
	_, err = DbMap.Select(&giveaways, "SELECT code FROM Giveaways WHERE winner_id=? ORDER BY id DESC LIMIT 10", userId)
	if err != nil {
		return
	}
	embed = &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Twoje ostatnie wygrane kody",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
	}
	embed.Description = ""
	for _, giveaway := range giveaways {
		embed.Description += giveaway.Code.String + "\n"
	}
	return
}

func createConfigurationIfNotExists(guildID string) {
	var serverConfig ServerConfig
	err := DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig")
	if err == sql.ErrNoRows {
		serverConfig.GuildId = guildID
		serverConfig.MainChannel = "giveaway"
		serverConfig.AdminRole = "CraftserveBotAdmin"
		err = DbMap.Insert(&serverConfig)
	}
	if err != nil {
		log.Panicln("createConfigurationIfNotExists DbMap.SelectOne " + err.Error())
	}
}

func getServerConfigForGuildId(guildID string) (serverConfig ServerConfig) {
	err := DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig")
	if err != nil {
		log.Panicln("createConfigurationIfNotExists DbMap.SelectOne " + err.Error())
	}
	return
}

func getAllMembers(guildId string) []*discordgo.Member {
	after := ""
	var allMembers []*discordgo.Member
	for {
		members, err := session.GuildMembers(guildId, after, 1000)
		if err != nil {
			log.Println("getAllMembers Error getting nicknames " + err.Error())
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
