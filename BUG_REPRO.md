# BUG_REPRO

## Bug 是什么

排班计算中的错误处理没有保持约定的错误类型。无效时间段、反向时间段和非法日期在部分路径上返回了 nil 错误、错误了 sentinel，或者把不应匹配的营业时间范围直接判为合法。

## 如何触发

```bash
go test ./...
```

## 错误信息

测试会在错误处理用例失败，并提示 `DurationMinutes should preserve ErrInvalidTimeRange, got <nil>`。继续修复其他路径后，还会出现 `NewInterval should preserve ErrInvalidTimeRange`、`ParseWeekRange should preserve ErrInvalidDateRange` 以及营业时间范围判断反向的失败。
