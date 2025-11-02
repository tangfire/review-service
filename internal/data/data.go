package data

import (
	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"review-service/internal/conf"
	"review-service/internal/data/query"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewReviewRepo, NewDB, NewESClient)

// Data .
type Data struct {
	// TODO wrapped database client
	//db *gorm.DB
	query *query.Query
	log   *log.Helper
	es    *elasticsearch.TypedClient
}

// NewData .
func NewData(db *gorm.DB, esClient *elasticsearch.TypedClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	// 非常重要!为GEN生成的query代码设置数据库连接对象
	query.SetDefault(db)
	return &Data{query: query.Q, es: esClient, log: log.NewHelper(logger)}, cleanup, nil
}

// NewESClient ES Client 的构造函数
func NewESClient(cfg *conf.Elasticsearch) (*elasticsearch.TypedClient, error) {
	c := elasticsearch.Config{
		Addresses: cfg.Addresses,
	}
	return elasticsearch.NewTypedClient(c)
}

func NewDB(c *conf.Data) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(c.Database.Source))
	if err != nil {
		return nil, err
	}
	return db, nil
}
