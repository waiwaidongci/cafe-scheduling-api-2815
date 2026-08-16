# BUG_REPRO

## Bug 是什么

排班工时统计和重叠处理使用了未初始化的 map，遇到正常排班数据时会向 nil map 写入，导致 panic。

## 如何触发

```bash
go test ./...
```

## 错误信息

测试进程会 panic，关键信息为：

```text
panic: assignment to entry in nil map
```

panic 发生在工时统计与重叠合并的 map 分组路径上。
