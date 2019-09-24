package main

import (
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"
)

func getRoleID(guildID string, roleName string) (string, error) {
	guild, err := session.Guild(guildID)
	if err != nil {
		log.Println("getRoleID session.Guild(" + guildID + ") " + err.Error())
		return "", err
	}
	roles := guild.Roles
	for _, role := range roles {
		if role.Name == roleName {
			return role.ID, nil
		}
	}
	return "", errors.New("no " + roleName + " role available")
}

func hasRole(member *discordgo.Member, roleName, guildID string) bool {
	//z jakiegos powodu w strukturze member GuildID jest puste...
	adminRole, err := getRoleID(guildID, roleName)
	if err != nil {
		log.Println("hasRole getRoleID(" + guildID + ", " + roleName + ") " + err.Error())
		return false
	}
	for _, role := range member.Roles {
		if role == adminRole {
			return true
		}
	}
	return false
}
func hasPermission(member *discordgo.Member, guildID string, permission int) bool {
	for _, roleID := range member.Roles {
		role, err := session.State.Role(guildID, roleID)
		if err != nil {
			log.Println("hasPermisson session.State.Role(" + guildID + ", " + roleID + ") " + err.Error())
			return false
		}
		if role.Permissions&permission != 0 {
			return true
		}
	}
	return false
}

func hasAdminPermissions(member *discordgo.Member, guildID string) bool {
	if hasRole(member, getAdminRoleForGuild(guildID), guildID) || hasPermission(member, guildID, 8) { // 8 - administrator
		return true
	}
	return false
}

func getAdminRoleForGuild(guildID string) string {
	var serverConfig ServerConfig
	err := DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig")
	if err != nil {
		log.Println("getAdminRoleForGuild(" + guildID + ") " + err.Error())
		return ""
	}
	return serverConfig.AdminRole
}

func getSavedRoles(guildId, memberId string) ([]string, error) {
	var memberRoles []MemberRole
	_, err := DbMap.Select(&memberRoles, "SELECT * FROM MemberRoles WHERE guild_id = ? AND member_id = ?", guildId, memberId)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, role := range memberRoles {
		ret = append(ret, role.RoleId)
	}
	return ret, nil
}

func updateMemberSavedRoles(member *discordgo.Member) {
	savedRoles, err := getSavedRoles(member.GuildID, member.User.ID)
	if err != nil {
		log.Println("updateMemberSavedRoles error getting saved roles " + err.Error())
		return
	}
	for _, memberRole := range member.Roles {
		found := false
		for i, savedRole := range savedRoles {
			if savedRole == memberRole {
				found = true
				savedRoles[i] = ""
				break
			}
		}
		if !found {
			memberrole := MemberRole{GuildId: member.GuildID, RoleId: memberRole, MemberId: member.User.ID}
			err = DbMap.Insert(&memberrole)
			if err != nil {
				log.Println("updateAllMembersInfo error saving new role info " + err.Error())
				continue
			}
		}
	}
	for _, savedRole := range savedRoles {
		if savedRole != "" {
			_, err = DbMap.Exec("DELETE FROM MemberRoles WHERE guild_id = ? AND role_id = ? AND member_id = ?", member.GuildID, savedRole, member.User.ID)
			if err != nil {
				log.Println("updateAllMembersInfo error deleting info about member role " + err.Error())
				continue
			}
		}
	}
}

func restoreMemberRoles(member *discordgo.Member) {
	var memberRoles []MemberRole
	_, err := DbMap.Select(&memberRoles, "SELECT * FROM MemberRoles WHERE guild_id = ? AND member_id = ?", member.GuildID, member.User.ID)
	if err != nil {
		log.Println("restoreMemberRoles error getting saved roles " + err.Error())
		return
	}
	for _, role := range memberRoles {
		err = session.GuildMemberRoleAdd(member.GuildID, member.User.ID, role.RoleId)
		if err != nil {
			log.Println("restoreMemberRoles error restoring member role " + err.Error())
			continue
		}
	}
}
