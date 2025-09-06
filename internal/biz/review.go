package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"review-service/internal/data/model"
)

type ReviewRepo interface {
	SaveReview(context.Context, *model.ReviewInfo) (*model.ReviewInfo, error)
}

type ReviewUsecase struct {
	repo ReviewRepo
	log  *log.Helper
}

func NewReviewUsecase(repo ReviewRepo, logger log.Logger) *ReviewUsecase {
	return &ReviewUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateReview 创建评价
// 实现业务逻辑的地方
// service层调用该方法
func (uc *ReviewUsecase) CreateReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	uc.log.WithContext(ctx).Debugf("create review, data:%+v", review)
	// 1. 数据校验

	// 2. 生成review Id

	// 3. 查询订单和商品快照信息

	// 4. 拼装数据入库

	return uc.repo.SaveReview(ctx, review)
}
