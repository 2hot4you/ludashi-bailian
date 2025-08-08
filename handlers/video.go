package handlers

import (
	"encoding/json"
	"fmt"
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
	var req models.VideoCreateRequest
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
	if req.TaskType != "i2v-first-frame" && req.TaskType != "i2v-keyframes" && req.TaskType != "t2v" &&
		req.TaskType != "image_reference" && req.TaskType != "video_repainting" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的任务类型"})
		return
	}

	// 验证模型
	validModels := []string{
		"wanx2.1-i2v-turbo", "wanx2.1-i2v-plus",
		"wanx2.1-t2v-turbo", "wanx2.1-t2v-plus",
		"wanx2.1-vace-plus", // 通用视频编辑模型
		"wanx2.2-t2v-plus", // 通义万相2.2专业版 - 文生视频
	}
	isValidModel := false
	for _, model := range validModels {
		if req.Model == model {
			isValidModel = true
			break
		}
	}
	if !isValidModel {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的模型"})
		return
	}

	// 验证prompt长度
	if len([]rune(req.Prompt)) > 800 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "提示词长度不能超过800个字符"})
		return
	}

	// 验证任务类型与模型的匹配
	switch req.TaskType {
	case "i2v-first-frame", "i2v-keyframes":
		if req.Model != "wanx2.1-i2v-turbo" && req.Model != "wanx2.1-i2v-plus" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "图生视频任务只支持wanx2.1-i2v-turbo和wanx2.1-i2v-plus模型"})
			return
		}
		// 检查是否提供了图片URL
		if req.ImageURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "图生视频任务需要提供图片"})
			return
		}
	case "t2v":
		if req.Model != "wanx2.1-t2v-turbo" && req.Model != "wanx2.1-t2v-plus" && req.Model != "wanx2.2-t2v-plus" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文生视频任务只支持wanx2.1-t2v-turbo, wanx2.1-t2v-plus 和 wanx2.2-t2v-plus 模型"})
			return
		}
	case "image_reference":
		if req.Model != "wanx2.1-vace-plus" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "多图参考任务只支持wanx2.1-vace-plus模型"})
			return
		}
		// 验证参考图片URL
		if len(req.RefImagesURL) == 0 || len(req.RefImagesURL) > 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "多图参考任务需要提供1-3张参考图片"})
			return
		}
		// 验证图片类型数组
		if len(req.ObjOrBg) > 0 && len(req.ObjOrBg) != len(req.RefImagesURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "图片类型数组长度必须与图片数量一致"})
			return
		}
		// 验证背景图片数量
		bgCount := 0
		for _, t := range req.ObjOrBg {
			if t == "bg" {
				bgCount++
			}
		}
		if bgCount > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "最多只能有一张背景图片"})
			return
		}
	case "video_repainting":
		if req.Model != "wanx2.1-vace-plus" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "视频重绘任务只支持wanx2.1-vace-plus模型"})
			return
		}
		// 验证输入视频URL
		if req.VideoURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "视频重绘任务需要提供输入视频URL"})
			return
		}
		// 验证控制条件（必选）
		if req.ControlCondition == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "视频重绘任务需要指定特征提取方式(control_condition)，这是必选参数"})
			return
		}
		validConditions := []string{"posebodyface", "posebody", "depth", "scribble"}
		isValidCondition := false
		for _, condition := range validConditions {
			if req.ControlCondition == condition {
				isValidCondition = true
				break
			}
		}
		if !isValidCondition {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的特征提取方式，支持: posebodyface(脸部+肢体), posebody(仅肢体), depth(构图轮廓), scribble(线稿)"})
			return
		}
		// 验证控制强度（可选，默认1.0）
		if req.Strength == 0.0 {
			req.Strength = 1.0 // 设置默认值
		}
		if req.Strength < 0.0 || req.Strength > 1.0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "控制强度必须在0.0-1.0范围内，默认为1.0"})
			return
		}
		// 验证参考图片数量
		if len(req.RefImagesURL) > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "视频重绘任务最多只能使用1张参考图片"})
			return
		}
	}

	// 验证首尾帧任务的特殊要求
	if req.TaskType == "i2v-keyframes" {
		if req.Model != "wanx2.1-i2v-plus" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "首尾帧任务只支持wanx2.1-i2v-plus模型"})
			return
		}
		if req.EndImageURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "首尾帧任务需要提供结束帧图片"})
			return
		}
	}

	// 验证分辨率
	if req.Resolution != "" {
		supportedResolutions := map[string][]string{
			"wanx2.1-i2v-turbo": {"480P", "720P"},
			"wanx2.1-i2v-plus":  {"720P"},
			"wanx2.1-t2v-turbo": {"480P", "720P"},
			"wanx2.1-t2v-plus":  {"720P"},
			"wanx2.1-vace-plus": {"1280*720", "720*1280", "960*960", "832*1088", "1088*832"}, // For VACE-Plus, resolution is the exact size
			"wanx2.2-t2v-plus":  {"480P", "1080P"},                                           // New model resolutions
		}

		if resolutions, ok := supportedResolutions[req.Model]; ok {
			isSupported := false
			for _, res := range resolutions {
				if req.Resolution == res {
					isSupported = true
					break
				}
			}
			if !isSupported {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("模型 %s 不支持分辨率 %s", req.Model, req.Resolution)})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("模型 %s 的分辨率列表未定义", req.Model)})
			return
		}
	}

	// 验证时长 - 固定为5秒
	if req.Duration != 5 {
		req.Duration = 5 // 强制设置为5秒
	}

	// 验证种子范围
	if req.Seed != nil && (*req.Seed < 0 || *req.Seed > 2147483647) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "随机数种子必须在0-2147483647范围内"})
		return
	}

	// 处理分辨率大小设置
	if req.Size == "" && req.Resolution != "" {
		if req.Model == "wanx2.1-vace-plus" {
			req.Size = req.Resolution // For VACE-Plus, resolution is directly the size
		} else if req.Model == "wanx2.2-t2v-plus" { // Handle wanx2.2-t2v-plus specific sizes
			switch req.Resolution {
			case "480P":
				req.Size = "832*480" // Example default size for 480P
			case "1080P":
				req.Size = "1920*1080" // Example default size for 1080P
			default:
				// Fallback or error if resolution doesn't match expected for wanx2.2-t2v-plus
				req.Size = "1920*1080" // Default to 1080P if not specified or unrecognized
			}
		} else {
			// Other models use default sizes based on resolution tier
			switch req.Resolution {
			case "480P":
				req.Size = "832*480"
			case "720P":
				req.Size = "1280*720"
			}
		}
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
		Size:         req.Size,
		Seed:         req.Seed,
		PromptExtend: req.PromptExtend,

		// 新增字段支持
		VideoURL2:        req.VideoURL,
		ControlCondition: req.ControlCondition,
		Strength:         req.Strength,
		Watermark:        req.Watermark,
	}

	// 处理数组字段的JSON序列化
	if len(req.RefImagesURL) > 0 {
		if refImagesJSON, err := json.Marshal(req.RefImagesURL); err == nil {
			taskRecord.RefImagesURL = string(refImagesJSON)
		}
	}
	if len(req.ObjOrBg) > 0 {
		if objOrBgJSON, err := json.Marshal(req.ObjOrBg); err == nil {
			taskRecord.ObjOrBg = string(objOrBgJSON)
		}
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
func (h *VideoHandler) processVideoTask(taskRecord *models.TaskRequest, req *models.VideoCreateRequest) {
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
