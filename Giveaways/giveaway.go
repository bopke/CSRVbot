package Giveaways

import (
	"database/sql"
	"time"

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
