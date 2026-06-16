package api

// ParseRequest 解析请求
type ParseRequest struct {
	URL string `json:"url" binding:"required"`
}

// VideoInfo 返回给前端的视频信息
type VideoInfo struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	AuthorAvatar  string `json:"author_avatar"`
	CoverURL      string `json:"cover_url"`
	Duration      int    `json:"duration"`
	VideoURL      string `json:"video_url"`
	DecryptKey    string `json:"decrypt_key"`
	MediaType     int    `json:"media_type"`
	LikeCount     string `json:"like_count"`
	FavCount      string `json:"fav_count"`
	CommentCount  string `json:"comment_count"`
	ForwardCount  string `json:"forward_count"`
	CreateTime    int    `json:"create_time"`
	H264URL       string `json:"h264_url"`
	H265URL       string `json:"h265_url"`
	OriginalURL   string `json:"original_url"`
}

// UpdateCookieRequest 更新 cookie 请求
type UpdateCookieRequest struct {
	Cookie string `json:"cookie" binding:"required"`
}

// Response 通用响应
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}
