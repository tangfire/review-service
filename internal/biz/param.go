package biz

// ReplyParam 商家回复评价的参数
type ReplyParam struct {
	ReviewId  int64
	StoreId   int64
	Content   string
	PicInfo   string
	VideoInfo string
}
