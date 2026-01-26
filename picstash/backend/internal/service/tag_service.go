package service

import (
	"database/sql"
	"fmt"
	"log/slog"

	"picstash/internal/model"
	"picstash/internal/repository"
)

type TagService struct {
	db *sql.DB
}

func NewTagService(db *sql.DB) *TagService {
	return &TagService{db: db}
}

func (s *TagService) Create(name, color string) (*model.Tag, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tagRepo := repository.NewTagRepository(tx)

	tag, err := tagRepo.Create(name, color)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	slog.Info("标签创建成功", "id", tag.ID, "name", name)

	return tag, nil
}

func (s *TagService) Update(id int64, name, color string) (*model.Tag, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tagRepo := repository.NewTagRepository(tx)

	tag, err := tagRepo.Update(id, name, color)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	slog.Info("标签更新成功", "id", id, "name", name)

	return tag, nil
}

func (s *TagService) Delete(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tagRepo := repository.NewTagRepository(tx)

	err = tagRepo.Delete(id)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	slog.Info("标签删除成功", "id", id)

	return nil
}

func (s *TagService) GetByID(id int64) (*model.Tag, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tagRepo := repository.NewTagRepository(tx)

	tag, err := tagRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *TagService) GetAll() ([]*model.Tag, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tagRepo := repository.NewTagRepository(tx)

	tags, err := tagRepo.GetAll()
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (s *TagService) GetByImageID(imageID int64) ([]*model.Tag, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tagRepo := repository.NewTagRepository(tx)

	tags, err := tagRepo.GetByImageID(imageID)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
