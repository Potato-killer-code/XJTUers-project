# 外卖柜

基于 **Golang + Vue 3 + MySQL** 的智能外卖柜系统，管理单个柜子的存入/取出操作。后端通过 TCP 长连接与单片机（MCU）实时通信，控制柜门开关。

---

## 系统架构

```
┌──────────┐   HTTP    ┌──────────────┐   TCP    ┌──────────┐
│  Vue 前端  │ ◄──────► │  Golang 后端  │ ◄──────► │  单片机   │
│ (存/取)   │          │   (本服务)    │          │  (MCU)   │
└──────────┘          └──────┬───────┘          └──────────┘
                             │
                             ▼
                        ┌──────────┐
                        │  MySQL   │
                        └──────────┘
```

### 柜子状态机

```
IDLE(空闲) ──存请求──► WAITING_CLOSE ──MCU报0──► OCCUPIED(已存物)
   ▲                                                  │
   └──────── MCU报0 ──── WAITING_CLOSE ◄──取请求──────┘
```

---

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.21 + 标准库 `net/http` |
| 数据库 | MySQL 8.0（自动建表） |
| 前端 | Vue 3 + Vue Router + Vite |
| MCU 通信 | TCP 长连接（纯文本协议） |
| 配置 | YAML + `.env` 环境变量 |

---

## 快速开始

### 1. 环境要求

- Go 1.21+
- Node.js 18+
- MySQL 8.0+

### 2. 配置数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS smart_cabinet DEFAULT CHARSET utf8mb4;"
```

### 3. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env，填入你的 MySQL 密码
```

### 4. 启动服务

**开发模式（前后端分离，推荐）：**

```bash
# 终端 1：启动后端 (http://localhost:8080)
go run main.go

# 终端 2：启动前端 (http://localhost:3000，自动代理 API)
cd web && npm install && npm run dev
```

**生产模式（一体部署）：**

```bash
# 构建前端
cd web && npm install && npm run build && cd ..

# 编译并启动
go build -o smart-cabinet main.go
./smart-cabinet
# 访问 http://localhost:8080 即可使用
```

---

## 项目结构

```
.
├── main.go                         # Go 入口
├── config.yaml                     # 服务配置（非敏感）
├── .env.example                    # 环境变量模板
├── api/
│   └── api.md                      # API 文档
├── internal/
│   ├── config/config.go            # 配置加载（.env + yaml）
│   ├── db/db.go                    # MySQL 操作 & 自动建表
│   ├── handler/handler.go          # HTTP 接口处理器
│   ├── mcu/mcu.go                  # MCU TCP 通信服务
│   ├── model/model.go              # 数据模型
│   ├── service/service.go          # 业务逻辑层
│   └── state/state.go              # 柜子状态机（线程安全）
└── web/                            # Vue 3 前端
    ├── src/
    │   ├── views/
    │   │   ├── StoreView.vue       # 存入页面
    │   │   └── RetrieveView.vue    # 取出页面
    │   ├── api/index.js            # API 请求封装
    │   ├── router/index.js         # 前端路由
    │   ├── App.vue                 # 根组件
    │   └── main.js                 # 入口
    ├── index.html
    └── vite.config.js              # Vite 配置 + API 代理
```

---

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/cabinet/store` | 存入外卖（携带 4 位密码） |
| POST | `/api/v1/cabinet/retrieve` | 取出外卖（校验密码） |
| GET | `/api/v1/cabinet/status` | 查询柜子状态 |

详细文档见 [api/api.md](api/api.md)

---

## MCU 通信协议

后端作为 TCP Server 监听 `:9090`，单片机主动连接。

| 方向 | 报文 | 说明 |
|------|------|------|
| 后端 → MCU | `1\n` | 开门指令 |
| MCU → 后端 | `0\n` | 关门上报 |
| MCU → 后端 | `HEARTBEAT\n` | 心跳（每 5s） |
| 后端 → MCU | `PONG\n` | 心跳回复 |

> 超时 15s 未收到心跳则认为 MCU 离线。

---

## 配置说明

配置优先级：**`.env` 环境变量 > `config.yaml` > 默认值**

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `DB_HOST` | MySQL 地址 | 127.0.0.1 |
| `DB_PORT` | MySQL 端口 | 3306 |
| `DB_USER` | MySQL 用户名 | root |
| `DB_PASSWORD` | MySQL 密码 | （必填） |
| `DB_NAME` | 数据库名 | smart_cabinet |
| `SERVER_HOST` | HTTP 监听地址 | 0.0.0.0 |
| `SERVER_PORT` | HTTP 端口 | 8080 |
| `MCU_HOST` | MCU TCP 监听地址 | 0.0.0.0 |
| `MCU_PORT` | MCU TCP 端口 | 9090 |

---

## 数据库表

系统启动时自动建表，无需手动执行 SQL。

- **`cabinet_records`** — 存取操作记录（code, action, created_at）
- **`current_item`** — 当前柜内外卖（code, stored_at, status）

