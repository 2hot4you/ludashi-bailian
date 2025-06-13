package models

import (
	"time"

	"gorm.io/gorm"
)

// TaskRequest 任务请求记录
type TaskRequest struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	TaskType     string         `json:"task_type" gorm:"type:varchar(100);not null"`       // i2v-first-frame, i2v-keyframes, t2v, image_reference, video_repainting
	Model        string         `json:"model" gorm:"type:varchar(100);not null"`           // 模型名称
	Prompt       string         `json:"prompt" gorm:"type:text"`                           // 文本提示词
	ImageURL     string         `json:"image_url" gorm:"type:text"`                        // 图片URL（首帧）
	EndImageURL  string         `json:"end_image_url" gorm:"type:text"`                    // 结束帧图片URL（可选）
	TaskID       string         `json:"task_id" gorm:"type:varchar(100)"`                  // DashScope返回的任务ID
	Status       string         `json:"status" gorm:"type:varchar(50);default:'pending'"`  // pending, running, succeeded, failed
	VideoURL     string         `json:"video_url" gorm:"type:text"`                        // 生成的视频URL
	Error        string         `json:"error" gorm:"type:text"`                            // 错误信息
	RequestID    string         `json:"request_id" gorm:"type:varchar(100)"`               // 请求ID
	Duration     int            `json:"duration" gorm:"default:5"`                         // 视频时长
	Resolution   string         `json:"resolution" gorm:"type:varchar(20);default:'720P'"` // 视频分辨率
	Size         string         `json:"size" gorm:"type:varchar(20)"`                      // 具体分辨率大小 (如 1280*720)
	SuccessTime  *time.Time     `json:"success_time" gorm:"type:datetime"`                 // 任务成功时间
	ExpireTime   *time.Time     `json:"expire_time" gorm:"type:datetime"`                  // 视频过期时间（成功后24小时）
	Seed         *int           `json:"seed" gorm:"type:int"`                              // 随机数种子
	PromptExtend *bool          `json:"prompt_extend" gorm:"type:boolean"`                 // 是否开启prompt智能改写

	// 新增：多图参考和视频重绘相关字段
	RefImagesURL     string  `json:"ref_images_url" gorm:"type:text"`           // 参考图片URL列表 JSON存储
	ObjOrBg          string  `json:"obj_or_bg" gorm:"type:text"`                // 图片类型数组 JSON存储
	VideoURL2        string  `json:"video_url_input" gorm:"type:text"`          // 输入视频URL（用于视频重绘）
	ControlCondition string  `json:"control_condition" gorm:"type:varchar(50)"` // 控制条件
	Strength         float64 `json:"strength" gorm:"type:decimal(3,2)"`         // 控制强度
	Watermark        *bool   `json:"watermark" gorm:"type:boolean"`             // 水印设置
}

// VideoGenerationRequest 视频生成请求
type VideoGenerationRequest struct {
	TaskType     string `json:"task_type" binding:"required"` // i2v-first-frame, i2v-keyframes, t2v
	Model        string `json:"model" binding:"required"`     // 模型名称
	Prompt       string `json:"prompt" binding:"required"`    // 文本提示词（最大800字符）
	ImageURL     string `json:"image_url"`                    // 图片URL（i2v任务需要）
	EndImageURL  string `json:"end_image_url"`                // 结束帧图片URL（仅i2v-keyframes需要）
	Duration     int    `json:"duration"`                     // 视频时长（秒）
	Resolution   string `json:"resolution"`                   // 视频分辨率（480P/720P）
	Size         string `json:"size"`                         // 具体分辨率大小 (如 1280*720)
	PromptExtend *bool  `json:"prompt_extend"`                // 是否开启prompt智能改写
	Seed         *int   `json:"seed"`                         // 随机数种子[0, 2147483647]
}

// TaskStatusResponse 任务状态响应
type TaskStatusResponse struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	VideoURL   string `json:"video_url,omitempty"`
	Error      string `json:"error,omitempty"`
	RequestID  string `json:"request_id"`
	SubmitTime string `json:"submit_time,omitempty"`
	EndTime    string `json:"end_time,omitempty"`
}

// DashScopeVideoResponse DashScope API响应结构
type DashScopeVideoResponse struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		VideoURL   string `json:"video_url"`
		SubmitTime string `json:"submit_time"`
		EndTime    string `json:"end_time"`
	} `json:"output"`
	RequestID string `json:"request_id"`
	Usage     struct {
		VideoDuration int    `json:"video_duration"`
		VideoRatio    string `json:"video_ratio"`
		VideoCount    int    `json:"video_count"`
	} `json:"usage"`
}

// DashScopeTaskRequest DashScope任务创建请求
type DashScopeTaskRequest struct {
	Model      string                  `json:"model"`
	Input      DashScopeTaskInput      `json:"input"`
	Parameters DashScopeTaskParameters `json:"parameters"`
}

type DashScopeTaskInput struct {
	Prompt       string   `json:"prompt"`
	ImageURL     string   `json:"img_url,omitempty"`
	EndImageURL  string   `json:"end_img_url,omitempty"`
	Function     string   `json:"function,omitempty"`       // 功能名称: image_reference, video_repainting
	RefImagesURL []string `json:"ref_images_url,omitempty"` // 参考图片URL列表
	VideoURL     string   `json:"video_url,omitempty"`      // 输入视频URL
}

type DashScopeTaskParameters struct {
	Size             string   `json:"size,omitempty"`
	Duration         int      `json:"duration"`
	PromptExtend     *bool    `json:"prompt_extend,omitempty"`
	Seed             *int     `json:"seed,omitempty"`
	Watermark        *bool    `json:"watermark,omitempty"`         // 水印
	ObjOrBg          []string `json:"obj_or_bg,omitempty"`         // 图片类型数组
	ControlCondition string   `json:"control_condition,omitempty"` // 控制条件
	Strength         float64  `json:"strength,omitempty"`          // 控制强度
}

// VideoCreateRequest 视频生成请求（统一接口）
type VideoCreateRequest struct {
	TaskType    string `json:"task_type" binding:"required"` // 任务类型: i2v-first-frame, i2v-keyframes, t2v, image_reference, video_repainting
	Model       string `json:"model" binding:"required"`     // 模型名称
	Prompt      string `json:"prompt" binding:"required"`    // 文本提示词
	ImageURL    string `json:"image_url"`                    // 图片URL (首帧图片)
	EndImageURL string `json:"end_image_url"`                // 结束帧图片URL
	Duration    int    `json:"duration"`                     // 视频时长
	Resolution  string `json:"resolution"`                   // 分辨率
	Size        string `json:"size"`                         // 具体分辨率大小

	// 新增：多图参考参数
	RefImagesURL []string `json:"ref_images_url"` // 参考图片URL列表 (1-3张)
	ObjOrBg      []string `json:"obj_or_bg"`      // 图片类型数组 (obj/bg)

	// 新增：视频重绘参数
	VideoURL         string  `json:"video_url"`         // 输入视频URL
	ControlCondition string  `json:"control_condition"` // 控制条件: posebodyface, posebody, depth, scribble
	Strength         float64 `json:"strength"`          // 控制强度 (0.0-1.0)

	// 高级参数
	PromptExtend *bool `json:"prompt_extend"` // 智能prompt改写
	Seed         *int  `json:"seed"`          // 随机种子
	Watermark    *bool `json:"watermark"`     // 水印
}
