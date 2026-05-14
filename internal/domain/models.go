package domain

import "time"

type ChatType int32

const (
	ChatTypePrivate ChatType = 1
	ChatTypeGroup   ChatType = 2
)

type MemberRole int32

const (
	MemberRoleOwner  MemberRole = 1
	MemberRoleMember MemberRole = 3
)

type Chat struct {
	ID                 string
	Type               ChatType
	Title              *string
	OwnerID            string
	LastMessageID      int64
	LastMessagePreview string
	LastMessageAt      *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Member struct {
	ChatID            string
	UserID            string
	Role              MemberRole
	JoinedAt          time.Time
	LastReadMessageID int64
}
