package grpc_api

import (
	"context"

	"github.com/itmo-lite-chat/chats-svc/internal/domain"
	"github.com/itmo-lite-chat/chats-svc/internal/service"
	pb "github.com/itmo-lite-chat/proto-registry/gen/services/chats_service/chats/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type API struct {
	pb.UnimplementedChatsServiceServer
	service *service.Service
}

func New(service *service.Service) *API {
	return &API{service: service}
}

func (a *API) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	chat, _, err := a.service.CreateChat(ctx, fromProtoChatType(req.GetType()), req.GetParticipantIds(), req.Title, req.GetOwnerId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateChatResponse{Chat: toProtoChat(chat)}, nil
}

func (a *API) GetOrCreatePrivateChat(ctx context.Context, req *pb.GetOrCreatePrivateChatRequest) (*pb.GetOrCreatePrivateChatResponse, error) {
	chat, members, created, err := a.service.GetOrCreatePrivateChat(ctx, req.GetOwnerId(), req.GetParticipantId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetOrCreatePrivateChatResponse{Chat: toProtoChat(chat), Members: toProtoMembers(members), Created: created}, nil
}

func (a *API) ListUserChats(ctx context.Context, req *pb.ListUserChatsRequest) (*pb.ListUserChatsResponse, error) {
	chats, err := a.service.ListUserChats(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.ListUserChatsResponse{Chats: make([]*pb.Chat, 0, len(chats))}
	for _, chat := range chats {
		resp.Chats = append(resp.Chats, toProtoChat(chat))
	}
	return resp, nil
}

func (a *API) GetChatDetails(ctx context.Context, req *pb.GetChatDetailsRequest) (*pb.GetChatDetailsResponse, error) {
	chat, members, err := a.service.GetChatDetails(ctx, req.GetChatId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "chat not found")
	}
	return &pb.GetChatDetailsResponse{Chat: toProtoChat(chat), Members: toProtoMembers(members)}, nil
}

func (a *API) CheckChatMember(ctx context.Context, req *pb.CheckChatMemberRequest) (*pb.CheckChatMemberResponse, error) {
	member, ok, err := a.service.CheckChatMember(ctx, req.GetChatId(), req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CheckChatMemberResponse{IsMember: ok, Member: toProtoMember(member)}, nil
}

func (a *API) UpdateLastReadMessage(ctx context.Context, req *pb.UpdateLastReadMessageRequest) (*pb.UpdateLastReadMessageResponse, error) {
	if err := a.service.UpdateLastReadMessage(ctx, req.GetChatId(), req.GetUserId(), req.GetLastReadMessageId()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UpdateLastReadMessageResponse{}, nil
}

func (a *API) TouchChatLastMessage(ctx context.Context, req *pb.TouchChatLastMessageRequest) (*pb.TouchChatLastMessageResponse, error) {
	chat, err := a.service.TouchChatLastMessage(ctx, req.GetChatId(), req.GetLastMessageId(), req.GetLastMessagePreview())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TouchChatLastMessageResponse{Chat: toProtoChat(chat)}, nil
}

func toProtoChat(chat domain.Chat) *pb.Chat {
	pbChat := &pb.Chat{
		ChatId:             chat.ID,
		Type:               toProtoChatType(chat.Type),
		Title:              chat.Title,
		OwnerId:            chat.OwnerID,
		LastMessageId:      chat.LastMessageID,
		LastMessagePreview: chat.LastMessagePreview,
		CreatedAt:          timestamppb.New(chat.CreatedAt),
		UpdatedAt:          timestamppb.New(chat.UpdatedAt),
	}
	if chat.LastMessageAt != nil {
		pbChat.LastMessageAt = timestamppb.New(*chat.LastMessageAt)
	}
	return pbChat
}

func toProtoMember(member domain.Member) *pb.ChatMember {
	return &pb.ChatMember{
		ChatId:            member.ChatID,
		UserId:            member.UserID,
		Role:              toProtoMemberRole(member.Role),
		JoinedAt:          timestamppb.New(member.JoinedAt),
		LastReadMessageId: member.LastReadMessageID,
	}
}

func toProtoMembers(members []domain.Member) []*pb.ChatMember {
	result := make([]*pb.ChatMember, 0, len(members))
	for _, member := range members {
		result = append(result, toProtoMember(member))
	}
	return result
}

func fromProtoChatType(chatType pb.ChatType) domain.ChatType {
	switch chatType {
	case pb.ChatType_CHAT_TYPE_GROUP:
		return domain.ChatTypeGroup
	default:
		return domain.ChatTypePrivate
	}
}

func toProtoChatType(chatType domain.ChatType) pb.ChatType {
	switch chatType {
	case domain.ChatTypeGroup:
		return pb.ChatType_CHAT_TYPE_GROUP
	default:
		return pb.ChatType_CHAT_TYPE_PRIVATE
	}
}

func toProtoMemberRole(role domain.MemberRole) pb.MemberRole {
	switch role {
	case domain.MemberRoleOwner:
		return pb.MemberRole_MEMBER_ROLE_OWNER
	default:
		return pb.MemberRole_MEMBER_ROLE_MEMBER
	}
}
