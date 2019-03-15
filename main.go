package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-sql-driver/mysql"
)

type Giveaway struct {
	Id            int            `db:"id, primarykey, autoincrement"`
	StartTime     time.Time      `db:"start_time"`
	EndTime       mysql.NullTime `db:"end_time"`
	GuildId       string         `db:"guild_id,size:255"`
	GuildName     string         `db:"guild_name,size:255"`
	WinnerId      sql.NullString `db:"winner_id,size:255"`
	WinnerName    sql.NullString `db:"winner_name,size:255"`
	InfoMessageId sql.NullString `db:"info_message_id,size:255"`
	Code          sql.NullString `db:"code,size:255"`
}

type Participant struct {
	Id           int            `db:"id, primarykey, autoincrement"`
	UserName     string         `db:"user_name,size:255"`
	UserId       string         `db:"user_id,size:255"`
	GiveawayId   int            `db:"giveaway_id"`
	CreateTime   time.Time      `db:"create_time"`
	GuildName    string         `db:"guild_name,size:255"`
	GuildId      string         `db:"guild_id,size:255"`
	MessageId    string         `db:"message_id,size:255"`
	ChannelId    string         `db:"channel_id,size:255"`
	AcceptTime   mysql.NullTime `db:"accept_time"`
	AcceptUser   sql.NullString `db:"accept_user,size:255"`
	AcceptUserId sql.NullString `db:"accept_user_id,size:255"`
}

type Blacklist struct {
	Id            int    `db:"id,primarykey,autoincrement"`
	GuildId       string `db:"guild_id,size:255"`
	UserId        string `db:"user_id,size:255"`
	BlacklisterId string `db:"blacklister_id,size:255"`
}

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

//
// config stuff
//

func loadConfig() (c Config) {
	configFile, err := ioutil.ReadFile("config.json")
	print(string(configFile))
	if err != nil {
		panic(err)
	}
	if len(configFile) < 1 {
		panic("Empty config file")
	}
	err = json.Unmarshal(configFile, &c)
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

	if err != nil {
		panic(err)
	}
	return
}

//
// giveaway utils
//

func main() {
	config = loadConfig()
	forceStart = make(chan string, 1)
	discord, err := discordgo.New("Bot " + config.SystemToken)
	if err != nil {
		panic(err)
	}

	discord.AddHandler(OnMessageCreate)
	discord.AddHandler(OnMessageReactionAdd)
	err = discord.Open()
	if err != nil {
		panic(err)
	}

	go waitForGiveaway()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	err = discord.Close()
	if err != nil {
		panic(err)
	}
}

//
// other
//

func printServerInfo(s *discordgo.Session, channelID *string, guildID *string) *discordgo.Message {
	embed := discordgo.MessageEmbed{}
	embed.Author = &discordgo.MessageEmbedAuthor{
		URL:     "https://craftserve.pl",
		Name:    "Informacje o serwerze",
		IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
	}
	guild, err := s.Guild(*guildID)
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
	msg, err := s.ChannelMessageSendEmbed(*channelID, &embed)
	if err != nil {
		fmt.Println(err)
	}
	return msg
}

func printGiveawayInfo(s *discordgo.Session, channelID *string, guildID *string) *discordgo.Message {
	info := "**Ten bot organizuje giveaway kodów na serwery Diamond.**\n" +
		"**Każdy kod przedłuża serwer o 7 dni.**\n" +
		"Aby wziąć udział pomagaj innym użytkownikom. Jeżeli komuś pomożesz, to poproś tą osobę aby napisala `!thx @TwojNick` - w ten sposób dostaniesz się do loterii. To jest nasza metoda na rozruszanie tego Discorda, tak, aby każdy mógł liczyć na pomoc. Każde podziękowanie to jeden los, więc warto pomagać!\n\n" +
		"**Sponsorem tego bota jest https://craftserve.pl/ - hosting serwerów Minecraft.**\n\n" +
		"Pomoc musi odbywać się na tym serwerze na tekstowych kanałach publicznych.\n\n" +
		"Uczestnicy: "
	info += strings.Join(getParticipants(guildID), ", ")
	info += "\n\nNagrody rozdajemy o 19:00, Powodzenia!"
	m, err := s.ChannelMessageSend(*channelID, info)
	if err != nil {
		fmt.Println(err)
	}
	return m
}

func getCSRVCode() string {
	return "TEST"
}
