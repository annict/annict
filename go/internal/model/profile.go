package model

import "database/sql"

// Profile はユーザープロフィールのドメインエンティティ
type Profile struct {
	ID          ProfileID
	UserID      UserID
	Name        string
	Description string
	URL         sql.NullString
	ImageData   sql.NullString
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
}
