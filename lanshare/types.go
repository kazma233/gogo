package main

import (
	"time"

	"github.com/google/uuid"
)

// Type 数据类型枚举
type Type string

// 数据类型常量
const (
	TEXT  Type = "text"
	IMAGE Type = "image"
	OTHER Type = "other"
)

// OP 操作类型
type OP int

// 操作类型常量
const (
	ADD    OP = 1
	REMOVE OP = 2
)

// Item 数据项结构
type Item struct {
	ID         string         `json:"id"`
	Type       Type           `json:"type"`
	Content    string         `json:"content"`
	Meta       map[string]any `json:"meta"`
	CreateTime time.Time      `json:"create_time"`
}

// ItemC 带操作类型的数据项
type ItemC struct {
	Item
	OP OP
}

// NewItem 创建新的数据项
func NewItem(t Type, content string) *Item {
	return &Item{
		ID:         uuid.New().String(),
		Type:       t,
		Content:    content,
		CreateTime: time.Now(),
	}
}

// BS 将操作类型转换为字节
func (op OP) BS() []byte {
	switch op {
	case ADD:
		return []byte{'1'}
	case REMOVE:
		return []byte{'0'}
	default:
		return []byte{'9'}
	}
}
