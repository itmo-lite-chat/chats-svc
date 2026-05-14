package service

import (
	"context"

	"github.com/itmo-lite-chat/chats-svc/internal/domain"
)

type Storage interface {
	CreateChat(ctx context.Context, chatType domain.ChatType, participantIDs []string, title *string, ownerID string) (domain.Chat, []domain.Member, error)
	GetOrCreatePrivateChat(ctx context.Context, ownerID, participantID string) (domain.Chat, []domain.Member, bool, error)
	ListUserChats(ctx context.Context, userID string) ([]domain.Chat, error)
	GetChatDetails(ctx context.Context, chatID string) (domain.Chat, []domain.Member, error)
	CheckChatMember(ctx context.Context, chatID, userID string) (domain.Member, bool, error)
	UpdateLastReadMessage(ctx context.Context, chatID, userID string, messageID int64) error
	TouchChatLastMessage(ctx context.Context, chatID string, messageID int64, preview string) (domain.Chat, error)
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) CreateChat(ctx context.Context, chatType domain.ChatType, participantIDs []string, title *string, ownerID string) (domain.Chat, []domain.Member, error) {
	return s.storage.CreateChat(ctx, chatType, participantIDs, title, ownerID)
}

func (s *Service) GetOrCreatePrivateChat(ctx context.Context, ownerID, participantID string) (domain.Chat, []domain.Member, bool, error) {
	return s.storage.GetOrCreatePrivateChat(ctx, ownerID, participantID)
}

func (s *Service) ListUserChats(ctx context.Context, userID string) ([]domain.Chat, error) {
	return s.storage.ListUserChats(ctx, userID)
}

func (s *Service) GetChatDetails(ctx context.Context, chatID string) (domain.Chat, []domain.Member, error) {
	return s.storage.GetChatDetails(ctx, chatID)
}

func (s *Service) CheckChatMember(ctx context.Context, chatID, userID string) (domain.Member, bool, error) {
	return s.storage.CheckChatMember(ctx, chatID, userID)
}

func (s *Service) UpdateLastReadMessage(ctx context.Context, chatID, userID string, messageID int64) error {
	return s.storage.UpdateLastReadMessage(ctx, chatID, userID, messageID)
}

func (s *Service) TouchChatLastMessage(ctx context.Context, chatID string, messageID int64, preview string) (domain.Chat, error) {
	return s.storage.TouchChatLastMessage(ctx, chatID, messageID, preview)
}
