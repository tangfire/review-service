package data

import (
	"context"
	"errors"
	"review-service/internal/data/model"
	"review-service/internal/data/query"

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

func (r *reviewRepo) SaveReply(ctx context.Context, reply *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	// 1. 数据校验
	// 1.1 数据合法性校验(已回复的评价不允许商家再次回复)
	// 先用评价ID查库，看下是否已回复
	review, err := r.data.query.ReviewInfo.WithContext(ctx).Where(r.data.query.ReviewInfo.ReviewID.Eq(reply.ReplyID)).First()
	if err != nil {
		return nil, err
	}
	if review.HasReply == 1 {
		return nil, errors.New("该评价已回复")
	}

	// 1.2 水平越权校验(A商家只能回复自己的不能回复B商家的)
	// 举例子: 用户A删除订单，userId + orderId 当条件去查询订单然后删除
	if review.StoreID != reply.StoreID {
		return nil, errors.New("水平越权")
	}

	// 2. 更新数据库中的数据(评价回复表和评价表要同时更新，涉及到事务操作)
	// 事务操作
	err = r.data.query.Transaction(func(tx *query.Query) error {
		// 回复表插入一条数据
		if err := tx.ReviewReplyInfo.
			WithContext(ctx).
			Save(reply); err != nil {
			r.log.WithContext(ctx).Errorf("SaveReply create reply fail,err:%v", err)
			return err
		}
		// 评价表更新hasReply字段
		if _, err := tx.ReviewInfo.WithContext(ctx).Where(tx.ReviewInfo.ReviewID.Eq(reply.ReplyID)).Update(tx.ReviewInfo.HasReply, 1); err != nil {
			r.log.WithContext(ctx).Errorf("SaveReply update reply fail,err:%v", err)
			return err
		}
		return nil
	})

	// 3. 返回
	return reply, err

}
