# 🎬 阿里云百炼视频生成平台

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)](LICENSE)
[![DashScope](https://img.shields.io/badge/Powered%20by-阿里云百炼-orange?style=for-the-badge)](https://bailian.console.aliyun.com/)

> 🚀 基于阿里云百炼DashScope API的智能视频生成平台，支持通义万相图生视频功能

## 📸 界面预览

<!-- TODO: 添加实际的截图 -->
> 💡 在这里添加您的应用截图，展示主要功能界面

## ✨ 功能特性

### 🎯 核心功能
- 🖼️ **基于首帧生成** - 上传单张图片生成动态视频
- 🎭 **基于首尾帧生成** - 提供首尾两帧生成过渡视频
- 🤖 **多模型支持** - turbo(快速) 和 plus(高质量) 两种模型
- 📝 **智能提示词** - 800字符限制，实时计数，AI改写

### 🛠️ 技术特性
- ⚡ **实时预览** - 图片URL输入后立即显示预览
- 🎲 **种子控制** - 支持固定种子保证生成结果稳定
- 📊 **任务管理** - 完整的历史记录、状态跟踪、分页浏览
- 🎨 **现代UI** - 响应式设计，支持移动端访问
- 🔄 **异步处理** - 后台任务处理，实时状态更新

## 🏗️ 技术栈

### 后端
- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **数据库**: SQLite (GORM ORM)
- **API**: 阿里云DashScope API

### 前端
- **框架**: 原生HTML + CSS + JavaScript
- **UI库**: Bootstrap 5.3
- **图标**: Font Awesome 6.0
- **样式**: CSS3 渐变 + 响应式设计

## 🚀 快速开始

### 📋 环境要求

- Go 1.21 或更高版本
- 有效的阿里云DashScope API Key
- 网络连接（访问阿里云API）

### 🔧 安装步骤

1. **克隆项目**
   ```bash
   git clone https://github.com/your-username/ludashi-bailian.git
   cd ludashi-bailian
   ```

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **设置API密钥**
   ```bash
   export DASHSCOPE_API_KEY="your_dashscope_api_key"
   ```
   
   或者在 `~/.bashrc` 中永久设置：
   ```bash
   echo 'export DASHSCOPE_API_KEY="your_dashscope_api_key"' >> ~/.bashrc
   source ~/.bashrc
   ```

4. **启动服务**
   ```bash
   go run main.go
   ```

5. **访问应用**
   打开浏览器访问: http://localhost:8082

### 🌐 页面导航

| 页面 | URL | 描述 |
|------|-----|------|
| 🏠 首页 | `/` | 平台介绍和功能入口 |
| 🎬 视频生成 | `/video-generation.html` | 统一的视频生成页面（支持首帧和首尾帧） |
| 📋 历史记录 | `/history.html` | 查看和管理任务历史 |
| 📄 任务详情 | `/task-detail.html` | 查看具体任务详情 |
| ❤️ 健康检查 | `/health` | 服务健康状态 |

## 📖 API 文档

### 🔥 主要接口

#### 创建视频任务
```http
POST /api/video/create
Content-Type: application/json

{
  "task_type": "i2v-first-frame",     // 任务类型
  "model": "wanx2.1-i2v-turbo",       // 模型选择
  "prompt": "描述视频内容",           // 提示词(≤800字符)
  "image_url": "https://example.com/image.jpg",
  "end_image_url": "https://example.com/end.jpg",  // 首尾帧模式必需
  "duration": 5,                      // 视频时长(秒)
  "resolution": "720P",               // 分辨率
  "prompt_extend": true,              // 智能改写
  "seed": 12345                       // 随机种子
}
```

#### 查询任务状态
```http
GET /api/video/status/{task_id}
```

#### 获取历史记录
```http
GET /api/video/history?page=1&page_size=10&task_type=i2v-first-frame&status=succeeded
```

#### 删除任务
```http
DELETE /api/video/{task_id}
```

### 📊 模型对比

| 模型 | 场景支持 | 分辨率 | 时长 | 特点 |
|------|----------|--------|------|------|
| `wanx2.1-i2v-turbo` | 基于首帧 | 480P/720P | 3-5秒 | ⚡ 速度快，适合预览 |
| `wanx2.1-i2v-plus` | 首帧/首尾帧 | 720P | 5秒 | 🎨 高质量，复杂场景 |
| `wanx2.1-t2v-turbo` | 文生视频 | 480P/720P | 5秒 | ⚡ 速度快，表现均衡 |
| `wanx2.1-t2v-plus` | 文生视频 | 720P | 5秒 | ✨ 细节丰富，画面质感 |
| `wanx2.1-vace-plus` | 通用视频编辑 | 多种比例 | 5秒 | 🎬 多图参考，视频重绘 |
| `wanx2.2-t2v-plus` | 文生视频 | 480P/1080P | 5秒 | 🚀 万相2.2专业版，细节和运动更佳 |

## 🏗️ 项目结构

```
ludashi-bailian/
├── 📁 handlers/            # HTTP请求处理器
│   └── video.go           # 视频任务相关处理
├── 📁 models/              # 数据模型定义
│   └── models.go          # 任务和响应模型
├── 📁 services/            # 业务逻辑服务
│   └── dashscope.go       # DashScope API集成
├── 📁 static/              # 静态资源文件
│   ├── index.html         # 🏠 首页
│   ├── video-generation.html  # 🎬 统一视频生成页
│   ├── history.html       # 📋 历史记录页
│   └── task-detail.html   # 📄 任务详情页
├── 📄 main.go              # 🚀 程序入口
├── 📄 go.mod              # 📦 依赖管理
├── 📄 go.sum              # 依赖校验
├── 📄 README.md           # 📖 项目文档
├── 📄 LICENSE             # 📄 开源许可证
├── 📄 .gitignore          # 🚫 Git忽略文件
└── 📄 bailian.db          # 🗄️ SQLite数据库(自动生成)
```

## 🎯 使用指南

### 🖼️ 基于首帧生成视频

1. 在视频生成页面选择"基于首帧生成"模式
2. 输入图片URL或上传图片
3. 编写描述视频内容的提示词
4. 选择模型和参数（turbo或plus）
5. 点击"开始生成"等待结果

### 🎭 基于首尾帧生成视频

1. 在视频生成页面选择"基于首尾帧生成"模式  
2. 提供起始帧和结束帧图片
3. 描述转换过程
4. 选择plus模型(turbo不支持此模式)
5. 生成平滑过渡视频

## 🛠️ 开发指南

### 🔧 本地开发

```bash
# 开发模式启动(热重载)
go run main.go

# 构建生产版本
go build -o bailian-platform main.go

# 运行测试
go test ./...
```

### 🐳 Docker部署

```dockerfile
# TODO: 添加Dockerfile
FROM golang:1.21-alpine AS builder
# ... 构建步骤
```

## 🔐 环境变量

| 变量名 | 必需 | 默认值 | 描述 |
|--------|------|--------|------|
| `DASHSCOPE_API_KEY` | ✅ | - | 阿里云DashScope API密钥 |
| `PORT` | ❌ | `8082` | 服务监听端口 |
| `DB_PATH` | ❌ | `./bailian.db` | SQLite数据库路径 |

## 🎨 自定义配置

### 修改端口
```go
// main.go
r.Run(":8080")  // 修改为你想要的端口
```

### 数据库配置
项目使用SQLite，数据库文件会自动创建在项目根目录。如需更改位置，修改相关代码中的数据库路径。

## 🔍 故障排除

### 常见问题

**Q: API Key未设置警告**
```
Warning: DASHSCOPE_API_KEY environment variable is not set
```
**A**: 设置环境变量 `export DASHSCOPE_API_KEY="your_key"`

**Q: 无法访问阿里云API**  
**A**: 检查网络连接和API Key有效性

**Q: 视频生成失败**  
**A**: 查看任务详情页面的错误信息，检查图片URL和提示词

**Q: 历史记录页面显示不完整**  
**A**: 已修复，支持水平滚动查看所有列

## 🔗 相关链接

- 📚 [阿里云百炼平台](https://bailian.console.aliyun.com/)
- 📖 [DashScope API 文档](https://help.aliyun.com/document_detail/2867393.html)
- 🤖 [通义万相模型介绍](https://help.aliyun.com/document_detail/2867393.html)
- 🎯 [Go Gin 框架](https://gin-gonic.com/)

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🤝 贡献

我们欢迎所有形式的贡献！

### 贡献方式
1. 🍴 Fork 本项目
2. 🌟 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 💫 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 📤 推送到分支 (`git push origin feature/AmazingFeature`)
5. 🎉 创建 Pull Request

### 问题反馈
- 🐛 [Bug报告](https://github.com/your-username/ludashi-bailian/issues)
- 💡 [功能建议](https://github.com/your-username/ludashi-bailian/issues)

## 👨‍💻 作者

**LUDASHI & IXIAOZU**

- 📧 Email: xibin@xistack.com
- 🐙 GitHub: [@2hot4you](https://github.com/2hot4you)

## 🙏 致谢

- 感谢阿里云百炼团队提供强大的AI模型服务
- 感谢开源社区的支持和贡献

## 📝 更新日志

### v1.2.1 (2025-01-22)
- ✨ **新增模型**: 在文生视频中集成 `wan2.2-t2v-plus` 模型
- 🔄 **模型支持**: 调整后端和前端，支持新模型的分辨率（480P, 1080P）
- 📖 **文档更新**: 更新 `README.md` 中模型对比表格和更新日志

### v1.2.0 (2025-01-21)
- 🔄 **界面优化**: 合并视频生成页面，统一首帧和首尾帧生成功能
- 🗑️ **清理文件**: 移除分离的 `i2v-first-frame.html` 和 `i2v-keyframes.html` 页面
- 🎨 **用户体验**: 改进页面布局和交互逻辑
- 📱 **响应式优化**: 增强移动端适配性

### v1.1.0 (2025-01-20)
- ✨ **新增功能**: 支持基于首帧和首尾帧的视频生成
- 🤖 **模型支持**: 集成 wanx2.1-i2v-turbo 和 wanx2.1-i2v-plus 模型
- 📊 **任务管理**: 完整的历史记录和状态跟踪系统
- 🎯 **API集成**: 阿里云DashScope API完全集成

### v1.0.0 (2025-01-19)
- 🎉 **首次发布**: 基础平台架构完成
- 🏗️ **技术栈**: Go + Gin + SQLite + Bootstrap
- 🎨 **UI设计**: 现代化响应式界面
- 📖 **文档**: 完整的README和API文档

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请点个星支持一下！ ⭐**

Made with ❤️ by LUDASHI & IXIAOZU

</div> 