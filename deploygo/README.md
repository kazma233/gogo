# DeployGo

DeployGo 是一个轻量级的 CI/CD 工具，使用容器（Docker/Podman）进行构建，并通过 SFTP 部署到远程服务器。

## 功能特性

| 命令 | 说明 |
|------|------|
| `deploygo pipeline` | 执行完整的构建+部署流程（包含 write + build + deploy） |
| `deploygo build` | 使用容器执行构建任务 |
| `deploygo deploy` | 部署应用到远程服务器 |
| `deploygo write` | 将 overlays 目录文件复制到 resource 目录 |
| `deploygo list` | 列出所有项目 |

## 安装方式

```bash
# 编译安装
git clone https://github.com/yourname/deploygo
cd deploygo
go build -o deploygo .

# 或直接使用 go run
go run main.go -P myproject pipeline
```

## 快速开始

### 1. 创建项目目录

```
workspace/
└── myproject/
    ├── config.yaml        # 配置文件
    ├── overlays/          # 可选：覆盖文件目录
    │   └── Dockerfile
    ├── resource/          # 可选：资源基础目录
    │   ├── Dockerfile
    │   └── config/
    └── src/               # 源代码目录
```

### 2. 编写配置文件

详见下方 [配置文件详解](#配置文件详解)

### 3. 执行部署

```bash
# 执行完整流水线（write + 构建 + 部署）
deploygo -P myproject pipeline

# 仅构建
deploygo -P myproject build

# 仅部署
deploygo -P myproject deploy

# 仅执行 write（overlays -> resource）
deploygo -P myproject write

# 执行指定的部署步骤
deploygo -P myproject deploy -s stop-server
```

## 配置文件详解

### 项目结构

配置文件位于 `workspace/<project>/config.yaml`，支持以下配置项：

```yaml
# ========== 容器配置 ==========
container:
  type: docker          # 容器类型: docker / podman

# ========== 构建配置 ==========
builds:
  - name: build         # 构建步骤名称（唯一标识）
    image: golang:1.21  # 使用的容器镜像
    working_dir: /app   # 容器内工作目录
    # 环境变量（支持 $VAR 展开）
    environment:
      - GOOS=linux
      - GOARCH=amd64
    # 将本地文件拷贝到容器内
    # from: 相对于 config.yaml 所在目录的本地路径，支持 glob 模式
    # to:   容器内的目标路径
    copy_to_container:
      - from: src/           # config.yaml 所在目录/src/ 目录
        to: /app/src
      - from: "*.go"        # config.yaml 所在目录下所有 .go 文件
        to: /app
      - from: config.json
        to: /app/config.json
    # 容器内命令（所有命令在 working_dir 下执行）
    commands:
      - go build -o bin/app
    # 将容器内文件拷贝回本地
    # from: 容器内的源路径，支持 glob 模式
    # to:   相对于 config.yaml 所在目录的本地目标目录
    copy_to_local:
      - from: bin/app        # 容器内的编译产物
        to: output/         # 拷贝到 config.yaml 所在目录/output/
      - from: "*.log"       # 容器内所有 .log 文件
        to: logs/

# ========== 服务器配置 ==========
servers:
  production:             # 服务器别名（部署时引用）
    host: 192.168.1.100   # 服务器地址
    user: deploy          # SSH 用户名
    port: 22              # SSH 端口
    key_path: ~/.ssh/id_rsa  # SSH 私钥路径（支持 ~ 展开）
    # password: xxx       # 可选：密码认证

# ========== 部署配置 ==========
deploys:
  - name: deploy-app
    server: production    # 引用 servers 中定义的服务器
    # 方式一：执行远程命令（直接通过 SSH 执行）
    commands:
      - cd /opt/myapp && docker compose down
      - cd /opt/myapp && docker compose up -d --no-deps

  - name: upload-files
    server: production
    # 方式二：传输文件
    # from: 相对于 config.yaml 所在目录的本地路径，支持 glob 模式
    # to:   远程服务器的绝对路径
    from: output/         # 传输 config.yaml 所在目录/output/ 目录下的内容
    to: /opt/myapp/
```

### 路径规则说明

| 配置项 | from/to | 路径基准 | 支持绝对路径 | 支持 glob |
|--------|---------|----------|--------------|-----------|
| `copy_to_container` | from | config.yaml 所在目录 | 否 | 是 |
| `copy_to_container` | to | 容器内 | 是（绝对路径） | 否 |
| `copy_to_local` | from | 容器内 | 是 | 是 |
| `copy_to_local` | to | config.yaml 所在目录 | 否 | 否 |
| `deploys` | from | config.yaml 所在目录 | 否 | 是 |
| `deploys` | to | 远程服务器 | 是（绝对路径） | 否 |

**术语说明**：
- `config.yaml 所在目录`: 即 `workspace/<project>/` 目录，所有相对路径都以此为基准
- `projectName`: 配置文件中的 `project.project_name`

### copy_to_local 拷贝结果示例

假设 `config.yaml 所在目录` 为 `workspace/myproject/`，配置如下：

```yaml
copy_to_local:
  - from: bin/app
    to: output/
```

实际效果：
- 容器内 `bin/app` 文件
- 拷贝到本地 `workspace/myproject/output/app`

假设容器内目录结构：
```
/app/
├── bin/
│   └── app
└── logs/
    ├── app.log
    └── error.log
```

配置：
```yaml
copy_to_local:
  - from: bin/
    to: output/
  - from: "*.log"
    to: logs/
```

结果：
```
workspace/myproject/
├── output/
│   └── bin/
│       └── app          # 保持目录结构
└── logs/
    ├── app.log          # *.log 匹配的文件
    └── error.log
```

### deploys 传输示例

假设 `config.yaml 所在目录` 为 `workspace/myproject/`，目录结构：
```
workspace/myproject/
├── output/
│   ├── app
│   └── config.yaml
└── config.yaml
```

**示例 1: 传输整个目录**
```yaml
deploys:
  - name: upload
    from: output/      # 传输 output 目录下的所有内容
    to: /opt/myapp/
```
结果：远程 `/opt/myapp/` 下有 `app` 和 `config.yaml`

**示例 2: 传输单个文件**
```yaml
deploys:
  - name: upload-config
    from: config.yaml
    to: /opt/myapp/config.yaml
```
结果：远程 `/opt/myapp/config.yaml`

**示例 3: 使用 glob 模式**
```yaml
deploys:
  - name: upload-binaries
    from: "output/*"
    to: /opt/myapp/bin/
```
结果：远程 `/opt/myapp/bin/` 下有 `app` 和 `config.yaml`

**注意**：`from` 不支持绝对路径，始终相对于 `config.yaml 所在目录`。`to` 必须是远程服务器的绝对路径。

## 命令详解

### deploygo build

使用容器执行构建任务。

```bash
# 构建所有阶段
deploygo -P myproject build

# 构建指定阶段
deploygo -P myproject build -s build
```

### deploygo deploy

部署应用到服务器。

```bash
# 部署所有步骤
deploygo -P myproject deploy

# 部署指定步骤
deploygo -P myproject deploy -s deploy-app
```

### deploygo pipeline

执行完整的构建和部署流程。

```bash
deploygo -P myproject pipeline
```

### deploygo pipeline

执行完整的构建和部署流程。

```bash
deploygo -P myproject pipeline
```

执行顺序：
1. 执行 `write` 复制 overlays 到 resource
2. 执行所有 build 阶段
3. 执行所有 deploy 步骤

### deploygo write

将 overlays 目录中的文件按照约定结构复制到 resource 目录。

```bash
deploygo -P myproject write
```

**约定优于配置的文件复制方式**：

```
workspace/myproject/
├── overlays/           # 用户覆盖文件目录
│   ├── Dockerfile
│   └── config/
│       └── app.yaml
└── resource/           # 资源基础目录（构建时的工作目录）
    ├── Dockerfile      # 会被 overlays/Dockerfile 覆盖
    └── config/
        └── app.yaml    # 会被 overlays/config/app.yaml 覆盖
```

**执行效果**：
- `overlays/Dockerfile` → `resource/Dockerfile`
- `overlays/config/app.yaml` → `resource/config/app.yaml`

此命令用于在执行 pipeline 前，将用户的定制文件覆盖到资源基础目录，保持 `resource/` 目录的纯净性。

### deploygo list

列出所有项目。

```bash
deploygo list
```

## 注意事项

1. **配置文件位置**：
   - 主配置：`workspace/<project>/config.yaml`
   - 部署配置：`deploy_conf/<project>.yaml`（备用）

2. **overlays 目录约定**：
   - `overlays/` 目录下的文件会覆盖到 `resource/` 目录
   - 执行 `pipeline` 前会自动调用 `write`
   - 目录结构保持一致：`overlays/xxx` → `resource/xxx`

3. **路径处理**：
   - `~` 会自动展开为用户 home 目录（如 `~/.ssh/id_rsa`）
   - 所有本地路径相对于 `config.yaml 所在目录`（`workspace/<project>/`）
   - 远程路径必须是绝对路径

4. **传输方式**：
   - sftp：使用 SFTP 协议传输文件

5. **环境变量**：
   - 配置文件中支持 `$VAR` 和 `${VAR}` 格式的环境变量展开

6. **Windows 注意事项**：
   - 本地开发可在 Windows 上运行
   - 远程服务器必须是 Linux

## 实现原理

### 构建流程

```
1. 拉取容器镜像
2. 创建并启动容器
3. 拷贝本地文件到容器（copy_to_container）
4. 在容器内执行命令（commands）
5. 拷贝容器文件到本地（copy_to_local）
6. 销毁容器
```

### 部署流程

**命令执行模式**：
```
本地 -> SSH -> 远程服务器执行 shell 命令
```

**文件传输模式**：
```
本地 -> SFTP -> 远程服务器
```

### 项目结构

```
workspace/<project>/           # projectDir = 配置文件所在目录
├── config.yaml               # 项目配置
├── overlays/                 # 覆盖文件（write 命令来源）
│   └── Dockerfile
├── resource/                 # 资源基础目录（构建工作目录）
│   ├── Dockerfile
│   └── config/
└── src/                      # 源代码目录

deploy_conf/<project>.yaml    # 备用配置文件位置
```

### 路径术语说明

| 术语 | 定义 | 说明 |
|------|------|------|
| `projectDir` | `workspace/<project>/` | 配置文件所在目录，from/to 路径的基准 |
| `basicPath` | 同 projectDir | 代码中使用的项目基础路径 |
| `overlays/` | `projectDir/overlays/` | 用户覆盖文件目录 |
| `resource/` | `projectDir/resource/` | 资源基础目录（pipeline 执行时的实际工作目录） |

### Overlays 机制

```
执行 pipeline 时的文件复制流程：

1. write 阶段：将 overlays/ 覆盖到 resource/
2. 构建阶段：使用 resource/ 作为项目基础目录
3. 所有 copy_to_local 路径都相对于 resource/

示例：
config.yaml 所在目录 = workspace/myproject/
overlays/Dockerfile        →  resource/Dockerfile
overlays/config/app.yaml   →  resource/config/app.yaml
overlays/bin/start.sh      →  resource/bin/start.sh
```

## License

MIT
