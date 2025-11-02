package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"gorm.io/gorm"
	"review-service/internal/data/model"
	"review-service/internal/data/query"
	"review-service/pkg/snowflake"

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
	review, err := r.data.query.ReviewInfo.WithContext(ctx).Where(r.data.query.ReviewInfo.ReviewID.Eq(reply.ReviewID)).First()
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
		if _, err := tx.ReviewInfo.WithContext(ctx).Where(tx.ReviewInfo.ReviewID.Eq(reply.ReviewID)).Update(tx.ReviewInfo.HasReply, 1); err != nil {
			r.log.WithContext(ctx).Errorf("SaveReply update reply fail,err:%v", err)
			return err
		}
		return nil
	})

	// 3. 返回
	return reply, err

}

func (r *reviewRepo) SaveAppeal(ctx context.Context, info *model.ReviewAppealInfo) (*model.ReviewAppealInfo, error) {
	var err error
	_, err = r.data.query.ReviewInfo.WithContext(ctx).Where(
		query.ReviewInfo.ReviewID.Eq(info.ReviewID),
		query.ReviewInfo.StoreID.Eq(info.StoreID)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("评价不存在或不属于该商店")
	}
	if err != nil {
		r.log.WithContext(ctx).Errorf("SaveAppeal|查询评价失败, reviewID:%d, storeID:%d, err:%v",
			info.ReviewID, info.StoreID, err)
		return nil, err
	}
	// 先查询有没有申述
	ret, err := r.data.query.ReviewAppealInfo.WithContext(ctx).Where(
		query.ReviewAppealInfo.ReviewID.Eq(info.ReviewID),
		query.ReviewAppealInfo.StoreID.Eq(info.StoreID),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		r.log.WithContext(ctx).Errorf("SaveAppeal|First fail,data:%v,err:%v", info, err)
		return nil, err
	}
	// 查询不到审核过的申述记录
	if ret != nil {
		if ret.Status > 10 {
			return nil, errors.New("该评价已有审核过的申述记录")
		}
		// 1. 有申述记录但是处于待审核状态,需要更新
		_, err = r.data.query.ReviewAppealInfo.WithContext(ctx).
			Where(r.data.query.ReviewAppealInfo.ReviewID.Eq(info.ReviewID)).
			UpdateColumns(map[string]interface{}{
				"status":     info.Status,
				"content":    info.Content,
				"reason":     info.Reason,
				"pic_info":   info.PicInfo,
				"video_info": info.VideoInfo,
			})
		if err != nil {
			r.log.WithContext(ctx).Errorf("SaveAppeal|UpdateColumns fail,err:%v", err)
			return nil, err
		}
		return ret, nil

	}
	// 2. 没有申述记录,需要创建
	info.AppealID = snowflake.GenerateID()
	err = r.data.query.ReviewAppealInfo.WithContext(ctx).Save(info)
	if err != nil {
		r.log.WithContext(ctx).Errorf("SaveAppeal|Save fail,err:%v", err)
		return nil, err
	}
	return info, nil

}

func (r *reviewRepo) UpdateAppeal(ctx context.Context, info *model.ReviewAppealInfo) error {
	var err error
	err = r.data.query.Transaction(func(tx *query.Query) error {
		_, err = tx.ReviewAppealInfo.WithContext(ctx).Where(r.data.query.ReviewAppealInfo.AppealID.Eq(info.AppealID)).
			UpdateColumns(map[string]interface{}{
				"status":  info.Status,
				"op_user": info.OpUser,
				"reason":  info.Reason,
			})
		if err != nil {
			r.log.WithContext(ctx).Errorf("SaveAppeal|UpdateColumns fail,err:%v", err)
			return err
		}
		if info.Status == biz.Approved {
			_, err = tx.ReviewInfo.WithContext(ctx).Where(tx.ReviewInfo.ReviewID.Eq(info.ReviewID)).UpdateColumns(map[string]interface{}{
				"status": biz.Hidden,
			})
			if err != nil {
				r.log.WithContext(ctx).Errorf("SaveAppeal|UpdateColumns fail,err:%v", err)
				return err
			}
		}
		return nil
	})
	return err
}

// ListReviewByStoreId 根据storeId 分页查询评价
func (r *reviewRepo) ListReviewByStoreId(ctx context.Context, storeId int64, offset, limit int) ([]*biz.MyReviewInfo, error) {
	// 去ES里面查询评价
	resp, err := r.data.es.Search().Index("review").From(offset).Size(limit).
		Query(&types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"store_id": {Value: storeId},
						},
					},
				},
			},
		}).Do(ctx)
	if err != nil {
		r.log.WithContext(ctx).Errorf("ListReviewByStoreId fail,err:%v", err)
		return nil, err
	}
	fmt.Printf("es result total:%v\n", resp.Hits.Total.Value)
	//b, _ := json.Marshal(resp.Hits.Hits)
	//fmt.Printf("es result:%v\n", b)
	// 反序列化数据
	list := make([]*biz.MyReviewInfo, 0, resp.Hits.Total.Value)
	for _, hit := range resp.Hits.Hits {
		tmp := &biz.MyReviewInfo{}
		if err := json.Unmarshal(hit.Source_, tmp); err != nil {
			r.log.Errorf("ListReviewByStoreId fail,err:%v", err)
			continue
		}
		list = append(list, tmp)
	}
	return list, nil
}
