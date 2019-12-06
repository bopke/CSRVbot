package Utils

import (
	"csrvbot/Database"
	"csrvbot/ServerConfiguration"
	"errors"
	"github.com/bwmarrin/discordgo"
	"log"
)

func GetRoleID(session *discordgo.Session, guildID string, roleName string) (string, error) {
	guild, err := session.Guild(guildID)
	if err != nil {
		log.Println("Utils GetRoleID Unable to get guild! ", err)
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

func HasRole(session *discordgo.Session, member *discordgo.Member, roleName, guildID string) bool {
	adminRole, err := GetRoleID(session, guildID, roleName)
	if err != nil {
		log.Println("Utils HasRole Unable to get role ID! ", err)
		return false
	}
	for _, role := range member.Roles {
		if role == adminRole {
			return true
		}
	}
	return false
}
func HasPermission(session *discordgo.Session, member *discordgo.Member, guildID string, permission int) bool {
	for _, roleID := range member.Roles {
		role, err := session.State.Role(guildID, roleID)
		if err != nil {
			log.Println("Utils HasPermission Unable to get role from state! ", err)
			return false
		}
		if role.Permissions&permission != 0 {
			return true
		}
	}
	return false
}

func HasAdminPermissions(session *discordgo.Session, member *discordgo.Member, guildID string) bool {
	if HasRole(session, member, GetAdminRoleForGuild(guildID), guildID) || HasPermission(session, member, guildID, discordgo.PermissionAdministrator) { // 8 - administrator
		return true
	}
	return false
}

func GetAdminRoleForGuild(guildId string) string {
	var serverConfig ServerConfiguration.ServerConfig
	err := Database.DbMap.SelectOne(&serverConfig, "SELECT * FROM ServerConfig WHERE guild_id = ?", guildId)
	if err != nil {
		log.Println("Utils GetAdminRoleForGuild Unable to get server configuration! ", err)
		return ""
	}
	return serverConfig.AdminRole
}

func GetSavedRoles(guildId, memberId string) ([]string, error) {
	var memberRoles []MemberRole
	_, err := Database.DbMap.Select(&memberRoles, "SELECT * FROM MemberRoles WHERE guild_id = ? AND member_id = ?", guildId, memberId)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, role := range memberRoles {
		ret = append(ret, role.RoleId)
	}
	return ret, nil
}

func UpdateMemberSavedRoles(member *discordgo.Member, guildId string) {
	savedRoles, err := GetSavedRoles(guildId, member.User.ID)
	if err != nil {
		log.Println("Utils UpdateMemberSavedRoles Unable to get saved roles ", err)
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
			memberrole := MemberRole{GuildId: guildId, RoleId: memberRole, MemberId: member.User.ID}
			err = Database.DbMap.Insert(&memberrole)
			if err != nil {
				log.Println("Utils UpdateAllMembersSavedRoles Unable to insert to database! ", err)
				continue
			}
		}
	}
	for _, savedRole := range savedRoles {
		if savedRole != "" {
			_, err = Database.DbMap.Exec("DELETE FROM MemberRoles WHERE guild_id = ? AND role_id = ? AND member_id = ?", guildId, savedRole, member.User.ID)
			if err != nil {
				log.Println("Utils UpdateAllMembersSavedRoles Unable to delete user saved role ", err)
				continue
			}
		}
	}
}

func RestoreMemberRoles(session *discordgo.Session, member *discordgo.Member, guildId string) {
	var memberRoles []MemberRole
	_, err := Database.DbMap.Select(&memberRoles, "SELECT * FROM MemberRoles WHERE guild_id = ? AND member_id = ?", guildId, member.User.ID)
	if err != nil {
		log.Println("Utils RestoreMemberRoles Unable to get user saved roles ", err)
		return
	}
	for _, role := range memberRoles {
		err = session.GuildMemberRoleAdd(guildId, member.User.ID, role.RoleId)
		if err != nil {
			log.Println("Utils RestoreMemberRoles Unable to restore user role! ", err)
			continue
		}
	}
}

func UpdateAllMembersSavedRoles(session *discordgo.Session, guildId string) {
	guildMembers := GetAllMembers(session, guildId)
	for _, member := range guildMembers {
		UpdateMemberSavedRoles(member, guildId)
	}
}
