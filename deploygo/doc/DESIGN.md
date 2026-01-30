# DeployGo CI/CD 工具设计文档

## 1. 项目概述

DeployGo 是一个基于 Go 语言和容器技术的轻量级 CI/CD 工具，专注于简单、高效的项目构建和部署。

### 主要特性

- **容器化构建** - 使用 Docker 或 Podman 容器进行构建，确保环境一致性
- **多运行时支持** - 原生支持 Docker 和 Podman，可通过配置切换
- **多阶段构建** - 支持多个构建阶段按顺序执行
- **灵活的文件拷贝** - 支持 glob pattern 匹配本地文件到容器的拷贝
- **多种部署方式** - 支持 SSH 命令执行和 SFTP 文件传输
- **Git 克隆支持** - 支持从 Git 仓库克隆代码到 source 目录
- **YAML 配置** - 使用 YAML 格式的配置文件

### 设计原则

- **约定大于配置** - 使用固定目录结构，减少配置复杂度
- **简单可靠** - MVP 优先，保持简单可靠
- **易于使用** - 简洁的命令行接口和配置格式

## 2. 项目结构

```
deploygo/
├── cmd/
│   ├── main.go           # 程序入口
│   ├── root.go           # 根命令定义
│   ├── build.go          # 构建命令实现
│   ├── clone.go          # Git 克隆命令实现
│   ├── deploy.go         # 部署命令实现
│   ├── pipeline.go       # 构建+部署流水线命令
│   ├── write.go          # 写入文件命令（overlays -> source）
│   └── list.go           # 列出所有项目
├── internal/
│   ├── config/
│   │   └── parser.go       # YAML 配置解析器
│   ├── container/
│   │   ├── runtime.go      # 容器运行时接口
│   │   ├── manager.go      # 运行时工厂
│   │   ├── docker.go       # Docker 运行时实现
│   │   └── podman.go       # Podman 运行时实现
│   ├── deploy/
│   │   └── executor.go     # SSH 和文件传输执行器
│   ├── git/
│   │   └── clone.go        # Git 克隆功能
│   └── stage/
│       ├── builds.go       # 构建阶段执行
│       ├── deploys.go      # 部署步骤执行
│       └── file.go         # 文件写入工具
├── workspace/             # 项目工作目录
│   ├── myapp/
│   │   ├── config.yaml     # 项目配置
    │   │   ├── source/         # 源代码目录（构建时 copy 到容器）
    │   │   └── overlays/       # 配置文件目录（会覆盖到 source）
│   └── playground/
│       ├── config.yaml
    │       ├── source/
    │       └── overlays/
├── doc/
│   └── DESIGN.md         # 设计文档
├── go.mod
└── README.md
```

## 3. 配置文件结构

### 3.1 项目目录

所有项目存放在 `workspace/` 目录下，每个子目录代表一个项目。项目名 = 目录名。

项目目录结构：
```
workspace/<project>/
├── config.yaml     # 项目配置
├── source/         # 源代码（构建时 copy 到容器）
└── overlays/       # 配置文件（write 时 copy 到 source，不带顶层目录）
```

### 3.2 完整配置示例

```yaml
container:
  type: podman

# Git 克隆配置（可选）
clone:
  url: https://github.com/example/myapp.git
  branch: main  # 可选，默认 master

servers:
  web-server:
    host: deploy@example.com
    user: deploy
    port: 22
    key_path: ~/.ssh/id_rsa

builds:
  - name: build
    image: golang:1.21-alpine
    working_dir: /app
    environment:
      - CGO_ENABLED=0
      - GOOS=linux
    copy_to_container:
      - from: ./source/
        to: /app/
    copy_to_local:
      - from: /app/output/
        to: ./output/
    commands:
      - go build -o /app/output/server main.go

deploys:
  - name: upload-files
    server: web-server
    from: ./output/
    to: /var/www/myapp/

  - name: restart-service
    server: web-server
    commands:
      - cd /var/www/myapp && systemctl restart myapp
```

### 3.3 配置项说明

