package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"somepro/internal/common"
	"somepro/internal/service"
)

type ItemHandler struct {
	svc *service.ItemService
}

func NewItem(svc *service.ItemService) *ItemHandler {
	return &ItemHandler{svc: svc}
}

func (h *ItemHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, common.OK("pong"))
}

// CreateItem：入参支持 query 与 form 两种承载（验收契约无关，两者都要能用）。
func (h *ItemHandler) CreateItem(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		name = c.PostForm("name")
	}
	var score *int
	if sv := c.Query("score"); sv != "" {
		if n, err := strconv.Atoi(sv); err == nil {
			score = &n
		}
	} else if sv := c.PostForm("score"); sv != "" {
		if n, err := strconv.Atoi(sv); err == nil {
			score = &n
		}
	}
	it, err := h.svc.CreateItem(c.Request.Context(), name, score)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(it))
}

func (h *ItemHandler) List(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	data, err := h.svc.PageItems(c.Request.Context(), pageNum, pageSize, c.Query("name"))
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(data))
}

func (h *ItemHandler) Cache(c *gin.Context) {
	v, err := h.svc.Cache(c.Request.Context(), c.Query("key"), c.Query("value"))
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(v))
}
