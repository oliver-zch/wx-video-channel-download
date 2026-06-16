package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"wx_channels_web/internal/service"

	"github.com/gin-gonic/gin"
)

const encLimit = 131072 // 128KB 加密长度

// handleStatus 健康检查
func (s *Server) handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "ok",
		Data: gin.H{"version": "1.0.0"},
	})
}

// handleParse 解析 SPH 分享链接
func (s *Server) handleParse(c *gin.Context) {
	var req ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{Code: 400, Msg: "参数错误: url 必填"})
		return
	}

	cookie := s.config.GetSphCookie()
	if cookie == "" {
		c.JSON(http.StatusOK, Response{Code: 400, Msg: "sph_cookie 未配置，请先通过 /api/config/cookie 设置"})
		return
	}

	feedResp, err := service.FetchVideoProfileWithShareUrl(req.URL, cookie)
	if err != nil {
		log.Printf("[handleParse] error: %v", err)
		c.JSON(http.StatusOK, Response{Code: 500, Msg: fmt.Sprintf("解析失败: %v", err)})
		return
	}

	if feedResp == nil {
		c.JSON(http.StatusOK, Response{Code: 500, Msg: "解析结果为空"})
		return
	}

	if feedResp.Errcode != 0 {
		c.JSON(http.StatusOK, Response{Code: feedResp.Errcode, Msg: feedResp.Errmsg})
		return
	}

	feed := feedResp.Data.Feedinfo
	author := feedResp.Data.Authorinfo

	// 清理视频 URL
	originalURL := service.CleanVideoURL(feed.Videourl)
	videoURL := originalURL
	if videoURL == "" {
		videoURL = feed.Videourl
	}

	// 提取解密 key
	decryptKey := fmt.Sprintf("%d", feed.Decodekey)

	info := VideoInfo{
		Title:        feed.Description,
		Author:       author.Nickname,
		AuthorAvatar: author.Headimgurl,
		CoverURL:     feed.Coverurl,
		Duration:     0,
		VideoURL:     videoURL,
		DecryptKey:   decryptKey,
		MediaType:    feed.Mediatype,
		LikeCount:    feed.Likecountfmt,
		FavCount:     feed.Favcountfmt,
		CommentCount: feed.Commentcountfmt,
		ForwardCount: feed.Forwardcountfmt,
		CreateTime:   feed.Createtime,
		H264URL:      service.CleanVideoURL(feed.H264videoinfo.Videourl),
		H265URL:      service.CleanVideoURL(feed.H265videoinfo.Videourl),
		OriginalURL:  originalURL,
	}

	c.JSON(http.StatusOK, Response{Code: 0, Msg: "成功", Data: info})
}

// handleProxy 视频代理播放/下载
func (s *Server) handleProxy(c *gin.Context) {
	videoURL := c.Query("url")
	keyStr := c.Query("key")
	filename := c.Query("filename")

	if videoURL == "" {
		http.Error(c.Writer, "url parameter is required", http.StatusBadRequest)
		return
	}

	var key uint64
	if keyStr != "" {
		if k, err := strconv.ParseUint(keyStr, 10, 64); err == nil {
			key = k
		}
	}

	decryptor := service.NewChannelsVideoDecryptor()

	if filename != "" {
		// 下载模式
		decryptor.DecryptOnly(c.Writer, c.Request, videoURL, key, encLimit, filename)
	} else {
		// 播放模式
		if key > 0 {
			decryptor.DecryptOnlyInline(c.Writer, c.Request, videoURL, key, encLimit)
		} else {
			// 无 key，直接代理
			decryptor.SimpleProxy(videoURL, c.Writer, c.Request)
		}
	}
}

// handleUpdateCookie 运行时更新 SPH cookie
func (s *Server) handleUpdateCookie(c *gin.Context) {
	var req UpdateCookieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{Code: 400, Msg: "参数错误: cookie 必填"})
		return
	}

	s.config.SetSphCookie(req.Cookie)
	log.Printf("[config] sph_cookie updated via API")
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "cookie 更新成功"})
}
