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
	if hasRole(member, config.AdminRole, r.GuildID) && (r.Emoji.Name == "✅" || r.Emoji.Name == "⛔") {
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
			updateThxInfoMessage(&r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, confirm)
		} else if r.Emoji.Name == "⛔" {
			log.Println(member.User.Username + "(" + member.User.ID + ") odrzucił udział " + participant.UserName + "(" + participant.UserId + ") w giveawayu o ID " + fmt.Sprintf("%d", participant.GiveawayId))
			participant.IsAccepted.Bool = false
			_, err := DbMap.Update(participant)
			if err != nil {
				log.Panicln("OnMessageReactionAdd DbMap.Update(participant) " + err.Error())
			}
			updateThxInfoMessage(&r.MessageID, r.ChannelID, participant.UserId, participant.GiveawayId, reject)
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
		if isBlacklisted(args[1], m.GuildID) {
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
		participant.MessageId = *updateThxInfoMessage(nil, m.ChannelID, args[1], participant.GiveawayId, wait)
		err = DbMap.Insert(participant)
		if err != nil {
			_, _ = session.ChannelMessageSend(m.ChannelID, "Coś poszło nie tak przy dodawaniu podziękowania :(")
			log.Panicln("OnMessageCreate DbMap.Insert(participant) " + err.Error())
		}
		_ = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "✅")
		_ = session.MessageReactionAdd(m.ChannelID, participant.MessageId, "⛔")
	case "giveaway":
		printGiveawayInfo(m.ChannelID, m.GuildID)
	case "csrvbot":
		if len(args) == 2 {
			switch args[1] {
			case "info":
				member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
				if err != nil {
					log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
					return
				}
				if !hasRole(member, config.AdminRole, m.GuildID) {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
					return
				}
				printServerInfo(m.ChannelID, m.GuildID)
			case "start":
				member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
				if err != nil {
					log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
					return
				}
				if !hasRole(member, config.AdminRole, m.GuildID) {
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
				if !hasRole(member, config.AdminRole, m.GuildID) {
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
				if err != nil {
					log.Println(m.Author.Username + " usunął " + member.User.Username + " z giveawaya na " + m.GuildID)
					log.Println(err)
					return
				}
				log.Println(m.Author.Username + " usunął " + member.User.Username + " z giveawaya na " + guild.Name)
				deleteFromGiveaway(args[2], m.GuildID)
				return
			case "blacklist":
				member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
				if err != nil {
					log.Println("OnMessageCreate s.GuildMember(" + m.GuildID + ", " + m.Message.Author.ID + ") " + err.Error())
					return
				}
				if !hasRole(member, config.AdminRole, m.GuildID) {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Brak uprawnień.")
					return
				}
				if len(args) == 2 {
					_, _ = s.ChannelMessageSend(m.ChannelID, "Musisz podać ID użytkownika!")
					return
				}
				guild, err := session.Guild(m.GuildID)
				if err != nil {
					log.Println(m.Author.Username + " zblacklistował " + member.User.Username + " na " + m.GuildID)
					log.Println("OnMessageCreate session.Guild(" + m.GuildID + ") " + err.Error())
					return
				}
				log.Println(m.Author.Username + " zblacklistował " + member.User.Username + " na " + guild.Name)
				blacklistUser(args[2], m.GuildID)
				return
			}
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, "!csrvbot <delete|resend|start|blacklist|info>")
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
	createMissingGiveaways()
}
