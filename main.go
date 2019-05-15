package main

import (
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
	MysqlString     string `json:"mysql_string"`
	MainChannel     string `json:"main_channel"`
	BlacklistedRole string `json:"blacklisted_role"`
	AdminRole       string `json:"admin_role"`
	GiveawayTimeS   string `json:"giveaway_time"`
	GiveawayTimeH   int    `json:"-"`
	GiveawayTimeM   int    `json:"-"`
	SystemToken     string `json:"system_token"`
	CsrvSecret      string `json:"csrv_secret"`
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
	InitLog()
	config = loadConfig()
	InitDB()
	var err error
	session, err = discordgo.New("Bot " + config.SystemToken)
	if err != nil {
		panic(err)
	}

	session.AddHandler(OnMessageCreate)
	session.AddHandler(OnMessageReactionAdd)
	err = session.Open()
	if err != nil {
		panic(err)
	}
	createMissingGiveaways()

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
			IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
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
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Kanały", Value: string(len(guild.Channels)) + " kanałów"})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: fmt.Sprintf("Użytkowników [%d]", guild.MemberCount), Value: "Wielu"})
	createTime, _ := guild.JoinedAt.Parse()
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Data utworzenia", Value: createTime.Format(time.RFC3339)})
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
		"\n\nNagrody rozdajemy o 19:00, Powodzenia!"
	m, _ := session.ChannelMessageSend(channelID, info)
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
