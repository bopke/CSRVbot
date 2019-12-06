package Utils

type MemberRole struct {
	Id       int    `db:"id,primarykey,autoincrement"`
	GuildId  string `db:"guild_id,size:255"`
	MemberId string `db:"member_id,size:255"`
	RoleId   string `db:"role_id,size:255"`
}
