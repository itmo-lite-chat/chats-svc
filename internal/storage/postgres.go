package storage

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/itmo-lite-chat/chats-svc/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var namespace = uuid.MustParse("12fb1769-c9f8-4f3d-9f2e-2b7de1d41831")

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	st := &Storage{pool: pool}
	if err := st.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return st, nil
}

func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) CreateChat(ctx context.Context, chatType domain.ChatType, participantIDs []string, title *string, ownerID string) (domain.Chat, []domain.Member, error) {
	chatID := uuid.New().String()
	if chatType == domain.ChatTypePrivate && len(participantIDs) == 2 {
		ids := append([]string(nil), participantIDs...)
		sort.Strings(ids)
		chatID = uuid.NewSHA1(namespace, []byte(ids[0]+":"+ids[1])).String()
	}

	now := time.Now()
	chat := domain.Chat{
		ID:        chatID,
		Type:      chatType,
		Title:     title,
		OwnerID:   ownerID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Chat{}, nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		insert into chats (chat_id, type, title, owner_id, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6)
	`, chat.ID, chat.Type, chat.Title, chat.OwnerID, chat.CreatedAt, chat.UpdatedAt); err != nil {
		return domain.Chat{}, nil, err
	}

	members := make([]domain.Member, 0, len(participantIDs))
	for _, userID := range participantIDs {
		role := domain.MemberRoleMember
		if userID == ownerID {
			role = domain.MemberRoleOwner
		}

		member := domain.Member{
			ChatID:   chat.ID,
			UserID:   userID,
			Role:     role,
			JoinedAt: now,
		}

		if _, err := tx.Exec(ctx, `
			insert into chat_members (chat_id, user_id, role, joined_at, last_read_message_id)
			values ($1, $2, $3, $4, $5)
		`, member.ChatID, member.UserID, member.Role, member.JoinedAt, member.LastReadMessageID); err != nil {
			return domain.Chat{}, nil, err
		}
		members = append(members, member)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Chat{}, nil, err
	}

	return chat, members, nil
}

func (s *Storage) GetOrCreatePrivateChat(ctx context.Context, ownerID, participantID string) (domain.Chat, []domain.Member, bool, error) {
	ids := []string{ownerID, participantID}
	sort.Strings(ids)
	chatID := uuid.NewSHA1(namespace, []byte(ids[0]+":"+ids[1])).String()

	chat, members, err := s.GetChatDetails(ctx, chatID)
	if err == nil {
		return chat, members, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Chat{}, nil, false, err
	}

	chat, members, err = s.CreateChat(ctx, domain.ChatTypePrivate, ids, nil, ownerID)
	if err != nil {
		return domain.Chat{}, nil, false, err
	}
	return chat, members, true, nil
}

func (s *Storage) ListUserChats(ctx context.Context, userID string) ([]domain.Chat, error) {
	rows, err := s.pool.Query(ctx, `
		select c.chat_id, c.type, c.title, c.owner_id, c.last_message_id, c.last_message_preview,
		       c.last_message_at, c.created_at, c.updated_at
		from chats c
		join chat_members cm on cm.chat_id = c.chat_id
		where cm.user_id = $1
		order by coalesce(c.last_message_at, c.updated_at) desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := make([]domain.Chat, 0)
	for rows.Next() {
		chat, err := scanChat(rows)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	return chats, rows.Err()
}

func (s *Storage) GetChatDetails(ctx context.Context, chatID string) (domain.Chat, []domain.Member, error) {
	chat, err := s.GetChat(ctx, chatID)
	if err != nil {
		return domain.Chat{}, nil, err
	}

	rows, err := s.pool.Query(ctx, `
		select chat_id, user_id, role, joined_at, last_read_message_id
		from chat_members
		where chat_id = $1
		order by joined_at
	`, chatID)
	if err != nil {
		return domain.Chat{}, nil, err
	}
	defer rows.Close()

	members := make([]domain.Member, 0)
	for rows.Next() {
		member, err := scanMember(rows)
		if err != nil {
			return domain.Chat{}, nil, err
		}
		members = append(members, member)
	}
	return chat, members, rows.Err()
}

func (s *Storage) GetChat(ctx context.Context, chatID string) (domain.Chat, error) {
	row := s.pool.QueryRow(ctx, `
		select chat_id, type, title, owner_id, last_message_id, last_message_preview,
		       last_message_at, created_at, updated_at
		from chats
		where chat_id = $1
	`, chatID)
	return scanChat(row)
}

func (s *Storage) CheckChatMember(ctx context.Context, chatID, userID string) (domain.Member, bool, error) {
	row := s.pool.QueryRow(ctx, `
		select chat_id, user_id, role, joined_at, last_read_message_id
		from chat_members
		where chat_id = $1 and user_id = $2
	`, chatID, userID)
	member, err := scanMember(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Member{}, false, nil
	}
	if err != nil {
		return domain.Member{}, false, err
	}
	return member, true, nil
}

func (s *Storage) UpdateLastReadMessage(ctx context.Context, chatID, userID string, messageID int64) error {
	_, err := s.pool.Exec(ctx, `
		update chat_members
		set last_read_message_id = $3
		where chat_id = $1 and user_id = $2
	`, chatID, userID, messageID)
	return err
}

func (s *Storage) TouchChatLastMessage(ctx context.Context, chatID string, messageID int64, preview string) (domain.Chat, error) {
	now := time.Now()
	row := s.pool.QueryRow(ctx, `
		update chats
		set last_message_id = $2, last_message_preview = $3, last_message_at = $4, updated_at = $4
		where chat_id = $1
		returning chat_id, type, title, owner_id, last_message_id, last_message_preview,
		          last_message_at, created_at, updated_at
	`, chatID, messageID, preview, now)
	return scanChat(row)
}

func (s *Storage) migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		create table if not exists chats (
			chat_id text primary key,
			type integer not null,
			title text,
			owner_id text not null,
			last_message_id bigint not null default 0,
			last_message_preview text not null default '',
			last_message_at timestamptz,
			created_at timestamptz not null,
			updated_at timestamptz not null
		);

		create table if not exists chat_members (
			chat_id text not null references chats(chat_id) on delete cascade,
			user_id text not null,
			role integer not null,
			joined_at timestamptz not null,
			last_read_message_id bigint not null default 0,
			primary key (chat_id, user_id)
		);
	`)
	return err
}

type row interface {
	Scan(dest ...any) error
}

func scanChat(r row) (domain.Chat, error) {
	var chat domain.Chat
	err := r.Scan(&chat.ID, &chat.Type, &chat.Title, &chat.OwnerID, &chat.LastMessageID, &chat.LastMessagePreview, &chat.LastMessageAt, &chat.CreatedAt, &chat.UpdatedAt)
	return chat, err
}

func scanMember(r row) (domain.Member, error) {
	var member domain.Member
	err := r.Scan(&member.ChatID, &member.UserID, &member.Role, &member.JoinedAt, &member.LastReadMessageID)
	return member, err
}
