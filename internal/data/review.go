package data

import (
	"context"
	"review-service/internal/data/model"

	"review-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type reviewRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return &reviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *reviewRepo) SaveReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	err := r.data.query.ReviewInfo.
		WithContext(ctx).
		Save(review)
	return review, err
}

// GetReviewByOrderId 根据订单Id查询评价
func (r *reviewRepo) GetReviewByOrderId(ctx context.Context, orderId int64) ([]*model.ReviewInfo, error) {
	return r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.OrderID.Eq(orderId)).
		Find()
}

func (r *reviewRepo) SaveReply(ctx context.Context, info *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	// 1. 数据校验
	// 1.1 数据合法性校验(已回复的评价不允许商家再次回复)
	// 1.2 水平越权校验(A商家只能回复自己的不能回复B商家的)
	// 2. 更新数据库中的数据(评价回复表和评价表要同时更新，涉及到事务操作)
	// 3. 返回
	return nil, nil

}
