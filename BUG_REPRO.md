# BUG_REPRO

## Bug 是什么

跨午夜时间算术会直接回绕到次日而不是拒绝，反向时间段被当成合法的跨日时长，营业时间边界判断也允许零长度班次。

## 如何触发

```bash
go test ./...
```

## 错误信息

测试失败时会看到：

- `AddMinutes should reject past-midnight result, got <nil>`
- `DurationMinutes should reject backward range, got <nil>`
- `zero-length shift should not fit`
