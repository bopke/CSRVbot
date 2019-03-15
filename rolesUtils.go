package main

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

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
