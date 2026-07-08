# 智能外卖柜后端 API 文档

## 概述

本项目为智能外卖柜后端服务，基于 Golang + MySQL 开发。系统管理 **单个柜子** 的状态，提供"存"和"取"两个核心功能。后端与单片机（MCU）通过 TCP 长连接保持实时通信。

---

## 系统架构

```
┌──────────┐   HTTP    ┌──────────────┐   TCP    ┌──────────┐
│  前端页面  │ ◄──────► │  Golang 后端  │ ◄──────► │  单片机   │
│ (存/取)   │          │   (本服务)    │          │  (MCU)   │
└──────────┘          └──────┬───────┘          └──────────┘
                             │
                             ▼
                        ┌──────────┐
                        │  MySQL   │
                        └──────────┘
```

- 前端与后端通过 HTTP REST API 通信
- 后端与单片机通过 TCP 长连接通信
- 后端与 MySQL 数据库交互存外卖取码

---

## 柜子状态机

```
                    ┌──────────────────────────┐
                    │                          │
    ┌─────── 存请求 ─┴──►  WAITING_DOOR_CLOSE   │
    │               │    (等待关门)              │
    │               └──────────┬───────────────┘
    │                          │ MCU上报关门(0)
    │                          ▼
    │               ┌──────────────────┐
    │               │    OCCUPIED      │
    │               │  (已存物, 门关)   │
    │               └────────┬─────────┘
    │                        │ 取请求验证通过
    │                        ▼
    │               ┌──────────────────────────┐
    │               │  WAITING_DOOR_CLOSE       │
    │               │  (等待取物关门)            │
    │               └──────────┬───────────────┘
    │                          │ MCU上报关门(0)
    │                          ▼
    └──────────────────┐ IDLE (空闲) ◄──────────┘
```

---

## API 接口

### 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json`

---

### 1. 存入外卖

**POST** `/api/v1/cabinet/store`

点击"存"按钮后，前端携带 4 位数字密码发起请求。后端记录密码，通知单片机开门。

#### 请求体

```json
{
  "code": "1234"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 4位数字密码，如 "1234" |

#### 成功响应 (200)

```json
{
  "code": 0,
  "message": "柜门已打开，请放入外卖后关门"
}
```

#### 错误响应

```json
{
  "code": 40001,
  "message": "密码必须为4位数字"
}
```

```json
{
  "code": 40002,
  "message": "柜子当前不可用，请稍后再试"
}
```

```json
{
  "code": 50001,
  "message": "与单片机通信失败"
}
```

---

### 2. 取出外卖

**POST** `/api/v1/cabinet/retrieve`

点击"取"按钮后，前端携带 4 位数字密码发起请求。后端校验密码，匹配则通知单片机开门。

#### 请求体

```json
{
  "code": "1234"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 4位数字密码，如 "1234" |

#### 成功响应 (200) — 验证通过

```json
{
  "code": 0,
  "message": "验证通过，柜门已打开，请取走外卖后关门"
}
```

#### 错误响应 — 密码不匹配

```json
{
  "code": 40003,
  "message": "密码错误，请重试"
}
```

#### 错误响应 — 柜子为空

```json
{
  "code": 40004,
  "message": "柜子内无外卖可取"
}
```

---

### 3. 查询柜子状态

**GET** `/api/v1/cabinet/status`

查询当前柜子的状态。

#### 成功响应 (200)

```json
{
  "code": 0,
  "data": {
    "status": "idle",
    "status_text": "空闲",
    "has_item": false,
    "door_open": false
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| status | string | idle / waiting_close / occupied |
| status_text | string | 状态中文说明 |
| has_item | bool | 柜内是否有外卖 |
| door_open | bool | 柜门是否打开 |

---

## 单片机通信协议 (TCP)

### 连接方式

- 后端作为 **TCP Server**，监听 `0.0.0.0:9090`
- 单片机作为 **TCP Client**，主动连接后端
- 保持 **长连接**，断开后单片机会自动重连

### 协议格式

报文为纯文本，以换行符 `\n` 分隔。

#### 后端 → 单片机（开门指令）

```
1\n
```

收到 `1` 后，单片机驱动舵机/电磁锁开门。

#### 单片机 → 后端（关门上报）

```
0\n
```

单片机检测到柜门关闭后，上报 `0`。

### 心跳机制

单片机每 **5 秒** 发送一次心跳：

```
HEARTBEAT\n
```

后端回复：

```
PONG\n
```

超过 **15 秒** 未收到心跳则认为连接断开，后端标记 MCU 离线。

---

## 数据库表结构

### cabinet_records 表

```sql
CREATE TABLE cabinet_records (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    code        VARCHAR(4)   NOT NULL COMMENT '4位存取密码',
    action      ENUM('store','retrieve') NOT NULL COMMENT '操作类型',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='存取记录表';
```

### current_item 表（当前柜内外卖）

```sql
CREATE TABLE current_item (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    code        VARCHAR(4)   NOT NULL COMMENT '当前外卖的取件密码',
    stored_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '存入时间',
    status      ENUM('stored','retrieved') NOT NULL DEFAULT 'stored' COMMENT '状态'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='当前柜内外卖';
```

---

## 错误码汇总

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 40001 | 密码格式不正确（必须为4位数字） |
| 40002 | 柜子当前不可用（门开着或正在操作中） |
| 40003 | 密码错误 |
| 40004 | 柜子为空，无外卖可取 |
| 50001 | 与单片机通信失败 |
| 50002 | 数据库操作失败 |
| 50003 | 单片机不在线 |
