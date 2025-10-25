package biz

// ReplyParam 商家回复评价的参数
type ReplyParam struct {
	ReviewId  int64
	StoreId   int64
	Content   string
	PicInfo   string
	VideoInfo string
}

type AppealParam struct {
	AppealId  int64
	ReviewId  int64
	StoreId   int64
	Content   string
	Status    int32
	PicInfo   string
	VideoInfo string
	OpUser    string
	Reason    string
}
