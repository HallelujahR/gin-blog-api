## 访问日志分析模块（新架构）

### 1. 总体目标
构建“访问请求 → Gin 中间件 → 日志/Redis → 周期 ETL → MySQL 快照 → 可视化”的全链路数据采集方案，实现：
- 近 30 日 UV、地区分布与热门文章统计。
- 请求链路全异步，避免阻塞主业务。
- Redis 仅作实时缓存，MySQL 持久化保证可回溯。
- 兼容旧版 `/api/stats` 返回字段。

### 2. 组件分层
| 层级 | 位置 | 说明 |
| --- | --- | --- |
| 接入层 | `middleware.AnalyticsLogger` | 捕获请求上下文，生成 `analytics.Event`。 |
| 流水线 | `analytics.Capture` + `consumeEvents` | 负责事件缓冲、写 `_raw.log` 与写 Redis。 |
| 数据缓存 | Redis（`analytics:*`） | UV/地区/请求量短期缓存，TTL 45 天。 |
| ETL | `analytics.StartETLWorker` | 每 5 分钟聚合近 30 日数据，写入 `traffic_snapshots`。 |
| 聚合 | `stats.BuildVisitSummary` + `stats.LoadTrafficSnapshot` | 组合文章热度与地区统计。 |
| API | `stats.Handler` | 先读 Redis 缓存，失效后回源 DB 并刷新缓存。 |

### 3. 数据流详解
1. **请求采集**  
   - 中间件记录 IP/方法/路径/状态码/耗时/User-Agent，打上发生时间。
2. **事件入队**  
   - `eventCh` 容量 2048，避免突发写满主线程。溢出时打印 `[analytics] event channel full` 并放弃该事件。
3. **后台消费**  
   - `writeRaw`：懒加载 `DailyLogWriter`，以 `YYYY-MM-DD_raw.log` 命名，便于日常审计与灾备。  
   - `cacheEvent`：Redis Pipeline 一次写入三类 Key：  
     - `analytics:uv:YYYYMMDD`（Set）：去重独立 IP。  
     - `analytics:req:YYYYMMDD`（String）：累计请求量。  
     - `analytics:region:YYYYMMDD`（Hash）：仅针对 `service.IsChinaIP` 判断为国内的 IP，调用 `service.LookupRegion` 获取省份，过滤 `LOCAL/UNKNOWN` 与所有非中国 IP。  
     - 所有键统一设置 `redisKeyTTL = 45 天`。
4. **ETL 周期任务**  
   - `analytics.StartETLWorker` 在 `initializer.Run` 中启动。  
   - 每 5 分钟 `flushSnapshot`：  
     - 连续回溯 30 天的 `UV Set` 做 union，得到窗口级独立访客数。  
     - 汇总每日省份计数 Hash，组成 `map[string]int`。  
     - 构造 `dao.TrafficSnapshotDTO` 并 `SaveTrafficSnapshot`，由 `traffic_snapshots` 表持久化。  
     - 若 Redis 不可用，记录错误等待下次周期，无需阻塞主线程。
5. **统计查询**  
   - `stats.Handler` 首先尝试 `GetStatsCache`（Redis 24h 缓存）。  
   - 缓存缺失：  
     - `LoadTrafficSnapshot(30)` → `SnapshotData`（若无记录返回空 map）。  
     - `BuildVisitSummary(30, 3)` → 读取 `post_view_stats` & `posts`，仅使用数值 ID（不再依赖 slug），按照创建时间倒序补全标题。  
     - `Aggregate` → 组装 `StatsResult` 并写回缓存。  
   - 返回结果已去除平均停留时长、地区 count 字段，仅保留百分比，且过滤 `UNKNOWN/LOCAL/非中国` 省份。

### 4. Redis Key 协议
| Key | 结构 | 说明 |
| --- | --- | --- |
| `analytics:uv:YYYYMMDD` | Set | 当日 UV，ETL 时做 union。 |
| `analytics:req:YYYYMMDD` | String | 当日请求量（扩展字段，当前未输出）。 |
| `analytics:region:YYYYMMDD` | Hash | 省份→访问量，仅统计中国 IP。 |
| `stats:cache` | String | `/api/stats` 最终响应缓存。 |

