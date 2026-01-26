package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

type MemoryStorage struct {
	l            []*Item
	storageFile  *os.File
	lock         sync.RWMutex
	autoSaveChan chan ItemC
}

func New() (*MemoryStorage, error) {
	fp := filepath.Join("./download", "metadata")
	log.Printf("MemoryStorage path %s", fp)

	dirPath := filepath.Dir(fp)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	h := &MemoryStorage{
		l:            make([]*Item, 0),
		storageFile:  f,
		autoSaveChan: make(chan ItemC, 100),
	}

	go h.autoHandle()

	return h, nil
}

func (s *MemoryStorage) autoHandle() {
	for {
		select {
		case ic := <-s.autoSaveChan:
			err := s.Save(ic)
			if err != nil {
				log.Printf("auto save error: %v", err)
			} else {
				log.Printf("auto save success: %s, op: %d", ic.ID, ic.OP)
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// IDataOperator

func (s *MemoryStorage) Add(item *Item) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.l = append(s.l, item)
	s.autoSaveChan <- ItemC{Item: *item, OP: ADD}
}

func (s *MemoryStorage) Remove(id string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.l = slices.DeleteFunc(s.l, func(item *Item) bool {
		return strings.EqualFold(item.ID, id)
	})
	s.autoSaveChan <- ItemC{Item: Item{ID: id}, OP: REMOVE}
}

func (s *MemoryStorage) Get(id string) *Item {
	s.lock.RLock()
	defer s.lock.RUnlock()

	index := slices.IndexFunc(s.l, func(item *Item) bool {
		return strings.EqualFold(item.ID, id)
	})
	if index == -1 {
		return nil
	}

	return s.l[index]
}

func (s *MemoryStorage) List() []*Item {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.l
}

func (s *MemoryStorage) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.l)
}

// IStorage

func (s *MemoryStorage) Load() (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// 重置文件指针到开头
	_, err = s.storageFile.Seek(0, 0)
	if err != nil {
		return err
	}

	// 创建一个scanner来读取文件
	scanner := bufio.NewScanner(s.storageFile)

	// 逐行读取文件内容
	for scanner.Scan() {
		var item Item
		fbs := scanner.Bytes()
		t := fbs[0]
		data := fbs[1:]
		switch t {
		case '1':
			err := json.Unmarshal(data, &item)
			if err != nil {
				return err
			}
			s.l = append(s.l, &item)
		case '0':
			id := string(data)
			s.l = slices.DeleteFunc(s.l, func(item *Item) bool {
				return strings.EqualFold(item.ID, id)
			})
		default:
			// Unknown operation type, skip
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (s *MemoryStorage) Save(ic ItemC) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, err := s.storageFile.Write(ic.OP.BS())
	if err != nil {
		return err
	}
	// always write the item, even if it's a remove operation
	// at load time, we will filter out the remove operations
	switch ic.OP {
	case ADD:
		encoder := json.NewEncoder(s.storageFile)
		if err := encoder.Encode(ic.Item); err != nil {
			return err
		}
	case REMOVE:
		// For remove operation, we just write the ID
		if _, err := s.storageFile.WriteString(ic.Item.ID + "\n"); err != nil {
			return err
		}
	default:
		log.Printf("Unknown operation: %d", ic.OP)
		return nil
	}

	return s.storageFile.Sync()
}

// CleanupExpired 清理过期数据 (旧存储实现的空方法)
func (s *MemoryStorage) CleanupExpired(maxAge time.Duration) {
	// 旧的存储实现不支持清理功能
	log.Printf("CleanupExpired called on legacy MemoryStorage - not implemented")
}

// Close 关闭存储
func (s *MemoryStorage) Close() error {
	return s.storageFile.Close()
}