| 配置项 | 类型 | 说明 |
|--------|------|------|
| `container.type` | string | 容器运行时类型，支持 `docker`、`podman` |
| `servers.*.host` | string | 服务器地址 |
| `servers.*.user` | string | 用户名 |
| `servers.*.port` | int | 端口号，默认 22 |
| `servers.*.key_path` | string | SSH 私钥路径 |
| `servers.*.password` | string | SSH 密码（可选，与 key_path 二选一） |
| `builds.*.name` | string | 构建名称，唯一标识 |
| `builds.*.image` | string | 容器镜像名称 |
| `builds.*.working_dir` | string | 容器内工作目录 |
| `builds.*.environment` | []string | 环境变量列表 |
| `builds.*.copy_to_container` | []CopyPath | 本地文件拷贝到容器 |
| `builds.*.copy_to_local` | []CopyPath | 容器文件拷贝到本地 |
| `builds.*.commands` | []string | 在容器内执行的命令列表 |
| `deploys.*.name` | string | 部署步骤名称 |
| `deploys.*.server` | string | 引用的服务器配置名称 |
| `deploys.*.commands` | []string | 在远程服务器执行的命令列表 |
| `deploys.*.from` | string | 源路径（本地，相对于项目目录） |
| `deploys.*.to` | string | 目标路径（远程服务器） |
| `clone.url` | string | Git 仓库地址 |
| `clone.branch` | string | Git 分支名称，默认 master |

## 4. 使用方法

### 4.1 列出所有项目

```bash
deploygo list
```

输出示例：
```
Available projects:

  - myapp
  - playground
```

### 4.2 克隆 Git 仓库

从配置的 Git 仓库克隆代码到 `source/` 目录，**会清空原有内容**。

```bash
# 克隆 Git 仓库到 source 目录
deploygo -P myapp clone
```

要求：
- 系统中必须安装 `git` 命令
- 需要在 `config.yaml` 中配置 `clone.url`
- 支持 SSH 和 HTTPS 协议的仓库地址

### 4.3 写入文件（overlays -> source）

将 `overlays/` 目录下的配置文件 copy 到 `source/` 目录，保留目录结构但不包含 `overlays/` 本身。

```bash
# 执行所有写入步骤
deploygo -P myapp write
```

### 4.4 构建项目

```bash
# 构建指定项目的所有步骤
deploygo -P myapp build

# 只构建特定步骤
deploygo -P myapp build -s docker-build
```

### 4.5 部署项目

```bash
# 部署指定项目的所有步骤
deploygo -P myapp deploy

# 只执行特定部署步骤
deploygo -P myapp deploy -s restart-service
```

### 4.6 运行完整流水线

```bash
# 克隆（如果配置了）→ 写入 → 构建 + 部署
deploygo -P myapp pipeline
```

Pipeline 执行顺序：
1. **Git 克隆**（如果配置了 `clone.url`）- 从 Git 仓库克隆代码到 source 目录
2. **写入 Overlays**（如果存在 `overlays/` 目录）- 将配置文件覆盖到 source 目录
3. **构建** - 执行所有构建阶段
4. **部署** - 执行所有部署步骤

### 4.7 文件传输说明

### 4.8 文件传输说明

**文件上传**：
```
from: ./docker-compose.yml
to:   ./playground/
结果: ./playground/docker-compose.yml
```

**目录上传**：
```
from: ./tmp/
to:   ./playground
结果: ./playground/a/b/c.txt  (不带 tmp/ 目录)
```

## 5. Glob Pattern 支持

### 5.1 支持的 Pattern 语法

| 模式 | 含义 | 示例 |
|------|------|------|
| `*` | 匹配任意字符（不含目录分隔符） | `*.go` 匹配所有 go 文件 |
| `**` | 匹配任意路径（含多级目录） | `**/*.log` 匹配所有 log 文件 |
| `?` | 匹配单个字符 | `file-?.go` 匹配 file-1.go, file-2.go |
| `[abc]` | 匹配括号内任意字符 | `file-[abc].go` 匹配 file-a.go, file-b.go, file-c.go |

### 5.2 排除文件

```yaml
copy_to_container:
  - from: ./src/**/*.go
    to: /app/src/
    exclude:
      - "**/*_test.go"
      - "**/vendor/**"
```

## 6. 常见问题

### Q: 配置文件修改后需要重新加载吗？
A: 不需要，程序每次运行时会重新读取配置文件。

### Q: 如何同时部署多个项目？
A: 为每个项目创建独立的目录，然后分别运行部署命令。

### Q: overlays 目录可以为空吗？
A: 可以，如果没有配置文件需要覆盖，overlays 目录可以不存在或为空。

### Q: 为什么需要 overlays？
A: 配置文件通常包含敏感信息（如数据库密码、API Key），不适合提交到代码仓库。使用 overlays 目录可以将配置文件与源代码分离：源代码放在 source/，配置文件放在 overlays/，write 时自动合并。
