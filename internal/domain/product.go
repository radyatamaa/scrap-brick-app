package domain

import (
	"context"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"gorm.io/gorm"
	"time"
)

type Product struct {
	ID            int       `gorm:"column:id;primarykey;autoIncrement:true"`
	Name     string    `gorm:"type:varchar(255);column:name"`
	Desc      string    `gorm:"type:varchar(255);column:desc"`
	Image        string    `gorm:"type:text;column:image"`
	Price   string    `gorm:"type:varchar(100);column:price"`
	Rating string    `gorm:"type:varchar(100);column:rating"`
	Merchant         string    `gorm:"type:varchar(255);column:merchant"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

// TableName name of table
func (r Product) TableName() string {
	return "products"
}


// PgProductRepository Repository Interface
type PgProductRepository interface {
	SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error
	FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error)
	Update(ctx context.Context, data Product) error
	UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error
	UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error
	Store(ctx context.Context, data Product) (Product, error)
	StoreWithTx(ctx context.Context, tx *gorm.DB, data Product) (int, error)
	Delete(ctx context.Context, id int) (int, error)
	SoftDelete(ctx context.Context, id int) (int, error)
	DB() *gorm.DB
}

// ProductUseCase UseCase Interface
type ProductUseCase interface {
	ScrapeProducts(beegoCtx *beegoContext.Context,maxCount int) error
}