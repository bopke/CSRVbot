package Models

type ServerConfig struct {
	Id          int    `db:"id,primarykey,autoincrement"`
	GuildId     string `db:"guild_id,size:255"`
	AdminRole   string `db:"admin_role,size:255"`
	MainChannel string `db:"main_channel,size:255"`
}
