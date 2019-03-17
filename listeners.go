package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func OnMessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if !isThxMessage(&r.MessageID) {
		return
	}
	if r.UserID == s.State.User.ID {
		return
	}
	member, _ := s.GuildMember(r.GuildID, r.UserID)
	if hasRole(member, config.AdminRole) {
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
		printGiveawayInfo(&m.ChannelID, &m.GuildID)
		return
	}
	if cmds[0] == "csrvbot" {
		_, err := s.ChannelMessageSend(m.ChannelID, "!csrvbot <delete|resend|start|blacklist|info>")
		if err != nil {
			fmt.Println(err)
		}
		if cmds[1] == "info" {
			printServerInfo(&m.ChannelID, &m.GuildID)
			return
		}
		if cmds[1] == "start" {
			//			forceStart <- m.GuildID
			return
		}
		if cmds[1] == "delete" {
			member, err := s.GuildMember(m.GuildID, m.Message.Author.ID)
			if err != nil {
				fmt.Println(err)
				return
			}
			if !hasRole(member, config.AdminRole) {
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
			if !hasRole(member, config.AdminRole) {
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
