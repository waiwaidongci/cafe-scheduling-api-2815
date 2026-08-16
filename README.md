# cafe-scheduling-api

咖啡店排班系统 API，使用 Gin + PostgreSQL 实现。角色分为店长（manager）和员工（employee），支持员工、班次类型、营业时间、排班生成/查看/调整、周视图、重叠班次检测和按周工时统计。

## 技术栈

- Go 1.26
- Gin
- PostgreSQL / pgx
- JWT（golang-jwt）
- bcrypt 密码哈希

## 目录结构

```text
cmd/api            HTTP 服务入口
cmd/migrate        数据库迁移命令
internal/config    配置读取
internal/model     员工、班次、营业时间等模型
internal/repository PostgreSQL 数据访问
internal/service   排班校验、冲突检测、工时统计
internal/handler   HTTP 参数处理
internal/middleware JWT 鉴权和角色控制
internal/router    路由分组注册
internal/pkg       响应封装和分页等公共能力
migrations         建表和种子 SQL
```

## 配置

环境变量均有默认值：

```text
DATABASE_URL=postgres://postgres:postgres@127.0.0.1:55432/cafe_scheduling?sslmode=disable
HTTP_PORT=18101
JWT_SECRET=cafe-scheduling-dev-secret
TOKEN_TTL=24h
```

## 运行

```bash
go mod tidy
go test ./...
go run ./cmd/migrate -migrations migrations
go run ./cmd/api
```

种子账号：

```text
username: manager
password: manager123
```

## API

### 登录

```http
POST /api/v1/auth/login
```

请求体：

```json
{
  "username": "manager",
  "password": "manager123"
}
```

后续请求使用 `Authorization: Bearer <token>`。

### 店长接口

```text
POST   /api/v1/employees
GET    /api/v1/employees
POST   /api/v1/shift-types
GET    /api/v1/shift-types
PUT    /api/v1/business-hours
GET    /api/v1/business-hours
POST   /api/v1/schedules/generate
GET    /api/v1/schedules/week?start=2026-08-17
GET    /api/v1/schedules/conflicts?start=2026-08-17
GET    /api/v1/statistics/hours?start=2026-08-17
PUT    /api/v1/schedules/:id
```

### 员工只读接口

```text
GET /api/v1/my/schedules?start=2026-08-17
GET /api/v1/my/statistics/hours?start=2026-08-17
```

`day_of_week` 使用 0=周一，1=周二，...，6=周日。时间统一使用 `HH:MM`。

## 关键行为

- 排班生成以 `week_start` 为基准，自动归一化到周一，默认覆盖所有在职员工。
- 生成时会检查班次是否落在营业时间内，并跳过与已有班次重叠的排班。
- 调整班次在 service 层做同员工同日期重叠校验，仓库层使用事务提交。
- 员工令牌只能访问自己的排班和工时统计。

## 测试

```bash
go test ./...
```

当前测试覆盖周起始归一化、重叠检测、周工时统计、时间解析和营业时间边界判断。
