package repository

import (
	"context"
	"fmt"
	"gonews/internal/core/domain/entity"
	"gonews/internal/core/domain/model"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContentRepository interface {
	GetContents(ctx context.Context, query entity.QueryString) ([]entity.ContentEntity, error)
	GetContentByID(ctx context.Context, id int64) (*entity.ContentEntity, error)
	CreateContent(ctx context.Context, req entity.ContentEntity) error
	UpdateContent(ctx context.Context, req entity.ContentEntity) error
	DeleteContent(ctx context.Context, id int64) error
}

type contentRepository struct {
	db *gorm.DB
}

func (c *contentRepository) GetContents(ctx context.Context, query entity.QueryString) ([]entity.ContentEntity, error) {
	var modelContents []model.Content

	
	status := ""
	
	sqlMain := c.db.Preload(clause.Associations).
	Where("title ilike ? OR excerpt ilike ? OR description ilike ?", "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%").
	Where("status LIKE ?", "%"+status+"%")

	if query.OrderBy == "" {
	query.OrderBy = "created_at"
	}
	if query.OrderType == "" {
		query.OrderType = "DESC"
	}
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Page <= 0 {
		query.Page = 1
	}

	if query.Status != "" {
		sqlMain = sqlMain.Where("status = ?", query.Status)
	}
	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	err := sqlMain.
			Order(order).
			Limit(query.Limit).
			Offset(offset).
			Find(&modelContents).Error
	if err != nil {
		code = "[REPOSITORY] GetContents - 1"
		log.Errorw(code, err)
		return nil, err
	}

	resps := []entity.ContentEntity{}
	for _,val := range modelContents {
		tags := strings.Split(val.Tags, ",")
		resp := entity.ContentEntity{
			ID: val.ID,
			Title: val.Title,
			Excerpt: val.Excerpt,
			Description: val.Description,
			Image: val.Image,
			Tags: tags,
			Status: val.Status,
			CategoryID: val.CategoryID,
			CreatedByID: val.CreatedByID,
			CreatedAt: val.CreatedAt,
			Category: entity.CategoryEntity{
				ID: val.Category.ID,
				Title: val.Category.Title,
				Slug: val.Category.Slug,
			},
			User: entity.UserEntity{
				ID: val.User.ID,
				Name: val.User.Name,
			},
		}

		resps = append(resps, resp)
	}
	
	return resps, nil
}

func (c *contentRepository) GetContentByID(ctx context.Context, id int64) (*entity.ContentEntity, error) {
	var modelContents model.Content

	err = c.db.Where("id = ?", id).Preload(clause.Associations).First(&modelContents).Error
	
	if err != nil {
		code = "[REPOSITORY] GetContentByID - 1"
		log.Errorw(code, err)
		return nil, err
	}

	tags := strings.Split(modelContents.Tags, ",")
		resp := entity.ContentEntity{
			ID: modelContents.ID,
			Title: modelContents.Title,
			Excerpt: modelContents.Excerpt,
			Description: modelContents.Description,
			Image: modelContents.Image,
			Tags: tags,
			Status: modelContents.Status,
			CategoryID: modelContents.CategoryID,
			CreatedByID: modelContents.CreatedByID,
			CreatedAt: modelContents.CreatedAt,
			Category: entity.CategoryEntity{
				ID: modelContents.Category.ID,
				Title: modelContents.Category.Title,
				Slug: modelContents.Category.Slug,
			},
			User: entity.UserEntity{
				ID: modelContents.User.ID,
				Name: modelContents.User.Name,
			},
		}

		return &resp,nil
}

func (c *contentRepository) CreateContent(ctx context.Context, req entity.ContentEntity) error {
	tags := strings.Join(req.Tags, ",")
	modelContent := model.Content{
		Title: req.Title,
		Excerpt: req.Excerpt,
		Description: req.Description,
		Image: req.Image,
		Tags: tags,
		Status: req.Status,
		CategoryID: req.CategoryID,
		CreatedByID: req.CreatedByID,
	}

	err = c.db.Create(&modelContent).Error

	if err != nil {
		code = "[REPOSITORY] CreateContent - 1"
		log.Errorw(code, err)
		return err
	}
	
	return nil
}

func (c *contentRepository) UpdateContent(ctx context.Context, req entity.ContentEntity) error {
	tags := strings.Join(req.Tags, ",")
	modelContent := model.Content{
		Title: req.Title,
		Excerpt: req.Excerpt,
		Description: req.Description,
		Image: req.Image,
		Tags: tags,
		Status: req.Status,
		CategoryID: req.CategoryID,
		CreatedByID: req.CreatedByID,
	}

	err = c.db.Where("id = ?", req.ID).Updates(&modelContent).Error
	if err != nil {
		code = "[REPOSITORY] UpdateContent - 1"
		log.Errorw(code, err)
		return err
	}
	
	return nil
}

func (c *contentRepository) DeleteContent(ctx context.Context, id int64) error {
	err = c.db.Where("id = ?", id).Delete(&model.Content{}).Error
	if err != nil {
		code = "[REPOSITORY] DeleteContent - 1"
		log.Errorw(code, err)
		return err
	}

	return nil
}

func NewContentRepository(db *gorm.DB) ContentRepository {
	return &contentRepository{db: db}
}