# BUG_REPRO

## Bug 是什么

周起始日期归一化使用了错误的星期索引，导致周三无法正确归一到周一；时间解析错误地拒绝零点；营业时间边界判断还允许零长度班次。

## 如何触发

```bash
go test ./...
```

## 错误信息

测试失败时会看到：

- `Monday index = 1, want 0`
- `week start = 2026-08-16, want 2026-08-17`
- `ParseClock should accept midnight, got invalid clock`
- `zero-length shift should not fit`
