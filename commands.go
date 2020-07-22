package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func handleThxCommand(m *discordgo.MessageCreate, args []string) {
	if len(args) != 2 {
		printGiveawayInfo(m.ChannelID, m.GuildID)
		return
	}
	match, _ := regexp.Match("<@[!]?[0-9]*>", []byte(args[1]))
	if !match {
		printGiveawayInfo(m.ChannelID, m.GuildID)
		return
	}
	args[1] = args[1][2 : len(args[1])-1]
	if strings.HasPrefix(args[1], "!") {
		args[1] = args[1][1:]
	}
	if m.Author.ID == args[1] {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Nie można dziękować sobie!")
		return
	}
	user, err := session.User(args[1])
	if err != nil {
		return
	}

	guild, err := session.Guild(m.GuildID)
	if err != nil {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
		log.Println("OnMessageCreate session.Guild(" + m.GuildID + ") " + err.Error())
		return
	}
	log.Println(m.Author.Username + " podziękował " + user.Username + " na " + guild.Name)
	if user.Bot {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Nie można dziękować botom!")
		return
	}
	if isBlacklisted(m.GuildID, m.Mentions[0].ID) {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Ten użytkownik jest na czarnej liście i nie może brać udziału :(")
		return
	}
	participant := &Participant{
		UserId:     args[1],
		GiveawayId: getGiveawayForGuild(m.GuildID).Id,
		CreateTime: time.Now(),
		GuildId:    m.GuildID,
		ChannelId:  m.ChannelID,
	}
	participant.GuildName = guild.Name
	participant.UserName = user.Username
	participant.MessageId = *updateThxInfoMessage(nil, m.GuildID, m.ChannelID, args[1], participant.GiveawayId, nil, wait)
	err = DbMap.Insert(participant)
	if err != nil {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
		log.Panicln("OnMessageCreate DbMap.Insert(participant) " + err.Error())
	}
	for err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "✅"); err != nil; err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "✅") {
	}
	for err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "⛔"); err != nil; err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "⛔") {
	}
}

func handleCsrvbotCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) >= 2 {
		switch args[1] {
		case "info":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			printServerInfo(m.ChannelID, m.GuildID)
			return
		case "start":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			finishGiveaway(m.GuildID)
			createMissingGiveaways()
			return
		case "delete":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			if len(args) == 2 {
				_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać ID użytkownika!")
				if err != nil {
					log.Println(err)
				}
				return
			}
			guild, err := session.Guild(m.GuildID)
			if len(m.Mentions) < 1 {
				if err != nil {
					log.Println(m.Author.Username + " usunął ID " + args[2] + " z giveawaya na " + m.GuildID)
					log.Println(err)
					return
				}
				log.Println(m.Author.Username + " usunął ID " + args[2] + " z giveawaya na " + guild.Name)
				deleteFromGiveaway(m.GuildID, args[2])
				_, _ = s.ChannelMessageSend(m.ChannelID, "Usunięto z giveawaya.")
				return
			}
			if err != nil {
				log.Println(m.Author.Username + " usunął " + m.Mentions[0].Username + " z giveawaya na " + m.GuildID)
				log.Println(err)
				return
			}
			log.Println(m.Author.Username + " usunął " + m.Mentions[0].Username + " z giveawaya na " + guild.Name)
			deleteFromGiveaway(m.GuildID, m.Mentions[0].ID)
			_, _ = s.ChannelMessageSend(m.ChannelID, "Usunięto z giveawaya.")
			return
		case "blacklist":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			if len(args) == 2 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Musisz podać użytkownika!")
				return
			}
			guild, err := session.Guild(m.GuildID)
			if len(m.Mentions) < 1 {
				if err != nil {
					log.Println(m.Author.Username + " zblacklistował ID " + args[2] + " na " + m.GuildID)
					log.Println("OnMessageCreate session.Guild(" + m.GuildID + ") " + err.Error())
					return
				}
				log.Println(m.Author.Username + " zblacklistował ID " + args[2] + " na " + guild.Name)
				if blacklistUser(m.GuildID, args[2], m.Author.ID) == nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Użytkownik został zablokowany z możliwości udziału w giveaway.")
				}
				return
			}
			if err != nil {
				log.Println(m.Author.Username + " zblacklistował " + m.Mentions[0].Username + " na " + m.GuildID)
				log.Println("OnMessageCreate session.Guild(" + m.GuildID + ") " + err.Error())
				return
			}
			log.Println(m.Author.Username + " zblacklistował " + m.Mentions[0].Username + " na " + guild.Name)
			if blacklistUser(m.GuildID, m.Mentions[0].ID, m.Author.ID) == nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Użytkownik został zablokowany z możliwości udziału w giveaway.")
			}
			return
		case "unblacklist":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			if len(args) == 2 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Musisz podać użytkownika!")
				return
			}
			guild, err := session.Guild(m.GuildID)
			if len(m.Mentions) < 1 {
				if err != nil {
					log.Println(m.Author.Username + " odblacklistował ID " + args[2] + " na " + m.GuildID)
					log.Println("OnMessageCreate session.Guild(" + m.GuildID + ") " + err.Error())
					return
				}
				log.Println(m.Author.Username + " odblacklistował ID " + args[2] + " na " + guild.Name)
				if unblacklistUser(m.GuildID, args[2]) == nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Użytkownik ponownie może brać udział w giveawayach.")
				}
				return
			}
			if err != nil {
				log.Println(m.Author.Username + " odblacklistował " + m.Mentions[0].Username + " na " + m.GuildID)
				log.Println("OnMessageCreate session.Guild(" + m.GuildID + ") " + err.Error())
				return
			}
			log.Println(m.Author.Username + " odblacklistował " + m.Mentions[0].Username + " na " + guild.Name)
			if unblacklistUser(m.GuildID, m.Mentions[0].ID) == nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Użytkownik ponownie może brać udział w giveawayach.")
			}
			return
		case "setGiveawayChannelName":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			if len(args) == 2 {
				_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać nazwę kanału!")
				if err != nil {
					log.Println(err)
				}
				return
			}
			serverConfig := getServerConfigForGuildId(m.GuildID)
			serverConfig.MainChannel = args[2]
			_, err = DbMap.Update(&serverConfig)
			if err != nil {
				log.Panic("OnMessageCreate DbMap.Update(&serverConfig) " + err.Error())
			}
			_, _ = s.ChannelMessageSend(m.ChannelID, "Ustawiono.")
			return
		case "setBotAdminRoleName":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			if len(args) == 2 {
				_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać nazwę roli!")
				if err != nil {
					log.Println(err)
				}
				return
			}
			serverConfig := getServerConfigForGuildId(m.GuildID)
			serverConfig.AdminRole = args[2]
			_, err = DbMap.Update(&serverConfig)
			if err != nil {
				log.Panic("OnMessageCreate DbMap.Update(&serverConfig) " + err.Error())
			}
			_, _ = s.ChannelMessageSend(m.ChannelID, "Ustawiono.")
			return
		case "setThxInfoChannel":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			if len(args) == 2 {
				_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać kanał!")
				if err != nil {
					log.Println(err)
				}
				return
			}
			serverConfig := getServerConfigForGuildId(m.GuildID)
			if strings.HasPrefix(args[2], "<#") {
				args[2] = args[2][2:]
				args[2] = args[2][:len(args[2])-1]
			}
			serverConfig.ThxInfoChannel = args[2]
			_, err = DbMap.Update(&serverConfig)
			if err != nil {
				log.Panic("OnMessageCreate DbMap.Update(&serverConfig) " + err.Error())
			}
			_, _ = s.ChannelMessageSend(m.ChannelID, "Ustawiono.")
			return
		case "resend":
			embed, err := generateResendEmbed(m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate generateResendEmbed(" + m.Message.Author.ID + ") " + err.Error())
			}
			dm, _ := session.UserChannelCreate(m.Message.Author.ID)
			_, _ = session.ChannelMessageSendEmbed(dm.ID, embed)
			_, _ = s.ChannelMessageSend(m.ChannelID, "Kody zostaly ponownie wyslane :innocent:")
			log.Println("Wysłano resend do " + m.Author.Username + "#" + m.Author.Discriminator)
			return
		case "setHelperRoleName":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			if len(args) == 2 {
				_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać nazwę roli!")
				if err != nil {
					log.Println(err)
				}
				return
			}
			serverConfig := getServerConfigForGuildId(m.GuildID)
			serverConfig.HelperRoleName = args[2]
			_, err = DbMap.Update(&serverConfig)
			if err != nil {
				log.Panic("OnMessageCreate DbMap.Update(&serverConfig) " + err.Error())
			}
			_, _ = s.ChannelMessageSend(m.ChannelID, "Ustawiono.")
			return
		case "setHelperRoleNeededThxAmount":
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
				return
			}
			if !hasAdminPermissions(member, m.GuildID) {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
				return
			}
			if len(args) == 2 {
				_, err := s.ChannelMessageSend(m.ChannelID, "Musisz podać nazwę roli!")
				if err != nil {
					log.Println(err)
				}
				return
			}
			serverConfig := getServerConfigForGuildId(m.GuildID)
			num, err := strconv.Atoi(args[2])
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Ale moze liczbe daj co")
				return
			}
			serverConfig.HelperRoleThxesNeeded = num
			_, err = DbMap.Update(&serverConfig)
			if err != nil {
				log.Panic("OnMessageCreate DbMap.Update(&serverConfig) " + err.Error())
			}
			_, _ = s.ChannelMessageSend(m.ChannelID, "Ustawiono.")
			checkHelpers(m.GuildID)
			return
		}

	}
	_, _ = s.ChannelMessageSend(m.ChannelID, "!csrvbot <delete|resend|start|blacklist|unblacklist|setGiveawayChannelName|setBotAdminRoleName|setThxInfoChannel|info|setHelperRoleName|setHelperRoleNeededThxAmount>")
}

func handleSetwinnerCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Na kogo ustawiamy?")
		return
	}
	_, _ = s.ChannelMessageSend(m.ChannelID, ":ok_hand:")
}

func handleThxmeCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) != 2 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Niepoprawna ilość argumentów.")
		return
	}

	match, _ := regexp.Match("<@[!]?[0-9]*>", []byte(args[1]))
	if !match {
		printGiveawayInfo(m.ChannelID, m.GuildID)
		return
	}

	args[1] = args[1][2 : len(args[1])-1]
	if strings.HasPrefix(args[1], "!") {
		args[1] = args[1][1:]
	}

	if m.Author.ID == args[1] {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Nie możesz poprosić siebie o podziękowanie.")
		return
	}

	user, err := session.User(args[1])
	if err != nil {
		return
	}

	guild, err := session.Guild(m.GuildID)
	if err != nil {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
		log.Println("OnMessageCreate session.Guild(" + m.GuildID + ") " + err.Error())
		return
	}

	if user.Bot {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Nie można prosić o podziękowanie bota!")
		return
	}
	if isBlacklisted(m.GuildID, m.Author.ID) {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Nie możesz poprosić o thx, gdyż jesteś na czarnej liście!")
		return
	}
	candidate := &ParticipantCandidate{
		CandidateName:         m.Author.Username,
		CandidateId:           m.Author.ID,
		CandidateApproverName: user.Username,
		CandidateApproverId:   user.ID,
		GuildName:             guild.Name,
		GuildId:               m.GuildID,
		ChannelId:             m.ChannelID,
	}
	messageId, err := s.ChannelMessageSend(m.ChannelID, user.Mention()+", czy chcesz podziękować użytkownikowi "+m.Author.Mention()+"?")
	if err != nil {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu kandydata do podziekowania :(")
		log.Panicln("OnMessageCreate ChannelMessageSend(candidate) " + err.Error())
	}
	candidate.MessageId = messageId.ID
	err = DbMap.Insert(candidate)
	if err != nil {
		_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu kandydata do podziekowania :(")
		log.Panicln("OnMessageCreate DbMap.Insert(candidate) " + err.Error())
	}
	for err = session.MessageReactionAdd(m.ChannelID, candidate.MessageId, "✅"); err != nil; err = session.MessageReactionAdd(m.ChannelID, candidate.MessageId, "✅") {
	}
	for err = session.MessageReactionAdd(m.ChannelID, candidate.MessageId, "⛔"); err != nil; err = session.MessageReactionAdd(m.ChannelID, candidate.MessageId, "⛔") {
	}
}
