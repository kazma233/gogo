package repository

import (
	"database/sql"
	"fmt"
	"time"

	"picstash/internal/model"
)

type TagRepositoryInterface interface {
	Create(name, color string) (*model.Tag, error)
	Update(id int64, name, color string) (*model.Tag, error)
	Delete(id int64) error
	GetByID(id int64) (*model.Tag, error)
	GetAll() ([]*model.Tag, error)
	GetByImageID(imageID int64) ([]*model.Tag, error)
}

type TagRepository struct {
	tx *sql.Tx
}

func NewTagRepository(tx *sql.Tx) *TagRepository {
	return &TagRepository{tx: tx}
}

func (r *TagRepository) Create(name, color string) (*model.Tag, error) {
	result, err := r.tx.Exec(`INSERT INTO tags (name, color) VALUES (?, ?)`, name, color)
	if err != nil {
		return nil, fmt.Errorf("创建标签失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取标签ID失败: %w", err)
	}

	tag := &model.Tag{
		ID:        id,
		Name:      name,
		Color:     color,
		CreatedAt: time.Now(),
	}
	return tag, nil
}

func (r *TagRepository) Update(id int64, name, color string) (*model.Tag, error) {
	_, err := r.tx.Exec(`UPDATE tags SET name = ?, color = ? WHERE id = ?`, name, color, id)
	if err != nil {
		return nil, fmt.Errorf("更新标签失败: %w", err)
	}

	tag, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (r *TagRepository) Delete(id int64) error {
	result, err := r.tx.Exec(`DELETE FROM tags WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("删除标签失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("标签不存在")
	}
	return nil
}

func (r *TagRepository) GetByID(id int64) (*model.Tag, error) {
	tag := &model.Tag{}

	err := r.tx.QueryRow(`SELECT id, name, color, created_at FROM tags WHERE id = ?`, id).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("标签不存在")
		}
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}
	return tag, nil
}

func (r *TagRepository) GetAll() ([]*model.Tag, error) {
	rows, err := r.tx.Query(`SELECT id, name, color, created_at FROM tags ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("查询标签列表失败: %w", err)
	}
	defer rows.Close()

	var tags []*model.Tag
	for rows.Next() {
		tag := &model.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描标签数据失败: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *TagRepository) GetByImageID(imageID int64) ([]*model.Tag, error) {
	rows, err := r.tx.Query(`
		SELECT t.id, t.name, t.color, t.created_at
		FROM tags t
		INNER JOIN image_tags it ON t.id = it.tag_id
		WHERE it.image_id = ?
		ORDER BY it.created_at DESC
	`, imageID)
	if err != nil {
		return nil, fmt.Errorf("查询图片标签失败: %w", err)
	}
	defer rows.Close()

	var tags []*model.Tag
	for rows.Next() {
		tag := &model.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描标签数据失败: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}
