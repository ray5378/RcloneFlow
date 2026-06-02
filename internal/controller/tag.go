package controller

import (
	"encoding/json"
	"net/http"
	"strings"

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

func (c *TagController) HandleTagActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if strings.HasSuffix(r.URL.Path, "/select") {
		c.handleSelect(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/create") {
		c.handleCreate(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/delete") {
		c.handleDelete(w, r)
		return
	}

	http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
}

func (c *TagController) handleSelect(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Tag      string `json:"tag"`
		Selected bool   `json:"selected"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	if body.Tag == "" {
		http.Error(w, `{"error":"tag is required"}`, http.StatusBadRequest)
		return
	}
	if err := c.svc.SelectTag(body.Tag, body.Selected); err != nil {
		http.Error(w, `{"error":"操作失败"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (c *TagController) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Tag string `json:"tag"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	if body.Tag == "" {
		http.Error(w, `{"error":"tag is required"}`, http.StatusBadRequest)
		return
	}
	if err := c.svc.CreateManualTag(body.Tag); err != nil {
		http.Error(w, `{"error":"创建失败"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (c *TagController) handleDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Tag string `json:"tag"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}
	if body.Tag == "" {
		http.Error(w, `{"error":"tag is required"}`, http.StatusBadRequest)
		return
	}
	if err := c.svc.DeleteManualTag(body.Tag); err != nil {
		http.Error(w, `{"error":"删除失败"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
