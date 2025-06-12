package main

import (
	"log"
	"net/http"
	"os"

	"ludashi-bailian/handlers"
	"ludashi-bailian/models"
	"ludashi-bailian/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 初始化数据库
	db, err := gorm.Open(sqlite.Open("bailian.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自动迁移数据库
	err = db.AutoMigrate(&models.TaskRequest{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建DashScope服务实例
	dashScopeService := services.NewDashScopeService()

	// 创建处理器实例
	videoHandler := handlers.NewVideoHandler(db, dashScopeService)

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 静态文件服务
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// 静态文件服务（支持上传的图片访问）
	r.Static("/uploads", "./uploads")

	// API路由
	api := r.Group("/api")
	{
		video := api.Group("/video")
		{
			video.POST("/create", videoHandler.CreateVideoTask)
			video.GET("/status/:id", videoHandler.GetTaskStatus)
			video.GET("/detail/:id", videoHandler.GetTaskDetail)
			video.GET("/history", videoHandler.GetTaskHistory)
			video.DELETE("/:id", videoHandler.DeleteTask)
		}
	}

	// 页面路由
	r.GET("/video-generation", func(c *gin.Context) {
		c.File("./static/video-generation.html")
	})

	r.GET("/history", func(c *gin.Context) {
		c.File("./static/history.html")
	})

	r.GET("/task-detail/:id", func(c *gin.Context) {
		c.File("./static/task-detail.html")
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Bailian video generation service is running",
		})
	})

	// 检查环境变量
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		log.Println("Warning: DASHSCOPE_API_KEY environment variable is not set")
		log.Println("Please set it before making API calls: export DASHSCOPE_API_KEY=your_api_key")
	}

	// 确保static目录存在
	if _, err := os.Stat("static"); os.IsNotExist(err) {
		err := os.MkdirAll("static", 0755)
		if err != nil {
			log.Fatal("Failed to create static directory:", err)
		}
	}

	log.Println("Starting Bailian video generation service...")
	log.Println("Server will start on http://localhost:8082")
	log.Println("API endpoints:")
	log.Println("  GET  /                     - Home page")
	log.Println("  GET  /i2v-first-frame     - First frame video generation")
	log.Println("  GET  /i2v-keyframes       - Keyframes video generation")
	log.Println("  POST /api/video/create    - Create video task")
	log.Println("  GET  /api/video/status/:id - Get task status")
	log.Println("  GET  /api/video/history   - Get task history")
	log.Println("  DELETE /api/video/:id     - Delete task")
	log.Println("  GET  /health              - Health check")

	// 启动服务器，端口设置为8082
	if err := r.Run(":8082"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
