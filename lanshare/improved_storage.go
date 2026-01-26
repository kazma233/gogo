package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type ImprovedMemoryStorage struct {
	items        map[string]*Item
	storageFile  *os.File
	lock         sync.RWMutex
	autoSaveChan chan ItemC
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func NewImprovedStorage(storagePath string) (*ImprovedMemoryStorage, error) {
	fp := filepath.Join(storagePath, "metadata.jsonl")
	log.Printf("ImprovedMemoryStorage path %s", fp)

	dirPath := filepath.Dir(fp)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open storage file: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &ImprovedMemoryStorage{
		items:        make(map[string]*Item),
		storageFile:  f,
		autoSaveChan: make(chan ItemC, 100),
		ctx:          ctx,
		cancel:       cancel,
	}

	s.wg.Add(1)
	go s.autoSaveWorker()

	return s, nil
}

func (s *ImprovedMemoryStorage) autoSaveWorker() {
	defer s.wg.Done()

	saveTicker := time.NewTicker(5 * time.Second)
	defer saveTicker.Stop()

	// 每天清理一次过期数据
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			// 处理剩余的保存任务
			for {
				select {
				case ic := <-s.autoSaveChan:
					if err := s.persistItem(ic); err != nil {
						log.Printf("Failed to save item during shutdown: %v", err)
					}
				default:
					return
				}
			}
		case ic := <-s.autoSaveChan:
			if err := s.persistItem(ic); err != nil {
				log.Printf("Auto save error: %v", err)
			}
		case <-saveTicker.C:
			// 定期同步文件
			if err := s.storageFile.Sync(); err != nil {
				log.Printf("File sync error: %v", err)
			}
		case <-cleanupTicker.C:
			// 定期清理30天前的数据
			s.CleanupExpired(30 * 24 * time.Hour)
		}
	}
}

func (s *ImprovedMemoryStorage) persistItem(ic ItemC) error {
	data := map[string]interface{}{
		"op":   ic.OP,
		"item": ic.Item,
		"ts":   time.Now().Unix(),
	}

	encoder := json.NewEncoder(s.storageFile)
	return encoder.Encode(data)
}

func (s *ImprovedMemoryStorage) Close() error {
	s.cancel()
	s.wg.Wait()
	return s.storageFile.Close()
}

// 添加清理过期数据的方法
// 实现 IDataOperator 接口
func (s *ImprovedMemoryStorage) Add(item *Item) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items[item.ID] = item
	s.autoSaveChan <- ItemC{Item: *item, OP: ADD}
}

func (s *ImprovedMemoryStorage) Remove(id string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.items, id)
	s.autoSaveChan <- ItemC{Item: Item{ID: id}, OP: REMOVE}
}

func (s *ImprovedMemoryStorage) Get(id string) *Item {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.items[id]
}

func (s *ImprovedMemoryStorage) List() []*Item {
	s.lock.RLock()
	defer s.lock.RUnlock()

	items := make([]*Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}

	// 按创建时间倒序排序（最新的在前面）
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreateTime.After(items[j].CreateTime)
	})

	return items
}

func (s *ImprovedMemoryStorage) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.items)
}

// 实现 IStorage 接口
func (s *ImprovedMemoryStorage) Load() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// 重置文件指针到开头
	_, err := s.storageFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	decoder := json.NewDecoder(s.storageFile)

	for {
		var data map[string]interface{}
		if err := decoder.Decode(&data); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Failed to decode line: %v", err)
			continue
		}

		opFloat, ok := data["op"].(float64)
		if !ok {
			continue
		}
		op := OP(opFloat)

		itemData, ok := data["item"].(map[string]interface{})
		if !ok {
			continue
		}

		// 重新编码为 JSON 然后解码为 Item
		itemBytes, err := json.Marshal(itemData)
		if err != nil {
			continue
		}

		var item Item
		if err := json.Unmarshal(itemBytes, &item); err != nil {
			continue
		}

		switch op {
		case ADD:
			s.items[item.ID] = &item
		case REMOVE:
			delete(s.items, item.ID)
		}
	}

	return nil
}

func (s *ImprovedMemoryStorage) Save() error {
	// 这个方法在新的实现中不需要，因为我们有自动保存
	return nil
}

// CleanupExpired 清理过期数据
func (s *ImprovedMemoryStorage) CleanupExpired(maxAge time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()

	cutoff := time.Now().Add(-maxAge)
	cleanedCount := 0

	for id, item := range s.items {
		if item.CreateTime.Before(cutoff) {
			// 如果是文件，需要删除物理文件
			if item.Type != TEXT {
				if saveName, ok := item.Meta["saveName"].(string); ok {
					if path, ok := item.Meta["path"].(string); ok {
						filePath := filepath.Join(path, saveName)
						if err := os.Remove(filePath); err != nil {
							log.Printf("Failed to delete expired file %s: %v", filePath, err)
						} else {
							log.Printf("Deleted expired file: %s", filePath)
						}
					}
				}
			}

			delete(s.items, id)
			s.autoSaveChan <- ItemC{Item: Item{ID: id}, OP: REMOVE}
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		log.Printf("Cleaned up %d expired items (older than %v)", cleanedCount, maxAge)
	}
}
