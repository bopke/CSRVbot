package main

import (
	"encoding/json"
	"fmt"
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
	MysqlPort       int    `json:"mysql_port"`
	MysqlUser       string `json:"mysql_user"`
	MainChannel     string `json:"main_channel"`
	BlacklistedRole string `json:"blacklisted_role"`
	AdminRole       string `json:"admin_role"`
	MysqlDatabase   string `json:"mysql_database"`
	MysqlPassword   string `json:"mysql_password"`
	MysqlHost       string `json:"mysql_host"`
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
		panic(err)
	}
	defer configFile.Close()
	err = json.NewDecoder(configFile).Decode(&c)
	if err != nil {
		panic(err)
	}
	colon := strings.Index(c.GiveawayTimeS, ":")
	h, err := strconv.Atoi(c.GiveawayTimeS[:colon])
	if err != nil {
		panic(err)
	}
	if h > 23 || h < 0 {
		panic("Hour must be greater or equal 0 and less than 24!")
	}
	m, err := strconv.Atoi(c.GiveawayTimeS[colon+1:])
	if err != nil {
		panic(err)
	}
	if m > 59 || m < 0 {
		panic("Minutes must be greater or equal 0 and less than 60!")
	}
	c.GiveawayTimeM = m
	c.GiveawayTimeH = h
	return
}

func main() {
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
	for i := 0; i < len(session.State.Guilds); i++ {
		// Jak się tak dziwnie nie wyciągnie gildii to nie działa
		guild, _ := session.Guild(session.State.Guilds[i].ID)
		for _, channel := range guild.Channels {
			if channel.Name == config.MainChannel {
				giveaway := getGiveawayForGuild(guild.ID)
				if giveaway == nil {
					giveaway = &Giveaway{
						StartTime: time.Now(),
						GuildId:   guild.ID,
						GuildName: guild.Name,
					}
					err := DbMap.Insert(giveaway)
					if err != nil {
						fmt.Print(err)
					}
				}
				break
			}
		}
	}

	c := cron.New()

	fmt.Println(fmt.Sprintf("0 %d %d * * *", config.GiveawayTimeM, config.GiveawayTimeH))
	_ = c.AddFunc(fmt.Sprintf("0 %d %d * * *", config.GiveawayTimeM, config.GiveawayTimeH), finishGiveaways)
	c.Start()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	err = session.Close()
	if err != nil {
		panic(err)
	}
}

func printServerInfo(channelID *string, guildID *string) *discordgo.Message {
	embed := discordgo.MessageEmbed{}
	embed.Author = &discordgo.MessageEmbedAuthor{
		URL:     "https://craftserve.pl",
		Name:    "Informacje o serwerze",
		IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
	}
	guild, err := session.Guild(*guildID)
	if err != nil {
		fmt.Println(err)
	}
	embed.Description = "ID:" + *guildID
	//TODO: kolor embedu jakis sensowny
	//	embed.Color
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Region", Value: guild.Region})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Kanały", Value: string(len(guild.Channels)) + " kanałów"})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Użytkowników [" + string(guild.MemberCount) + "]"})
	//TODO: czas!
	//	embed.Fields = append(embed.Fields,&discordgo.MessageEmbedField{Name: "Data utworzenia",Value:time.Unix(guild.JoinedAt)})
	msg, err := session.ChannelMessageSendEmbed(*channelID, &embed)
	if err != nil {
		fmt.Println(err)
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
	m, err := session.ChannelMessageSend(channelID, info)
	if err != nil {
		fmt.Println(err)
	}
	return m
}

func getCSRVCode() string {
	return "TEST"
}
