-- 修复 token_usage_daily 表中因时区问题导致的日期偏移
--
-- 问题根因：
--   pgx 驱动将 Go 的 time.Time(Asia/Shanghai) 写入 PostgreSQL DATE 列时，
--   先转为 UTC 再截取日期，导致东八区 2月17日 00:00+08 变成 UTC 2月16日 16:00，
--   存入数据库后日期变成了 2月16日。
--
-- 修复方式：
--   基于 token_usage 明细表的 created_at（timestamptz，始终正确）重建每日汇总数据。
--   对 created_at 按 Asia/Shanghai 时区提取日期，重新聚合统计。

BEGIN;

-- 清除旧的（可能有时区偏移的）汇总数据
TRUNCATE token_usage_daily;

-- 从明细表重建每日汇总（按 Asia/Shanghai 时区提取日期）
INSERT INTO token_usage_daily (user_id, usage_date, prompt_tokens, completion_tokens, total_tokens, request_count, created_at, updated_at)
SELECT
    user_id,
    (created_at AT TIME ZONE 'Asia/Shanghai')::date AS usage_date,
    SUM(prompt_tokens),
    SUM(completion_tokens),
    SUM(total_tokens),
    COUNT(*),
    MIN(created_at),
    MAX(created_at)
FROM token_usage
GROUP BY user_id, (created_at AT TIME ZONE 'Asia/Shanghai')::date;

COMMIT;

-- 验证查询（可在执行前用 BEGIN + ROLLBACK 测试）
-- SELECT usage_date, SUM(total_tokens), SUM(request_count) FROM token_usage_daily GROUP BY usage_date ORDER BY usage_date DESC LIMIT 10;
