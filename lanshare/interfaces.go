package main

import "time"

// Storage 通用存储接口
type Storage interface {
	IDataOperator
	IStorage
	Close() error
	CleanupExpired(maxAge time.Duration)
}

// IDataOperator 数据操作接口
type IDataOperator interface {
	Add(item *Item)
	Remove(id string)
	Get(id string) *Item
	List() []*Item
	Len() int
}

// IStorage 持久化接口
type IStorage interface {
	Save() error
	Load() error
}
