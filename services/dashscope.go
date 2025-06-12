package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"ludashi-bailian/models"
)

const (
	DashScopeBaseURL       = "https://dashscope.aliyuncs.com/api/v1"
	VideoSynthesisEndpoint = "/services/aigc/video-generation/video-synthesis"
	TaskStatusEndpoint     = "/tasks"
)

type DashScopeService struct {
	APIKey string
	Client *http.Client
}

func NewDashScopeService() *DashScopeService {
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		log.Println("Warning: DASHSCOPE_API_KEY环境变量未设置，请在使用前设置正确的API Key")
		apiKey = "placeholder-key" // 设置占位符，避免panic
	}

	return &DashScopeService{
		APIKey: apiKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateVideoGenerationTask 创建视频生成任务
func (s *DashScopeService) CreateVideoGenerationTask(req *models.VideoGenerationRequest) (*models.DashScopeVideoResponse, error) {
	// 截断提示词到800字符
	prompt := req.Prompt
	if len([]rune(prompt)) > 800 {
		prompt = string([]rune(prompt)[:800])
	}

	// 构建请求数据
	taskReq := models.DashScopeTaskRequest{
		Model: req.Model, // 使用用户选择的模型
		Input: models.DashScopeTaskInput{
			Prompt: prompt,
		},
		Parameters: models.DashScopeTaskParameters{
			Duration: req.Duration,
		},
	}

	// 根据任务类型设置不同的输入参数
	switch req.TaskType {
	case "i2v-first-frame":
		taskReq.Input.ImageURL = req.ImageURL
	case "i2v-keyframes":
		taskReq.Input.ImageURL = req.ImageURL
		if req.EndImageURL != "" {
			taskReq.Input.EndImageURL = req.EndImageURL
		}
	case "t2v":
		// 文生视频不需要图片输入
	}

	// 设置分辨率参数
	if req.Size != "" {
		taskReq.Parameters.Size = req.Size
	} else if req.Resolution != "" {
		// 如果没有指定具体尺寸，根据分辨率档位设置默认值
		switch req.Resolution {
		case "480P":
			taskReq.Parameters.Size = "832*480" // 默认16:9
		case "720P":
			taskReq.Parameters.Size = "1280*720" // 默认16:9
		}
	}

	// 添加高级参数
	if req.PromptExtend != nil {
		taskReq.Parameters.PromptExtend = req.PromptExtend
	}
	if req.Seed != nil {
		taskReq.Parameters.Seed = req.Seed
	}

	// 序列化请求
	jsonData, err := json.Marshal(taskReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	url := DashScopeBaseURL + VideoSynthesisEndpoint
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.APIKey)
	httpReq.Header.Set("X-DashScope-Async", "enable")

	// 发送请求
	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var dashScopeResp models.DashScopeVideoResponse
	if err := json.Unmarshal(body, &dashScopeResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &dashScopeResp, nil
}

// GetTaskStatus 获取任务状态
func (s *DashScopeService) GetTaskStatus(taskID string) (*models.DashScopeVideoResponse, error) {
	url := fmt.Sprintf("%s%s/%s", DashScopeBaseURL, TaskStatusEndpoint, taskID)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.APIKey)

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var dashScopeResp models.DashScopeVideoResponse
	if err := json.Unmarshal(body, &dashScopeResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &dashScopeResp, nil
}

// PollTaskStatus 轮询任务状态直到完成
func (s *DashScopeService) PollTaskStatus(taskID string, maxWaitTime time.Duration) (*models.DashScopeVideoResponse, error) {
	pollInterval := 10 * time.Second
	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("任务轮询超时")
		case <-ticker.C:
			resp, err := s.GetTaskStatus(taskID)
			if err != nil {
				return nil, err
			}

			switch resp.Output.TaskStatus {
			case "SUCCEEDED":
				return resp, nil
			case "FAILED":
				return nil, fmt.Errorf("任务执行失败")
			case "PENDING", "RUNNING":
				// 继续等待
				continue
			default:
				return nil, fmt.Errorf("未知任务状态: %s", resp.Output.TaskStatus)
			}
		}
	}
}
