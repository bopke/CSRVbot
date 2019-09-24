package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
)

type Config struct {
	MysqlString   string `json:"mysql_string"`
	GiveawayTimeS string `json:"giveaway_time"`
	GiveawayTimeH int    `json:"-"`
	GiveawayTimeM int    `json:"-"`
	SystemToken   string `json:"system_token"`
	CsrvSecret    string `json:"csrv_secret"`
}

var config Config

var session *discordgo.Session

func loadConfig() (c Config) {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Panic(err)
	}
	defer configFile.Close()
	err = json.NewDecoder(configFile).Decode(&c)
	if err != nil {
		log.Panic("loadConfig Decoder.Decode(&c) " + err.Error())
	}
	colon := strings.Index(c.GiveawayTimeS, ":")
	h, err := strconv.Atoi(c.GiveawayTimeS[:colon])
	if err != nil {
		log.Panic("loadConfig strconv.Atoi(" + c.GiveawayTimeS[:colon] + ") " + err.Error())
	}
	if h > 23 || h < 0 {
		panic("Hour must be greater or equal 0 and less than 24!")
	}
	m, err := strconv.Atoi(c.GiveawayTimeS[colon+1:])
	if err != nil {
		log.Panic("loadConfig strconv.Atoi(" + c.GiveawayTimeS[colon+1:] + ") " + err.Error())
	}
	if m > 59 || m < 0 {
		log.Panic("Minutes must be greater or equal 0 and less than 60!")
	}
	c.GiveawayTimeM = m
	c.GiveawayTimeH = h
	return
}

func InitLog() {
	file, err := os.OpenFile("csrvbot.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(file)
}

func main() {
	//	InitLog()
	config = loadConfig()
	InitDB()
	var err error
	session, err = discordgo.New("Bot " + config.SystemToken)
	if err != nil {
		panic(err)
	}

	session.AddHandler(OnMessageCreate)
	session.AddHandler(OnMessageReactionAdd)
	session.AddHandler(OnGuildCreate)
	session.AddHandler(OnGuildMemberUpdate)
	session.AddHandler(OnGuildMemberAdd)
	err = session.Open()
	if err != nil {
		panic(err)
	}

	c := cron.New()
	_ = c.AddFunc(fmt.Sprintf("0 %d %d * * *", config.GiveawayTimeM, config.GiveawayTimeH), finishGiveaways)
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
	info := "**Ten bot organizuje giveaway kodów na serwery Diamond.**\n" +
		"**Każdy kod przedłuża serwer o 7 dni.**\n" +
		"Aby wziąć udział pomagaj innym użytkownikom. Jeżeli komuś pomożesz, to poproś tą osobę aby napisala `!thx @TwojNick` - w ten sposób dostaniesz się do loterii. To jest nasza metoda na rozruszanie tego Discorda, tak, aby każdy mógł liczyć na pomoc. Każde podziękowanie to jeden los, więc warto pomagać!\n\n" +
		"**Sponsorem tego bota jest https://craftserve.pl/ - hosting serwerów Minecraft.**\n\n" +
		"Pomoc musi odbywać się na tym serwerze na tekstowych kanałach publicznych.\n\n" +
		"Uczestnicy: " +
		getParticipantsNamesString(getGiveawayForGuild(guildID).Id) +
		"\n\nNagrody rozdajemy o " + config.GiveawayTimeS + ", Powodzenia!"
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
		after = members[999].GuildID
	}
	return allMembers
}

func updateAllMembersInfo(guildId string) {
	guildMembers := getAllMembers(guildId)
	for _, member := range guildMembers {
		updateMemberSavedRoles(member)
	}
}
