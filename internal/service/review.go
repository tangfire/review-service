package service

import (
	"context"
	"fmt"
	"review-service/internal/biz"
	"review-service/internal/data/model"

	pb "review-service/api/review/v1"
)

type ReviewService struct {
	pb.UnimplementedReviewServer

	uc *biz.ReviewUsecase
}

func NewReviewService(uc *biz.ReviewUsecase) *ReviewService {
	return &ReviewService{uc: uc}
}

func (s *ReviewService) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewReply, error) {
	fmt.Printf("[service] CreateReview, req:%+v\n", req)
	// 参数转换

	// 调用biz层
	var anonymous int32 = 0
	if req.Anonymous {
		anonymous = 1
	}
	review, err := s.uc.CreateReview(ctx, &model.ReviewInfo{
		UserID:       req.GetUserId(),
		OrderID:      req.GetOrderId(),
		Score:        req.GetScore(),
		ServiceScore: req.GetServiceScore(),
		ExpressScore: req.GetExpressScore(),
		Content:      req.GetContent(),
		PicInfo:      req.GetPicInfo(),
		VideoInfo:    req.GetVideoInfo(),
		Anonymous:    anonymous,
		Status:       0,
	})
	if err != nil {
		return nil, err
	}

	// 拼装返回结果
	return &pb.CreateReviewReply{ReviewId: review.ReviewID}, nil
}

func (s *ReviewService) ReplyReview(ctx context.Context, req *pb.ReplyReviewRequest) (*pb.ReplyReviewReply, error) {
	fmt.Printf("[service] ReplyReview, req:%+v\n", req)

	// 调用biz层
	reply, err := s.uc.CreateReply(ctx, &biz.ReplyParam{
		ReviewId:  req.GetReviewId(),
		StoreId:   req.GetStoreId(),
		Content:   req.GetContent(),
		PicInfo:   req.GetPicInfo(),
		VideoInfo: req.GetVideoInfo(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ReplyReviewReply{ReplyId: reply.ReplyID}, nil

}

func (s *ReviewService) AppealReview(ctx context.Context, req *pb.AppealReviewRequest) (*pb.AppealReviewReply, error) {
	fmt.Printf("[service] AppealReview, req:%+v\n", req)
	ret, err := s.uc.CreateAppeal(ctx, &biz.AppealParam{
		ReviewId:  req.GetReviewId(),
		StoreId:   req.GetStoreId(),
		Content:   req.GetContent(),
		PicInfo:   req.GetPicInfo(),
		VideoInfo: req.GetVideoInfo(),
		OpUser:    req.GetOpUser(),
		Reason:    req.GetReason(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.AppealReviewReply{AppealId: ret.AppealID}, nil
}

func (s *ReviewService) AuditAppeal(ctx context.Context, req *pb.AuditAppealRequest) (*pb.AuditAppealReply, error) {
	fmt.Printf("[service] AuditAppeal, req:%+v\n", req)
	var resp *pb.AuditAppealReply
	err := s.uc.UpdateAppeal(ctx, &biz.AppealParam{
		AppealId: req.GetAppealId(),
		ReviewId: req.GetReviewId(),
		Status:   req.GetStatus(),
		OpUser:   req.GetOpUser(),
		Reason:   req.GetOpReason(),
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ReviewService) TestConn(context.Context, *pb.TestConnRequest) (*pb.TestConnReply, error) {
	return &pb.TestConnReply{
		Pong: "pong!",
	}, nil
}

func (s *ReviewService) ListReviewByStoreId(ctx context.Context, req *pb.ListReviewByStoreIdRequest) (*pb.ListReviewByStoreIdReply, error) {
	fmt.Printf("[service] ListReviewByStoreId, req:%+v\n", req)
	reviewList, err := s.uc.ListReviewByStoreId(ctx, req.StoreId, int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}
	retList := make([]*pb.ReviewInfo, 0, len(reviewList))
	for _, v := range reviewList {
		retList = append(retList, &pb.ReviewInfo{
			ReviewId:     v.ReviewID,
			UserId:       v.UserID,
			OrderId:      v.OrderID,
			Score:        v.Score,
			ServiceScore: v.ServiceScore,
			ExpressScore: v.ExpressScore,
			Content:      v.Content,
			PicInfo:      v.PicInfo,
			VideoInfo:    v.VideoInfo,
		})

	}
	return &pb.ListReviewByStoreIdReply{List: retList}, nil
}

func (s *ReviewService) UpdateReview(ctx context.Context, req *pb.UpdateReviewRequest) (*pb.UpdateReviewReply, error) {
	return &pb.UpdateReviewReply{}, nil
}
func (s *ReviewService) DeleteReview(ctx context.Context, req *pb.DeleteReviewRequest) (*pb.DeleteReviewReply, error) {
	return &pb.DeleteReviewReply{}, nil
}
func (s *ReviewService) GetReview(ctx context.Context, req *pb.GetReviewRequest) (*pb.GetReviewReply, error) {
	return &pb.GetReviewReply{}, nil
}
func (s *ReviewService) ListReview(ctx context.Context, req *pb.ListReviewRequest) (*pb.ListReviewReply, error) {
	return &pb.ListReviewReply{}, nil
}
