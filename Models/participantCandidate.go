package Models

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
)

type ParticipantCandidate struct {
	Id                    int            `db:"id, primarykey, autoincrement"`
	CandidateId           string         `db:"candidate_id,size:255"`
	CandidateName         string         `db:"candidate_name,size:255"`
	CandidateApproverId   string         `db:"candidate_approver_id,size:255"`
	CandidateApproverName string         `db:"candidate_approver_name,size:255"`
	GuildId               string         `db:"guild_id,size:255"`
	GuildName             string         `db:"guild_name,size:255"`
	MessageId             string         `db:"message_id,size:255"`
	ChannelId             string         `db:"channel_id,size:255"`
	IsAccepted            sql.NullBool   `db:"is_accepted"`
	AcceptTime            mysql.NullTime `db:"accept_time"`
}
