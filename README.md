# pay-log

`pay-log` 是一个个人账单管理与分析工具，用于导入支付宝、微信账单，按月份或时间区间查看收支统计，并支持生成 AI 消费分析。

## 主要功能

- 登录认证：基于账号密码登录，使用 Redis 保存登录态。
- 账单导入：支持支付宝 CSV、微信 XLSX 账单文件上传解析。
- 手动记账：支持手动新增账单记录。
- 账单管理：按月或时间区间查看账单明细，支持金额排序和删除记录。
- 统计汇总：展示月度收入、支出、支付宝/微信渠道统计、理财金额等数据。
- AI 分析：对指定时间范围内的账单生成消费分析报告。

## 技术栈

- 后端：Go、Gin、GORM、MySQL、Redis
- 前端：React、TypeScript、Vite、Tailwind CSS
- 构建：Makefile、npm

## 效果展示

![overview](./assets/overview.png)

![summary](./assets/summary.png)

## 环境准备

请先准备以下环境：

- Go 1.26.3 或兼容版本
- Node.js 与 npm
- MySQL
- Redis

## 配置说明

项目启动时默认读取根目录下的 `config.yaml`。可以参考 `config.yaml.tmplate` 创建配置文件：

```bash
cp config.yaml.tmplate config.yaml
```

需要根据本地环境修改以下配置：

- `server.port`：后端服务端口，默认示例为 `39975`
- `mysql`：MySQL 地址、端口、用户名、密码、数据库名
- `redis`：Redis 地址、端口、密码、DB
- `logger`：日志目录和日志级别
- `model`：AI 分析使用的模型服务地址、模型 ID、API Key
- `auth`：登录 token 名称、过期时间、续期时间

> 注意：`config.yaml` 中通常包含数据库密码、Redis 密码和 API Key，请不要提交真实配置到公开仓库。

## 启动方式

### 1. 安装前端依赖

```bash
cd web
npm install
```

### 2. 启动后端服务

在项目根目录执行：

```bash
make run
```

等价于：

```bash
go run main.go
```

服务会根据 `config.yaml` 中的 `server.port` 启动，默认示例端口为 `39975`。

### 3. 启动前端开发服务

在 `web` 目录执行：

```bash
npm run dev
```

前端开发服务会通过 Vite 代理 `/api` 到后端 `http://localhost:39975`。

## 构建

### 构建前端

在项目根目录执行：

```bash
make build-web
```

该命令会在 `web` 目录执行 `npm run build`，并将前端产物输出到根目录的 `static` 目录。

### 构建后端

```bash
make build
```

构建产物位于：

```bash
bin/pay-log
```

构建后可以运行：

```bash
./bin/pay-log -f config.yaml
```

## 创建用户

首次使用前需要创建登录用户：

```bash
make add-user USER=your_username PASSWORD=your_password
```

等价于：

```bash
go run cmd/adduser/main.go -u your_username -p your_password -f config.yaml
```

## 常用命令

```bash
make run        # 启动后端服务
make build      # 构建后端二进制文件
make build-web  # 构建前端静态资源
make add-user USER=xxx PASSWORD=xxx  # 创建登录用户
```
