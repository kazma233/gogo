# ShareWeb

局域网文件和文本分享工具。

## 功能特性

- 📝 文本消息分享
- 📁 文件上传和下载
- 📱 二维码移动端访问
- 🔒 文件类型和大小验证
- 🚀 基于 Fiber 框架的高性能
- 🐳 支持 Docker 和 Podman

## 配置说明

应用支持以下配置方式（按优先级排序）：

1. **环境变量**（最高优先级）
2. **默认值**（最低优先级）

### 配置选项

- `LC_PORT`: 服务器端口（默认：5667）
- `LC_DOWNLOAD_PATH`: 文件存储路径（默认：./download）
- `LC_MAX_FILE_SIZE`: 最大文件大小，单位字节（默认：100MB）
- `LC_ALLOWED_TYPES`: 允许的 MIME 类型，逗号分隔

### 环境变量示例

```bash
export LC_PORT=5667
export LC_DOWNLOAD_PATH=./download
export LC_MAX_FILE_SIZE=104857600
export LC_ALLOWED_TYPES=image/jpeg,image/png,image/gif,text/plain
```

## 快速开始

### 使用 Docker Compose（推荐）

```bash
docker-compose up -d
```

### 使用 Podman 和 Kubernetes YAML

```bash
# 构建镜像
podman build -f Containerfile -t localhost/lan-connector:latest .

# 启动服务
podman play kube kube.yaml

# 停止服务
podman play kube --down kube.yaml

# 查看状态
podman pod ps
podman ps
```

### 使用 Go 直接运行

```bash
go mod tidy
go run .
```

## API 接口

- `GET /` - Web 界面
- `GET /msg` - 列出所有消息
- `POST /msg` - 添加文本消息
- `DELETE /msg/:id` - 删除消息
- `POST /fs` - 上传文件
- `GET /download/:id` - 下载文件

## 安全特性

- 文件类型验证
- 文件大小限制
- 安全的文件名生成
- CORS 支持
- 优雅关闭

## 开发说明

项目结构：

- `main.go` - 应用入口
- `config.go` - 配置管理
- `handlers.go` - HTTP 请求处理器
- `middleware.go` - 自定义中间件
- `interfaces.go` - 存储接口
- `storage.go` - 原始存储实现，已废弃
- `improved_storage.go` - 增强的存储实现（当前使用）

## 存储特性

改进的存储系统提供：

- ✅ **异步持久化** - 非阻塞保存
- ✅ **优雅关闭** - 退出时正确清理
- ✅ **JSONL 格式** - 人类可读的存储格式
- ✅ **错误恢复** - 更好的错误处理和日志记录
- ✅ **内存效率** - 基于 Map 的快速查找
- ✅ **数据清理** - 自动清理过期数据