package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ludashi-bailian/models"
	"ludashi-bailian/services"
)

type VideoHandler struct {
	DB               *gorm.DB
	DashScopeService *services.DashScopeService
}

func NewVideoHandler(db *gorm.DB, dashScopeService *services.DashScopeService) *VideoHandler {
	return &VideoHandler{
		DB:               db,
		DashScopeService: dashScopeService,
	}
}

// CreateVideoTask 创建视频生成任务
func (h *VideoHandler) CreateVideoTask(c *gin.Context) {
	var req models.VideoGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	// 设置默认值
	if req.Duration == 0 {
		req.Duration = 5
	}
	if req.Resolution == "" {
		req.Resolution = "720P"
	}

	// 验证任务类型
	if req.TaskType != "i2v-first-frame" && req.TaskType != "i2v-keyframes" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的任务类型"})
		return
	}

	// 验证模型
	if req.Model != "wanx2.1-i2v-turbo" && req.Model != "wanx2.1-i2v-plus" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的模型"})
		return
	}

	// 验证prompt长度
	if len([]rune(req.Prompt)) > 800 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "提示词长度不能超过800个字符"})
		return
	}

	// 验证模型与任务类型的匹配
	if req.TaskType == "i2v-keyframes" && req.Model != "wanx2.1-i2v-plus" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "首尾帧任务只支持wanx2.1-i2v-plus模型"})
		return
	}

	// 验证分辨率
	if req.Model == "wanx2.1-i2v-turbo" && req.Resolution != "480P" && req.Resolution != "720P" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wanx2.1-i2v-turbo模型只支持480P和720P分辨率"})
		return
	}
	if req.Model == "wanx2.1-i2v-plus" && req.Resolution != "720P" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wanx2.1-i2v-plus模型只支持720P分辨率"})
		return
	}

	// 验证时长
	if req.Model == "wanx2.1-i2v-turbo" && (req.Duration < 3 || req.Duration > 5) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wanx2.1-i2v-turbo模型支持3-5秒时长"})
		return
	}
	if req.Model == "wanx2.1-i2v-plus" && req.Duration != 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wanx2.1-i2v-plus模型只支持5秒时长"})
		return
	}

	// 验证种子范围
	if req.Seed != nil && (*req.Seed < 0 || *req.Seed > 2147483647) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "随机数种子必须在0-2147483647范围内"})
		return
	}

	// 如果是首尾帧任务，检查是否提供了结束帧
	if req.TaskType == "i2v-keyframes" && req.EndImageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "首尾帧任务需要提供结束帧图片"})
		return
	}

	// 创建数据库记录
	taskRecord := models.TaskRequest{
		TaskType:     req.TaskType,
		Model:        req.Model, // 使用用户选择的模型
		Prompt:       req.Prompt,
		ImageURL:     req.ImageURL,
		EndImageURL:  req.EndImageURL,
		Status:       "pending",
		Duration:     req.Duration,
		Resolution:   req.Resolution,
		Seed:         req.Seed,
		PromptExtend: req.PromptExtend,
	}

	if err := h.DB.Create(&taskRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务记录失败"})
		return
	}

	// 异步调用DashScope API
	go h.processVideoTask(&taskRecord, &req)

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已创建",
		"task_id": taskRecord.ID,
		"status":  "pending",
	})
}

// GetTaskStatus 获取任务状态
func (h *VideoHandler) GetTaskStatus(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	var task models.TaskRequest
	if err := h.DB.First(&task, uint(taskID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务失败"})
		return
	}

	c.JSON(http.StatusOK, models.TaskStatusResponse{
		TaskID:    task.TaskID,
		Status:    task.Status,
		VideoURL:  task.VideoURL,
		Error:     task.Error,
		RequestID: task.RequestID,
	})
}

// GetTaskHistory 获取历史任务列表
func (h *VideoHandler) GetTaskHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	taskType := c.Query("task_type")
	status := c.Query("status")

	offset := (page - 1) * pageSize

	query := h.DB.Model(&models.TaskRequest{})
	if taskType != "" {
		query = query.Where("task_type = ?", taskType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var tasks []models.TaskRequest
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询历史记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tasks,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetTaskDetail 获取任务详情
func (h *VideoHandler) GetTaskDetail(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	var task models.TaskRequest
	if err := h.DB.First(&task, uint(taskID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": task})
}

// DeleteTask 删除任务
func (h *VideoHandler) DeleteTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	if err := h.DB.Delete(&models.TaskRequest{}, uint(taskID)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已删除"})
}

// processVideoTask 处理视频生成任务（后台执行）
func (h *VideoHandler) processVideoTask(taskRecord *models.TaskRequest, req *models.VideoGenerationRequest) {
	// 更新状态为运行中
	h.DB.Model(taskRecord).Update("status", "running")

	// 调用DashScope API创建任务
	resp, err := h.DashScopeService.CreateVideoGenerationTask(req)
	if err != nil {
		h.DB.Model(taskRecord).Updates(models.TaskRequest{
			Status: "failed",
			Error:  err.Error(),
		})
		return
	}

	// 更新任务ID和请求ID
	h.DB.Model(taskRecord).Updates(models.TaskRequest{
		TaskID:    resp.Output.TaskID,
		RequestID: resp.RequestID,
	})

	// 轮询任务状态
	finalResp, err := h.DashScopeService.PollTaskStatus(resp.Output.TaskID, 15*time.Minute)
	if err != nil {
		h.DB.Model(taskRecord).Updates(models.TaskRequest{
			Status: "failed",
			Error:  err.Error(),
		})
		return
	}

	// 更新最终结果
	successTime := time.Now()
	expireTime := successTime.Add(24 * time.Hour) // 24小时后过期
	h.DB.Model(taskRecord).Updates(models.TaskRequest{
		Status:      "succeeded",
		VideoURL:    finalResp.Output.VideoURL,
		SuccessTime: &successTime,
		ExpireTime:  &expireTime,
	})
}

// getModelByTaskType 根据任务类型获取模型名称
func getModelByTaskType(taskType string) string {
	switch taskType {
	case "i2v-first-frame":
		return "wanx2.1-i2v-turbo"
	case "i2v-keyframes":
		return "wanx2.1-i2v-plus"
	default:
		return "unknown"
	}
}
