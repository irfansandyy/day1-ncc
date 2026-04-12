package repositories

import (
	"context"
	"database/sql"
	"errors"

	"app-backend/models"
)

var ErrChatNotFound = errors.New("chat not found")

type ChatRepository interface {
	CreateChat(ctx context.Context, userID int64, title string) (models.Chat, error)
	ListChatsByUser(ctx context.Context, userID int64) ([]models.Chat, error)
	GetChatByID(ctx context.Context, chatID, userID int64) (models.Chat, error)
	CreateMessage(ctx context.Context, chatID int64, role, content string) (models.Message, error)
	ListMessagesByChat(ctx context.Context, chatID, userID int64, limit int) ([]models.Message, error)
	UpdateChatTimestamp(ctx context.Context, chatID int64) error
}

type PostgresChatRepository struct {
	db *sql.DB
}

func NewPostgresChatRepository(db *sql.DB) *PostgresChatRepository {
	return &PostgresChatRepository{db: db}
}

func (r *PostgresChatRepository) CreateChat(ctx context.Context, userID int64, title string) (models.Chat, error) {
	query := `
		INSERT INTO chats (user_id, title)
		VALUES ($1, $2)
		RETURNING id, user_id, title, created_at, updated_at`

	var chat models.Chat
	err := r.db.QueryRowContext(ctx, query, userID, title).
		Scan(&chat.ID, &chat.UserID, &chat.Title, &chat.CreatedAt, &chat.UpdatedAt)
	if err != nil {
		return models.Chat{}, err
	}

	return chat, nil
}

func (r *PostgresChatRepository) ListChatsByUser(ctx context.Context, userID int64) ([]models.Chat, error) {
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM chats
		WHERE user_id = $1
		ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := make([]models.Chat, 0)
	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(&chat.ID, &chat.UserID, &chat.Title, &chat.CreatedAt, &chat.UpdatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *PostgresChatRepository) GetChatByID(ctx context.Context, chatID, userID int64) (models.Chat, error) {
	query := `
		SELECT id, user_id, title, created_at, updated_at
		FROM chats
		WHERE id = $1 AND user_id = $2`

	var chat models.Chat
	err := r.db.QueryRowContext(ctx, query, chatID, userID).
		Scan(&chat.ID, &chat.UserID, &chat.Title, &chat.CreatedAt, &chat.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Chat{}, ErrChatNotFound
		}
		return models.Chat{}, err
	}

	return chat, nil
}

func (r *PostgresChatRepository) CreateMessage(ctx context.Context, chatID int64, role, content string) (models.Message, error) {
	query := `
		INSERT INTO messages (chat_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING id, chat_id, role, content, created_at`

	var message models.Message
	err := r.db.QueryRowContext(ctx, query, chatID, role, content).
		Scan(&message.ID, &message.ChatID, &message.Role, &message.Content, &message.CreatedAt)
	if err != nil {
		return models.Message{}, err
	}

	return message, nil
}

func (r *PostgresChatRepository) ListMessagesByChat(ctx context.Context, chatID, userID int64, limit int) ([]models.Message, error) {
	if _, err := r.GetChatByID(ctx, chatID, userID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, chat_id, role, content, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC`

	args := []any{chatID}
	if limit > 0 {
		query += ` LIMIT $2`
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]models.Message, 0)
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *PostgresChatRepository) UpdateChatTimestamp(ctx context.Context, chatID int64) error {
	query := `UPDATE chats SET updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, chatID)
	return err
}
