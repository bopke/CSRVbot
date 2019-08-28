package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func OnMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if !isThxMessage(r.MessageID) {
		return
	}
	if r.UserID == s.State.User.ID {
		return
	}
	member, _ := s.GuildMember(r.GuildID, r.UserID)
	if hasAdminPermissions(member, r.GuildID) && (r.Emoji.Name == "✅" || r.Emoji.Name == "⛔") {
		reactionists, _ := session.MessageReactions(r.ChannelID, r.MessageID, "⛔", 10)
		for _, user := range reactionists {
			if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "⛔") {
				continue
			}
			_ = s.MessageReactionRemove(r.ChannelID, r.MessageID, "⛔", user.ID)
		}
		reactionists, _ = session.MessageReactions(r.ChannelID, r.MessageID, "✅", 10)
		for _, user := range reactionists {
			if user.ID == session.State.User.ID || (user.ID == r.UserID && r.MessageReaction.Emoji.Name == "✅") {
				continue
			}
			_ = s.MessageReactionRemove(r.ChannelID, r.MessageID, "✅", user.ID)
		}
		participant := getParticipantByMessageId(r.MessageID)
		participant.AcceptTime.Time = time.Now()
		participant.AcceptTime.Valid = true
		participant.AcceptUser.String = member.User.Username
		participant.AcceptUser.Valid = true
		participant.AcceptUserId.String = r.UserID
		participant.AcceptUserId.Valid = true
		participant.IsAccepted.Valid = true
		if r.Emoji.Name == "✅" {
			log.Println(member.User.Username + "(" + member.User.ID + ") zaakceptował udział " + participant.UserName + "(" + participant.UserId + ") w giveawayu o ID " + fmt.Sprintf("%d", participant.GiveawayId))
			participant.IsAccepted.Bool = true
			_, err := DbMap.Update(participant)
			if err != nil {
				log.Panicln(err)
			}
			updateThxInfoMessage(&r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, &r.UserID, confirm)
		} else if r.Emoji.Name == "⛔" {
			log.Println(member.User.Username + "(" + member.User.ID + ") odrzucił udział " + participant.UserName + "(" + participant.UserId + ") w giveawayu o ID " + fmt.Sprintf("%d", participant.GiveawayId))
			participant.IsAccepted.Bool = false
			_, err := DbMap.Update(participant)
			if err != nil {
				log.Panicln("OnMessageReactionAdd DbMap.Update(participant) " + err.Error())
			}
			updateThxInfoMessage(&r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, &r.UserID, reject)
		}
		return
	} else {
		_ = s.MessageReactionRemove(r.ChannelID, r.MessageID, r.Emoji.Name, r.UserID)
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

	if m.Content == "!" {
		return
	}
	// remove prefix
	m.Content = m.Content[1:]

	args := strings.Fields(m.Content)
	switch args[0] {
	case "thx":
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
		user, _ := session.User(args[1])
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
		participant.MessageId = *updateThxInfoMessage(nil, m.ChannelID, args[1], participant.GiveawayId, nil, wait)
		err = DbMap.Insert(participant)
		if err != nil {
			_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
			log.Panicln("OnMessageCreate DbMap.Insert(participant) " + err.Error())
		}
		for err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "✅"); err != nil; err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "✅") {
		}
		for err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "⛔"); err != nil; err = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "⛔") {
		}
	case "giveaway":
		printGiveawayInfo(m.ChannelID, m.GuildID)
	case "csrvbot":
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
			}
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, "!csrvbot <delete|resend|start|blacklist|unblacklist|setGiveawayChannelName|setBotAdminRoleName|info>")
	case "setwinner":
		if len(args) == 1 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Na kogo ustawiamy?")
			return
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, ":ok_hand:")
	}
}

func OnGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	log.Printf("Zarejestrowałem utworzenie gildii")
	createConfigurationIfNotExists(g.Guild.ID)
	createMissingGiveaways()
}