### 5. 数据库表 `traffic_snapshots`
| 字段 | 类型 | 描述 |
| --- | --- | --- |
| `id` | bigint | 自增主键。 |
| `lookback_days` | int | 聚合窗口大小（默认 30），建唯一索引便于 upsert。 |
| `unique_visitors` | int | 窗口内 UV。 |
| `region_json` | longtext | 省份访问量 `map[string]int` 序列化。 |
| `generated_at` | datetime | 快照生成时间，供可视化显示。 |
| `created_at/updated_at` | datetime | GORM 维护。 |

### 6. 灾备与历史数据
1. **原始日志**：`logs/YYYY-MM-DD_raw.log` 永久保存（可配合对象存储归档），即便 Redis 数据丢失也可重放恢复。  
2. **历史转换工具**：`cmd/tools/export_legacy_logs` 将旧格式 `*_log` 转换为 `_raw.log`，保证 ETL 可读取。  
3. **回放工具**：`cmd/tools/replay_raw_logs` 支持按照日期范围扫描 `_raw.log` 并重新写入 Redis Key，流程如下：  
   - `go run ./cmd/tools/replay_raw_logs -log-dir=./logs -start=2025-11-01 -end=2025-11-19`。  
   - 逐行解析 `timestamp|ip|method|path|status|latency|ua`，复用 `service.IsChinaIP` 与 `LookupRegion`，与在线 `cacheEvent` 行为保持一致。  
   - 回放结束后等待下一次 ETL 或手动重启服务即可刷新快照。  
4. **恢复顺序建议**：`export_legacy_logs`（如需） → `replay_raw_logs` → 等待 ETL/触发 `/api/stats` 回源。

### 7. 并发与性能设计
- **通道隔离**：请求线程仅负责写 channel，O(1) 开销；后台消费者串行处理，避免多线程争用文件句柄。  
- **文件写入**：`DailyLogWriter` 使用互斥锁和按日轮转，创建文件时先判断是否存在，避免覆盖旧日志。  
- **Redis Pipeline**：每次事件写入仅一次 RTT，设置 500ms 超时，异常打印日志但不影响主流程。  
- **ETL**：使用 2s 超时批量读取 Redis，`defaultLookbackDays=30` 保证最多 30 轮操作；出错时打印 `[analytics] flush snapshot error`，待下次重试。  
- **Stats 缓存**：24 小时 TTL，命中率高时几乎零数据库压力；Redis 不可用时自动回源 DB 并 `fmt.Printf` 告警。

### 8. 与旧版能力对比
- **来源**：不再解析 Gin 访问日志（`[GIN] ...`），改为自定义原始格式，字段齐全且易回放。  
- **地区统计**：统一使用 `GeoLite2-City.mmdb`，省级精度，过滤本地回环与未知地区。  
- **热门文章**：完全依赖 MySQL `post_view_stats` 与 `posts`，不再依靠 slug。  
- **字段裁剪**：移除平均停留时间与地区 count 字段，满足最新业务要求。  
- **部署**：去除 Docker，配合 `scripts/deploy.sh` 裸机部署 + systemd，Redis/MySQL/Nginx 均使用宿主机实例。

### 9. 运维操作速览
| 操作 | 命令 | 说明 |
| --- | --- | --- |
| 转换旧日志 | `go run ./cmd/tools/export_legacy_logs -log-dir=./logs -overwrite=false` | 将历史 `*_log` 转成 `_raw.log`。 |
| 回放原始数据 | `go run ./cmd/tools/replay_raw_logs -log-dir=./logs -start=YYYY-MM-DD -end=YYYY-MM-DD` | Redis 数据损坏时恢复。 |
| 手工触发统计 | 请求一次 `/api/stats` 即可，如果缓存缺失会自动回源→写缓存。 |
| 检查 ETL | 关注日志 `[analytics] flush snapshot error`，或直接查询 `traffic_snapshots`。 |

### 10. 目录索引
- `middleware/analytics_logger.go`：中间件入口，含详细注释。  
- `analytics/pipeline.go`：事件通道、写日志 + Redis 缓存。  
- `analytics/etl.go`：定时任务，将 Redis 数据固化到 MySQL。  
- `dao/traffic_snapshot_dao.go` & `models/traffic_snapshot.go`：快照表定义与 Upsert。  
- `stats/*.go`：与 UI 对接的聚合逻辑。  
- `cmd/tools/export_legacy_logs` / `cmd/tools/replay_raw_logs`：历史数据导入/重放工具。  
- `service/log_init.go`：每日日志写入器，实现 `_log` 与 `_raw.log` 共存。  

以上即新一代访问日志分析模块的完整设计文档，可放置在 `docs/analytics-module.md` 供团队查阅。

