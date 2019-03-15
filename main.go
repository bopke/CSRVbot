package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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
var waiter chan bool

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

func getNextGiveawayTime() time.Time {
	now := time.Now()
	if now.Hour() > config.GiveawayTimeH || (now.Hour() == config.GiveawayTimeH && now.Minute() >= config.GiveawayTimeM) {
		now = now.Add(24 * time.Hour)
	}
	return time.Date(now.Year(),
		now.Month(),
		now.Day(),
		config.GiveawayTimeH,
		config.GiveawayTimeM,
		0,
		0,
		now.Location())
}

func confirmParticipant(messageId, adminId *string) {
	//TODO: PRZYTUL BAZE
	return
}

func refuseParticipant(messageId, adminId *string) {
	//TODO: PRZYTUL BAZE
	return
}

func isThxMessage(messageID *string) bool {
	//TODO: PRZYTUL BAZE
	return true
}

func deleteFromGiveaway(userID, guildID *string) {
	//TODO: PRZYTUL BAZE
	return
}

func blacklistUser(userID, guildID *string) {
	//TODO: PRZYTUL BAZE
	return
}

func getRoleID(s *discordgo.Session, guildID *string, roleName string) (string, error) {
	guild, err := s.Guild(*guildID)
	if err != nil {
		fmt.Println(err)
		return "", errors.New("unable to retrieve guild")
	}
	roles := guild.Roles
	for _, role := range roles {
		if role.Name == roleName {
			return role.ID, nil
		}
	}
	return "", errors.New("no " + roleName + " role available")
}

func hasRole(s *discordgo.Session, member *discordgo.Member, roleName string) bool {
	adminRole, err := getRoleID(s, &member.GuildID, roleName)
	if err != nil {
		fmt.Println(err)
		return false
	}
	for _, role := range member.Roles {
		if role == adminRole {
			return true
		}
	}
	return false
}

func getCurrentGiveawayTime() time.Time {
	//TODO: przytul baze
	return time.Now().Add(5 * time.Minute)
}

func getCSRVCode() string {
	return "TEST"
}

func waitForGiveaway() {
	giveawayTime := getCurrentGiveawayTime()
	select {
	case <-waiter:
	case <-time.After(time.Until(giveawayTime)):
	}
	finishGiveaways()
}

func finishGiveaways() {
	//TODO: przytul baze ogłoś wygranko
	// notifyWinner()
	// printGiveawayInfo()
	return
}

func getParticipants(guildID *string) (participants []string) {
	//TODO: przytul baze
	return
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

func notifyWinner(s *discordgo.Session, guildID *string, channelID *string, winnerID *string) {

	if winnerID == nil {
		_, err := s.ChannelMessageSend(*channelID, "Dzisiaj nikt nie wygrywa, ponieważ nikt nie pomagał ;(")
		if err != nil {
			fmt.Println(err)
		}
	}
	embed := discordgo.MessageEmbed{}
	embed.Author = &discordgo.MessageEmbedAuthor{
		URL:     "https://craftserve.pl",
		Name:    "Wygrałeś kod na serwer diamond!",
		IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
	}
	embed.Description = "Gratulacje! W loterii wygrałeś darmowy kod na serwer w CraftServe!"
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "KOD:", Value: getCSRVCode()})
	winner, err := s.GuildMember(*guildID, *winnerID)
	if err != nil {
		fmt.Println(err)
	}
	dm, err := s.UserChannelCreate(*winnerID)
	if err != nil {
		fmt.Println(err)
	}
	_, err = s.ChannelMessageSendEmbed(dm.ID, &embed)
	if err != nil {
		fmt.Println(err)
	}
	embed = discordgo.MessageEmbed{}
	embed.Author = &discordgo.MessageEmbedAuthor{
		URL:     "https://craftserve.pl",
		Name:    "Wyniki giveaway!",
		IconURL: "https://images-ext-1.discordapp.net/external/OmO5hbzkaQiEXaEF7S9z1AXSop-hks2K7QgmOtTsQO0/https/akimg0.ask.fm/assets2/067/455/391/744/normal/10378269_696841953685468_93044818520950595_n.png",
	}
	embed.Description = winner.User.Username + " wygrał kod. Moje gratulacje ;)"
	_, err = s.ChannelMessageSendEmbed(*channelID, &embed)
	if err != nil {
		fmt.Println(err)
	}
}

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

func main() {
	config = loadConfig()
	waiter = make(chan bool, 1)
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
func OnMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if !isThxMessage(&r.MessageID) {
		return
	}
	if r.UserID == s.State.User.ID {
		return
	}
	member, _ := s.GuildMember(r.GuildID, r.UserID)
	if hasRole(s, member, config.AdminRole) {
		//TODO: TAK NIE
		if r.Emoji.Name == "tak" {
			confirmParticipant(&r.MessageID, &r.UserID)
		} else if r.Emoji.Name == "nie" {
			refuseParticipant(&r.MessageID, &r.UserID)
		}
	}
}

func OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// ignore own messages
	if m.Author.ID == s.State.User.ID {
		return
	}
	// ignore other bots messages
	if m.Author.Bot {
		return
	}

	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	// remove prefix
	m.Content = m.Content[1:]
	cmds := strings.Fields(m.Content)
	if cmds[0] == "giveaway" {
		printGiveawayInfo(s, &m.ChannelID, &m.GuildID)
		return
	}
	if cmds[0] == "csrvbot" {
		_, err := s.ChannelMessageSend(m.ChannelID, "!csrvbot <delete|resend|start|blacklist|info>")
		if err != nil {
			fmt.Println(err)
		}
		if cmds[1] == "info" {
			return
		}
		if cmds[1] == "start" {
			waiter <- true
			return
		}
		if cmds[1] == "delete" {
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				fmt.Println(err)
				return
			}
			if !hasRole(s, member, config.AdminRole) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			if len(cmds) == 2 {
				_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać ID użytkownika!")
				if err != nil {
					fmt.Println(err)
				}
				return
			}
			deleteFromGiveaway(&cmds[2], &m.GuildID)
		}
		if cmds[1] == "blacklist" {
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				fmt.Println(err)
				return
			}
			if !hasRole(s, member, config.AdminRole) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
			}
			if len(cmds) == 2 {
				_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać ID użytkownika!")
				if err != nil {
					fmt.Println(err)
				}
				return
			}
			blacklistUser(&cmds[2], &m.GuildID)
		}
	}
}
