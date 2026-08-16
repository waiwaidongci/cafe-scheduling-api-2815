# BUG_REPRO

## Bug 是什么

排班重叠合并没有正确合并连续重叠班次，冲突检测会漏掉参与冲突的班次，去重工时统计也没有按重叠区间去重，导致工时被多算。

## 如何触发

```bash
go test ./...
```

## 错误信息

测试失败时会看到：

- `expected 2 merged spans, got 3`
- `expected 1 conflict, got 0`
- `expected 360 unique minutes, got 420`
