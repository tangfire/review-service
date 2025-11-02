package biz

import (
	"context"
	"errors"
	"fmt"
	"review-service/pkg/snowflake"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"review-service/internal/data/model"
)

type ReviewRepo interface {
	SaveReview(context.Context, *model.ReviewInfo) (*model.ReviewInfo, error)
	GetReviewByOrderId(context.Context, int64) ([]*model.ReviewInfo, error)
	SaveReply(ctx context.Context, info *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error)
	SaveAppeal(ctx context.Context, info *model.ReviewAppealInfo) (*model.ReviewAppealInfo, error)
	UpdateAppeal(ctx context.Context, info *model.ReviewAppealInfo) error
	ListReviewByStoreId(ctx context.Context, storeId int64, offset, limit int) ([]*MyReviewInfo, error)
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
	// 1.1 参数基础校验: 正常来说不应该放在这一层，你在上一层或者框架层都应该能拦住(validate参数校验)

	// 1.2 参数业务校验: 带业务逻辑的参数校验，比如已经评价过的订单不能再创建评价
	reviews, err := uc.repo.GetReviewByOrderId(ctx, review.OrderID)
	if err != nil {
		return nil, errors.New("查询数据库失败")
	}
	if len(reviews) > 0 {
		// 已经评价过
		return nil, errors.New(fmt.Sprintf("订单:%d已评价", review.OrderID))
	}
	// 2. 生成review Id
	// 这里可以使用雪花算法自己生成
	// 也可以直接接入公司内部的分布式ID生成服务(前提是公司内部有这种服务)
	review.ReviewID = snowflake.GenerateID()

	// 3. 查询订单和商品快照信息
	// 实际业务场景下就需要查询订单服务和商家服务(比如说通过RPC调用订单服务和商家服务)

	// 4. 拼装数据入库

	return uc.repo.SaveReview(ctx, review)
}

func (uc *ReviewUsecase) CreateReply(ctx context.Context, param *ReplyParam) (*model.ReviewReplyInfo, error) {
	// 调用data层创建一个评价的回复
	uc.log.WithContext(ctx).Debugf("[biz] CreateReply, param:%+v", param)
	reply := &model.ReviewReplyInfo{
		ReplyID:   snowflake.GenerateID(),
		ReviewID:  param.ReviewId,
		StoreID:   param.StoreId,
		Content:   param.Content,
		PicInfo:   param.PicInfo,
		VideoInfo: param.VideoInfo,
	}
	return uc.repo.SaveReply(ctx, reply)
}

func (uc *ReviewUsecase) CreateAppeal(ctx context.Context, param *AppealParam) (*model.ReviewAppealInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] CreateAppeal, param:%+v", param)
	appeal := &model.ReviewAppealInfo{
		ReviewID:  param.ReviewId,
		StoreID:   param.StoreId,
		Content:   param.Content,
		PicInfo:   param.PicInfo,
		VideoInfo: param.VideoInfo,
		OpUser:    param.OpUser,
		Reason:    param.Reason,
		Status:    PendingReview,
	}

	return uc.repo.SaveAppeal(ctx, appeal)

}

func (uc *ReviewUsecase) UpdateAppeal(ctx context.Context, param *AppealParam) error {
	uc.log.WithContext(ctx).Debugf("[biz] UpdateAppeal, param:%+v", param)
	appeal := &model.ReviewAppealInfo{
		AppealID: param.AppealId,
		ReviewID: param.ReviewId,
		OpUser:   param.OpUser,
		Reason:   param.Reason,
		Status:   param.Status,
	}
	return uc.repo.UpdateAppeal(ctx, appeal)
}

func (uc *ReviewUsecase) ListReviewByStoreId(ctx context.Context, storeId int64, page, size int) ([]*MyReviewInfo, error) {
	uc.log.WithContext(ctx).Debugf("[biz] ListReviewByStoreId")
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 50 {
		size = 10
	}
	offset := (page - 1) * size
	limit := size
	uc.log.WithContext(ctx).Debugf("[biz] ListReviewByStoreId:%v", storeId)
	return uc.repo.ListReviewByStoreId(ctx, storeId, offset, limit)

}

type MyReviewInfo struct {
	*model.ReviewInfo
	CreateAt     MyTime `json:"create_at"`
	UpdateAt     MyTime `json:"update_at"`
	Anonymous    int32  `json:"anonymous,string"`
	Score        int32  `json:"score,string"`
	ServiceScore int32  `json:"service_score,string"`
	ExpressScore int32  `json:"express_score,string"`
	HasMedia     int32  `json:"has_media,string"`
	Status       int32  `json:"status,string"`
	IsDefault    int32  `json:"is_default,string"`
	HasReply     int32  `json:"has_reply,string"`
	ID           int64  `json:"id,string"`
	Version      int32  `json:"version,string"`
	ReviewID     int64  `json:"review_id,string"`
	OrderID      int64  `json:"order_id,string"`
	SkuID        int64  `json:"sku_id,string"`
	SpuID        int64  `json:"spu_id,string"`
	StoreID      int64  `json:"store_id,string"`
	UserID       int64  `json:"user_id,string"`
}

type MyTime time.Time

// UnmarshalJSON 自定义时间解析
func (t *MyTime) UnmarshalJSON(data []byte) error {
	// 去除JSON字符串的引号
	timeStr := strings.Trim(string(data), `"`)

	// 如果时间是空值，设置为零值
	if timeStr == "" || timeStr == "null" {
		*t = MyTime(time.Time{})
		return nil
	}

	// 尝试多种可能的时间格式
	var parsedTime time.Time
	var err error

	// 格式1: "2025-11-02 05:41:02" (你遇到的格式)
	parsedTime, err = time.Parse(time.DateTime, timeStr)
	if err != nil {
		return fmt.Errorf("无法解析时间字符串: %s, 错误: %v", timeStr, err)
	}

	*t = MyTime(parsedTime)
	return nil

}
