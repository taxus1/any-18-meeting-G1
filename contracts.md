# go-gin-m · 基座机制契约（S1 坑弹药库）

> Go 栈（Gin + GORM + go-redis）。出题人设计 S1「机制契约盲区」埋点时按本清单取；
> S2（跨模块不变量）/ S3（隐性矛盾）/ S4（边界）基座无关，照旧。

## 一、审计操作人（context 透传 + GORM 钩子）
- 登录用户由 `middleware.OperatorAuth`（HTTP Basic）写入**请求 context**（`model.WithOperator`）；
  仓储必须 `db.WithContext(ctx)` 把 ctx 透传给 GORM，`BeforeCreate/BeforeUpdate` 钩子从
  `tx.Statement.Context` 读操作人填 `create_by/update_by`。
- **来源是登录态，不是请求参数/请求体**。验收断言「新增/修改后 create_by 为登录用户且不可伪造」。
- 掉坑写法：`db` 直接调用（漏 `WithContext(ctx)`）→ 钩子读到 `system`；从 query/body 取 `createBy`；
  仓储方法签名丢了 ctx。

## 二、软删除（GORM `gorm.DeletedAt`）
- `deleted_at` 上做软删除：查询自动追加 `deleted_at IS NULL`，`Delete` 自动改写为 `UPDATE ... SET deleted_at=?`。
- 验收断言「删除后列表/详情不可见」→ 手写物理 `DELETE` / 用 `Unscoped()` 绕过 / 自定义 SQL 不带条件 → 掉坑。

## 三、分页（offset/limit + total 同源）
- `Count` 的过滤条件必须与 `Find` 一致；`total` 与当前页 `content` 同源。
- 验收断言 `pageSize` 生效、翻页不重复、总数准确 → 两处 where 不一致 / 写死每页数 → 掉坑。

## 四、统一返回（Result.code）
- 所有接口返回 `common.Result`（`code=0` 成功）。异常也要收口成 Result，**不要**靠 HTTP 状态码表达业务失败。
- 验收断言「业务失败也走统一结构、带中文原因」→ 直接 `c.AbortWithStatus(500)` 不带 body / 抛 panic → 掉坑。

## 五、Redis 契约
- 取同一个 `*redis.Client`；key 前缀统一（`demo:`）、TTL 明确。
- 验收断言「跨模块读得出对方写的数据、TTL 覆盖查询窗口」→ key 不一致 / 忘了 TTL → 掉坑。

## 六、事务
- 跨多次写库用 `db.Transaction(func(tx *gorm.DB) error {...})`；tx 也要带 ctx（`tx.WithContext(ctx)`）。
- 验收断言「失败整体回滚、部分写入不留痕」→ 分多次独立 `db` 调用（无事务）→ 掉坑。
