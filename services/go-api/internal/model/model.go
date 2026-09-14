package model

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OpenID    string    `gorm:"size:128;not null;uniqueIndex" json:"-"`
	UnionID   string    `gorm:"size:128;index" json:"-"`
	Nickname  string    `gorm:"size:100" json:"nickname"`
	AvatarURL string    `gorm:"size:500" json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Book struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	OwnerUserID uint      `gorm:"not null;index" json:"ownerUserId"`
	Name        string    `gorm:"size:80;not null" json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type BookMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BookID    uint      `gorm:"not null;uniqueIndex:idx_book_user" json:"bookId"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_book_user;index" json:"userId"`
	Role      string    `gorm:"size:20;not null" json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BookID    uint      `gorm:"not null;uniqueIndex:idx_book_category" json:"bookId"`
	Type      string    `gorm:"size:20;not null;uniqueIndex:idx_book_category;index" json:"type"`
	Name      string    `gorm:"size:80;not null;uniqueIndex:idx_book_category" json:"name"`
	SortOrder int       `gorm:"not null;default:0" json:"sortOrder"`
	IsSystem  bool      `gorm:"not null;default:false" json:"isSystem"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Account struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BookID    uint      `gorm:"not null;uniqueIndex:idx_book_account" json:"bookId"`
	Name      string    `gorm:"size:80;not null;uniqueIndex:idx_book_account" json:"name"`
	SortOrder int       `gorm:"not null;default:0" json:"sortOrder"`
	IsSystem  bool      `gorm:"not null;default:false" json:"isSystem"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Transaction struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	BookID       uint      `gorm:"not null;uniqueIndex:idx_book_client;index:idx_book_date" json:"bookId"`
	ClientID     string    `gorm:"size:128;not null;uniqueIndex:idx_book_client" json:"clientId"`
	Type         string    `gorm:"size:20;not null;index" json:"type"`
	AmountFen    int64     `gorm:"not null" json:"amountFen"`
	CategoryID   uint      `gorm:"not null;index" json:"categoryId"`
	CategoryName string    `gorm:"size:80;not null" json:"category"`
	AccountID    uint      `gorm:"not null;index" json:"accountId"`
	AccountName  string    `gorm:"size:80;not null" json:"account"`
	TxDate       string    `gorm:"column:tx_date;size:10;not null;index:idx_book_date" json:"date"`
	Note         string    `gorm:"size:500" json:"note"`
	Source       string    `gorm:"size:40;not null;default:'manual';index" json:"source"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Budget struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BookID    uint      `gorm:"not null;uniqueIndex:idx_book_month" json:"bookId"`
	Month     string    `gorm:"size:7;not null;uniqueIndex:idx_book_month" json:"month"`
	AmountFen int64     `gorm:"not null;default:0" json:"amountFen"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
