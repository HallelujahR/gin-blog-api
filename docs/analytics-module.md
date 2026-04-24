# 访问统计模块

## 目标

链路为“请求 → Gin 中间件 → 原始日志/Redis → 周期 ETL → MySQL 快照 → `/api/stats`”。请求线程只入队，统计失败不影响主业务。

## 组件

| 层级 | 位置 | 说明 |
| --- | --- | --- |
| 采集 | `internal/middleware/analytics_logger.go` | 捕获 IP、路径、状态码、耗时、UA |
| 流水线 | `internal/modules/analytics/pipeline.go` | 调用 `internal/platform/logstore` 和 `redisstore` 写日志与 Redis |
| ETL | `internal/modules/analytics/etl.go` | 定期汇总 Redis，写 `traffic_snapshots` |
| API | `internal/modules/analytics/stats/` | 组合访问快照、文章热度和 Redis 响应缓存 |
| 工具 | `cmd/tools/export_legacy_logs`、`cmd/tools/replay_raw_logs` | 历史日志转换与回放 |

## Redis Key

| Key | 结构 | 说明 |
| --- | --- | --- |
| `analytics:uv:YYYYMMDD` | Set | 当日 UV，ETL 时做 union。 |
| `analytics:req:YYYYMMDD` | String | 当日请求量（扩展字段，当前未输出）。 |
| `analytics:region:YYYYMMDD` | Hash | 省份→访问量，仅统计中国 IP。 |
| `stats:cache` | String | `/api/stats` 最终响应缓存。 |

## 数据表

`traffic_snapshots` 保存窗口级统计快照：

| 字段 | 类型 | 描述 |
| --- | --- | --- |
| `id` | bigint | 自增主键。 |
| `lookback_days` | int | 聚合窗口大小（默认 30），建唯一索引便于 upsert。 |
| `unique_visitors` | int | 窗口内 UV。 |
| `region_json` | longtext | 省份访问量 `map[string]int` 序列化。 |
| `generated_at` | datetime | 快照生成时间，供可视化显示。 |
| `created_at/updated_at` | datetime | GORM 维护。 |

## 运维

| 操作 | 命令 | 说明 |
| --- | --- | --- |
| 转换旧日志 | `go run ./cmd/tools/export_legacy_logs -log-dir=./logs -overwrite=false` | 将历史 `*_log` 转成 `_raw.log`。 |
| 回放原始数据 | `go run ./cmd/tools/replay_raw_logs -log-dir=./logs -start=YYYY-MM-DD -end=YYYY-MM-DD` | Redis 数据损坏时恢复。 |
| 手工触发统计 | 请求一次 `/api/stats` 即可，如果缓存缺失会自动回源→写缓存。 |
| 检查 ETL | 关注日志 `[analytics] flush snapshot error`，或直接查询 `traffic_snapshots`。 |

## 设计注意

- Redis 只是实时缓存，原始日志和 MySQL 快照承担恢复能力。
- 地区统计过滤本地、未知和非中国 IP。
- `/api/stats` 优先读 Redis 缓存，缓存缺失时回源数据库并刷新缓存。
