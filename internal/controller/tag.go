package controller

import (
	"encoding/json"
	"net/http"

	"rcloneflow/internal/service"
	"rcloneflow/internal/store"
)

type TagController struct {
	svc *service.TagService
}

func NewTagController(svc *service.TagService) *TagController {
	return &TagController{svc: svc}
}

func (c *TagController) HandleTags(w http.ResponseWriter, r *http.Request) {
	tags, err := c.svc.ListTags()
	if err != nil {
		http.Error(w, `{"error":"获取标签失败"}`, http.StatusInternalServerError)
		return
	}
	if tags == nil {
		tags = []store.Tag{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"tags": tags,
	})
}