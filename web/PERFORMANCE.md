# Go Web 性能优化验证

2026-09-29，linux/amd64，QEMU Virtual CPU version 2.5+，默认 GOMAXPROCS=16。

实现分为四个提交：请求取消检查、Allow 缓存、全局值不可变 map、快照构建工作切片复用。保持公开 API、请求生命周期取消、原子路由发布和在途快照隔离。

## 验证范围

- 取消：Stop、预先取消、handler 之间取消、过期 deadline、主动 Done、正常返回和 panic 清理。
- Allow：静态/参数/通配路径、多方法排序、HEAD 回落、ANY 排除、热注册和旧快照隔离。
- 全局值：覆盖、缺失、显式 nil、parent 与外部 wrapper、非法/非自反/包含不可比较值的 key、typed nil、并发更新和在途旧值。
- 快照：已有分组分裂、链隔离与并发注册测试，以及新增深层双分支、多方法、中间件注册顺序测试。

| 工具链 | go test ./... | -race | -tags=jsoniter | -tags=sonic |
| --- | --- | --- | --- | --- |
| Go 1.27.1 | 通过 | 通过 | 通过 | 依赖构建失败 |
| Go 1.22.12 | 通过 | 通过 | 通过 | 通过 |

Go 1.27.1 的 sonic v1.14.2 在 internal/rt/stubs.go 报 GoMapIterator 未定义，优化前代码同样失败。本轮不变更 JSON 依赖。

## 测量方法

优化前版本取自 `85072f2^`，复制相同的 performance_bench_test.go。前后在同一机器、同一工具链、同一参数下串行运行，无并行基准进程。每项五次，表中使用中位数；虚拟机耗时有噪声，分配数据更稳定。

```sh
go test -run '^$' -bench 'Benchmark(Request|Global|Build|Allow)Optimization' -benchmem -benchtime=100ms -count=5
GOTOOLCHAIN=go1.22.12 go test -run '^$' -bench 'Benchmark(Request|Global|Build|Allow)Optimization' -benchmem -benchtime=100ms -count=5
```

请求矩阵包含 static/param/wildcard/404/405、0/5/20 个中间件、background/cancelable parent；另测三方法 Allow。全局值固定 key 集合，预更新 0/1,000/100,000 次；命中查询选择早期 key，以暴露历史链遍历成本。更新测量不包含初始化。构建测量仅包含 publish，注册在计时外完成，组合 100/1,000 条路由、1/10 层、1/3 方法、每层 0/5/20 个中间件。

## 兼容性与成本

全局 map 为浅复制，写入成本为 O(不同 key 数量)，适合读多写少；短历史的查询和更新可能比原 value chain 慢。只有直接持有的本库 map 快照会被合并，外部 context wrapper、非自反或不能安全作为 map key 的合法 key 保留 context.WithValue 路径，因此这些特殊路径仍可能形成历史链。

路由仍深拷贝并原子发布整棵树；逐条热注册仍反复重建。中间件链各自拥有独立数组，DFS 临时切片不会保存在快照中。

## 实测结果

完整矩阵的五轮中位数见 [performance-results.csv](performance-results.csv)。其中 `matrix` 为 100ms 样本，`global-final` 是剖析修复后的最终全局值实现，`allow` 和 `recheck-500ms` 为 500ms 样本。时间单位为 ns/op，分配为 B/op、allocs/op。

| 场景 | Go 1.27.1 前 → 后 | Go 1.22.12 前 → 后 |
| --- | --- | --- |
| 静态请求，0 中间件，background | 965.5 → 606.1 ns；208 B/3 → 96 B/2 | 756 → 561.6 ns；192 B/3 → 96 B/2 |
| 参数请求，0 中间件，background | 1007 → 713.2 ns；208 B/3 → 96 B/2 | 979.2 → 634.4 ns；192 B/3 → 96 B/2 |
| 多方法 405 | 1133 → 708.1 ns；176 B/5 → 112 B/3 | 1185 → 752.7 ns；176 B/5 → 112 B/3 |
| 单方法 405，0 中间件，background | 838.8 → 694.7 ns；112 B/3 → 112 B/3 | 897.3 → 674.2 ns；128 B/4 → 112 B/3 |
| 1,000 路由、10 层、3 方法、每层 5 中间件构建 | 10.76 → 8.40 ms；26,099 → 14,090 次分配 | 13.05 → 8.81 ms；29,115 → 14,107 次分配 |

普通成功请求在所有 parent/中间件组合下均少一次分配。Go 1.22 的 Done channel 比 Go 1.27 小，因此节省的字节数分别是 96 和 112。Go 1.27 的单方法 Allow 原切片已被编译器优化，不表现为分配减少；多方法拼接在两版本均减少两次分配。

全局值最终查询全部零分配。Go 1.27 的普通 key 已避开 `reflect.Value.Comparable` 的递归路径；剖析时该函数产生的额外分配已消除，复合 key 仍保留递归可比较性检查。

| 历史更新次数 | Go 1.27 命中 前 → 后 | Go 1.27 缺失 前 → 后 | Go 1.22 命中 前 → 后 |
| --- | --- | --- | --- |
| 0 | 33.10 → 71.13 ns | 34.84 → 61.58 ns | 31.81 → 81.56 ns |
| 1,000 | 11,100 → 67.63 ns | 6,788 → 57.31 ns | 10,981 → 68.50 ns |
| 100,000 | 1,100,880 → 63.49 ns | 808,999 → 61.38 ns | 1,341,910 → 85.13 ns |

固定两 key 更新在 Go 1.27 从约 148–171ns、48 B/1 次分配变为约 979–1045ns、360 B/3 次分配；Go 1.22 的量级相近。这是用写时复制换取稳定查询成本的明确代价，零历史查询也增加了数十纳秒。

## 耗时阈值复测

首次 100ms 采样有数个请求超过 +5%，不能直接按噪声忽略。延长到 500ms、五轮后：

- Go 1.22 的全部被标记请求均低于 +5%；最大为一个 404 场景 +4.6%，此前 +5.6% 的 wildcard/mw20/background 变为 -7.4%。
- Go 1.27 的深层多方法高数量中间件构建由首次 +7.2% 变为 -20.7%，分配仍从 26,099 降至 14,090。
- Go 1.27 的 404 场景大多落在 -4.0% 至 +3.3%；mw20/cancelable 剩余 +5.5%，进一步使用预编译测试二进制交替运行前后版本，避免先跑完全部 baseline 再跑 after 的时间漂移。

最后的交替复测使用相同参数 `-test.run=^$ -test.bench='^BenchmarkRequestOptimization/404/mw20/cancelabletrue$' -test.benchmem -test.benchtime=1s`，五对样本交替先运行 baseline/after。中位数 1247 → 1254ns，即 +0.6%，分配不变；记录为 CSV 的 `paired-1s`。本次测量没有发现稳定超过 5% 的请求耗时回退。虚拟机结果不代表生产环境固定收益。
