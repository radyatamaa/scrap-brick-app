package repository

import (
	"context"
	"strings"

	"github.com/radyatamaa/scrap-brick-app/internal/domain"
	"github.com/radyatamaa/scrap-brick-app/pkg/database/paginator"
	"github.com/radyatamaa/scrap-brick-app/pkg/zaplogger"
	"gorm.io/gorm"
)

type pgProductRepository struct {
	zapLogger zaplogger.Logger
	db        *gorm.DB
}

func NewPgProductRepository(db *gorm.DB, zapLogger zaplogger.Logger) domain.PgProductRepository {
	return &pgProductRepository{
		db:        db,
		zapLogger: zapLogger,
	}
}

func (c pgProductRepository) DB() *gorm.DB {
	return c.db
}

func (c pgProductRepository) FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error) {
	p := paginator.NewPaginator(c.db, offset, limit, model)
	if err := p.FindWithFilter(ctx, order, fields, associate, filter, args).Select(strings.Join(fields, ",")).Error; err != nil {
		return nil, err
	}
	return model, nil
}

func (c pgProductRepository) SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error {

	db := c.db.WithContext(ctx)

	if len(fields) > 0 {
		db = db.Select(strings.Join(fields, ","))
	}
	if len(associate) > 0 {
		for _, v := range associate {
			db.Joins(v)
		}
	}

	if len(filter) > 0 && len(args) == len(filter) {
		for i := range filter {
			db = db.Where(filter[i], args[i])
		}
	}

	if err := db.First(model).Error; err != nil {
		return err
	}
	return nil
}

func (c pgProductRepository) Update(ctx context.Context, data domain.Product) error {

	err := c.db.WithContext(ctx).Updates(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (c pgProductRepository) UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error {

	return c.db.WithContext(ctx).Table(domain.Product{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c pgProductRepository) Store(ctx context.Context, data domain.Product) (domain.Product, error) {

	err := c.db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func (c pgProductRepository) Delete(ctx context.Context, id int) (int, error) {

	err := c.db.WithContext(ctx).Exec("delete from "+domain.Product{}.TableName()+" where id =?", id).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c pgProductRepository) SoftDelete(ctx context.Context, id int) (int, error) {
	var data domain.Product

	err := c.db.WithContext(ctx).Where("id = ?", id).Delete(&data).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c pgProductRepository) DeleteAll(ctx context.Context) error {
	var data domain.Product

	err := c.db.WithContext(ctx).Where("created_at is not null").Delete(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (c pgProductRepository) UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error {

	return tx.WithContext(ctx).Table(domain.Product{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c pgProductRepository) StoreWithTx(ctx context.Context, tx *gorm.DB, data domain.Product) (int, error) {

	err := tx.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data.ID, err
	}
	return data.ID, nil
}
