# operator-go issue 草稿（hbase Gen 3 迁移产出）

<!-- markdownlint-disable MD024 -->

> **待观察（尚未成稿）**：apply 路径把**可重试的写冲突**当成硬错误上报 —— 每次都在 CR 上留下
> ERROR 日志和 `ReconcileError` Warning 事件，且置 Degraded。一轮 e2e 里观察到两种形态，
> 两者都在下一次 reconcile 自愈、e2e 用例全部通过：
>
> 1. 集群删除时，已被 GC 回收的 Service 触发 UID precondition 失效：
>    `Precondition failed: UID in precondition: ab5e2e42-..., UID in object meta: <empty>` ——
>    即一次**正常的删除**会报错。
> 2. StatefulSet 的标准 409：`Operation cannot be fulfilled on statefulsets.apps ...
>    the object has been modified` —— StatefulSet controller 写 status 与 operator 写 spec 撞车，
>    在任何有 StatefulSet 的产品上都会周期性出现。
>
> 两者都属于"对象在读与写之间变了"，与框架已经特殊处理的 429 `*RateLimitError` 同类，
> 应当退避重试而非置 Degraded。当前行为会让 `Degraded` 在稳态运行中偶发抖动，
> 而这正是文档所说"唯一值得告警"的那个 condition。
>
> 复现环境：operator-go `e8a9495`，k8s 1.35.0。提 issue 前应确认框架是否有意如此
> （apply 失败即 Degraded），以及是否已有 issue 覆盖。

三条均已由 framework-steward 对照本地 operator-go 源码核实。第 2 条已在 operator-go v0.13.0
以 `RoleGroupResolver` 修复，保留原始草稿作为迁移过程记录；第 1、3 条仍可作为上游 issue 候选。
**均未阻塞本次迁移**。提交前请确认是否已有重复 issue。

---

<!-- markdownlint-disable-next-line MD013 -->
## 1. `pkg/builder`: promote the default pod-affinity builder that four operators ship verbatim

**Labels**: `enhancement`, `pkg/builder`
**Priority**: HIGH

### Motivation

hbase-operator 的 Gen 2b→Gen 3 迁移把 Gen 2b 的 `AffinityBuilder` 原样搬进了新代码 —— 这是一次
本可以删掉 125 行的迁移。框架侧 `pkg/reconciler/affinity.go` 只有 `DecodeAffinity`
（RawExtension → `*corev1.Affinity`），即只覆盖"用户写的 affinity"这一半；"产品默认 affinity"
这一半从未上提，于是每个 operator 各留一份。

### Downstream evidence

去空白后**逐字节相同**的四份副本：

- `hdfs-operator/internal/common/affinity.go`(123 行)
- `zookeeper-operator/internal/common/affinity.go`(123 行)
- `superset-operator/internal/controller/common/affinity.go`(123 行)
- `hbase-operator/internal/controller/common/affinity.go`(129 行) → 迁移后 `internal/controller/affinity.go`(125 行)

```bash
diff <(sed 's/[[:space:]]//g' hdfs-operator/internal/common/affinity.go) \
     <(sed 's/[[:space:]]//g' zookeeper-operator/internal/common/affinity.go)   # 空输出
```

默认策略同样同构（instance 亲和 weight 20 + instance/component 反亲和 weight 70）：
`hdfs-operator/internal/common/role_config.go:42-43`、
`hbase-operator/internal/controller/common/statefulset.go:251-252`、
`superset-operator/internal/controller/common/statefulset.go:274`、
`zookeeper-operator/internal/common/role_config.go:109`。
产品差异只有标签值（产品名、role 名），是纯参数。

### Proposed API

`pkg/builder/affinity_builder.go`，与 `pkg/reconciler/affinity.go` 的 `DecodeAffinity` 配对：

```go
package builder

type PodAffinityTermSpec struct {
    Labels   map[string]string
    Anti     bool
    Required bool
    Weight   int32 // 0 => corev1.DefaultHardPodAffinitySymmetricWeight
}

type AffinityBuilder struct{ /* terms */ }

func NewAffinityBuilder(terms ...PodAffinityTermSpec) *AffinityBuilder
func (b *AffinityBuilder) Build() *corev1.Affinity

// 四个 operator 共同的默认策略，产品只传身份参数
func DefaultRoleAffinity(productName, clusterName, roleName string) *corev1.Affinity
```

建议同时在 `BaseRoleGroupHandler` 提供"config/podOverrides 均未给出 affinity 时才生效"的兜底钩子，
避免每个产品在 `BuildResources` 后自己写 `if podSpec.Affinity == nil`（hbase 目前正是这么做的）。

### Migration impact

四个 operator 各删约 123 行。若 `DefaultRoleAffinity` 保持上述权重与标签集，渲染 YAML **无变化**。
注意 Gen 2b 版本的 `RequiredDuringScheduling...` 恒为非 nil 空 slice，上提时需保持该行为或在
CHANGELOG 显式说明 —— 否则 `affinity` 字段的序列化会有差异。

---

## 2. [已解决于 v0.13.0] `ProductConfig` 的文档承诺了签名做不到的事（ZooKeeper connection string）

**Labels**: `documentation`, `pkg/reconciler`
**Priority**: HIGH（docs）/ MEDIUM（API）

### Motivation

hbase Gen 3 迁移需要把 zookeeper znode discovery ConfigMap 的三个键拆进 `hbase-site.xml`
（`hbase.zookeeper.quorum` / `zookeeper.znode.parent` / `hbase.zookeeper.property.clientPort`）。
文档明确把这件事写成 `ProductConfig` 的用途，实现时才发现签名里没有 ctx/client。
**代价主要是被文档误导的时间，而非最终代码质量** —— workaround 本身是正确的。

### Evidence

矛盾出现在**同一段 doc comment 内部**（`pkg/reconciler/generic_reconciler.go:167-175`）：

> `...may derive from live cluster state — e.g. a ZooKeeper connection string built from the actual resources...`
>
> `...It is a pure function of the CR and the role/role group identity...`

同一措辞另见 `AGENTS.md:604` 与 `docs/architecture.md:171`。
签名：`ProductConfig func(cr CR, roleName, roleGroupName string) *v1alpha1.OverridesSpec`。

下游 workaround：`hbase-operator/internal/controller/handler.go` 的 `injectZookeeperSiteKeys`
—— 在 `BuildResources` 里往已 merge 的 `buildCtx.MergedConfig` 写（`setIfAbsent`，用户
`configOverrides` 仍然胜出）。

**需求面**比初估要窄：兄弟 operator 绝大多数按引用透传，operator 端从不读取 discovery 内容
（kafka `container.go:70`、dolphinscheduler `container.go:79-81`、nifi `configmap.go:431-435`
用容器启动期模板、hdfs `container.go:41`）。只有需要在 operator 端**拆分/变换** discovery 值的
产品才需要读取，目前仅 hbase。

框架内已有同形状先例：`pkg/vector/discovery.go:43 DiscoverAggregatorAddress(ctx, c, ns, cmName)`，
由 `generic_reconciler.go:1082 resolveVectorAggregatorAddress` 调用 —— 框架自己为 vector 做了
这件事，只是没把能力开放给产品。

### Proposed fix

**A（最小、建议先做）**：修文档。删掉 "a ZooKeeper connection string built from the actual
resources" 这个例子（三处），保留 "JVM heap sized from resources"、"quorum peer list from pod
ordinals"（确为 CR 的纯函数），并显式写明："需要读取 live 对象的配置，请在
`RoleGroupHandler.BuildResources` 里写入 `buildCtx.MergedConfig`。"

**B（可选，与 A 不互斥）**：补齐能力，旧钩子不变：

```go
ProductConfigWithContext func(ctx context.Context, c client.Client, cr CR,
    roleName, roleGroupName string) (*v1alpha1.OverridesSpec, error)
```

两者都做时需明确优先级（建议：设置 B 则忽略 A 并在启动期 log 一次）。B 引入了可返回 error 的语义，
失败时是 Degraded 还是 abort 需决定，建议复用 `resolveVectorAggregatorAddress` 的"响亮失败"先例。

### Migration impact

(A) 纯文档，零代码影响。(B) 新增可选字段，未设置时行为不变 → 现有 operator 渲染 YAML 无变化。

---

## 3. 缺少 Gen 2b → Gen 3 迁移文档：`managed-by` 值变更落在不可变的 `.spec.selector` 里

**Labels**: `documentation`, `breaking-change`
**Priority**: HIGH

### Motivation

hbase 迁移中发现：**存量集群**升级到 Gen 3 会被 API server 硬拒，需要人工
`kubectl delete sts --cascade=orphan`。这不是 hbase 特有 —— 它是**每一个**带存量集群迁移到
Gen 3 的 operator 的必经故障，而框架 `docs/` 下没有任何迁移文档。org 规则是"文档对代码具有
权威性"，因此缺失的迁移文档是一等 bug。

### Evidence

- v0.12.6：`pkg/builder/object.go:93` 写入 `LabelKubernetesManagedBy: constants.KubedoopDomain`，
  而 `pkg/constants/constants.go:32` `KubedoopDomain = "kubedoop.dev"`
- main：`pkg/reconciler/base_role_group_handler.go:48` `managedByValue = "operator-go"`，
  由 `frameworkSelectorLabels` 放进 selector
- 两条路径都无法兼容：`LabelDomain == ""` 走 `frameworkSelectorLabels`（含变更后的 `managed-by`）；
  设置 `LabelDomain` 则整体切到 `<domain>/cluster` 等产品标签 —— 同样与存量 selector 不匹配

**失败形态是两段式的，且第一段是静默的**：apply 路径按 `copyDesiredState` 保留 live 的不可变
`selector` 并发 `ImmutableFieldIgnored` 事件，但 pod template 标签已带新值 → 随后的 StatefulSet
更新被 API server 以 "selector does not match template labels" 拒绝。用户看到的是一个 Warning
事件加一个看似无关的更新失败。

其他正在迁移的 operator 会撞到同一堵墙（superset-operator、dolphinscheduler-operator 均有
进行中的 gen3 worktree）。

### Proposed fix

新增 `docs/migration-gen2-to-gen3.md`，至少覆盖：

1. `.spec.selector` 中 `managed-by` 由 `kubedoop.dev` 变为 `operator-go` —— 存量 StatefulSet
   **必须** `kubectl delete sts --cascade=orphan` 后由新 operator 重建，给出确切命令与
   "pod 不重启"的前提条件
2. `LabelDomain` 设 / 不设两条路径的 selector 形状，及各自与 v0.12.6 的差异
3. 其余每个迁移者都要重新发现的既定差异：新增 per-role-group `<name>-headless` Service 与
   `spec.serviceName` 的 `ImmutableFieldIgnored`、Vector 三重门控与 `<role>.log4j.xml` 命名、
   `ServiceAccountNameFunc` 带来的 SA 变更、config 卷名 `config`

hbase-operator 的 `docs/gen3-migration-notes.md` 已逐条列出，可直接作为文档初稿来源。

代码侧出路（**不建议优先**，会把迁移期变量固化进框架）：

```go
// BaseRoleGroupHandler
// LegacySelectorManagedBy: 设置后 frameworkSelectorLabels 用该值而非 managedByValue，
// 让存量集群原地升级。默认空 => 现行为。
LegacySelectorManagedBy string
```

仅在确认多个 operator 有不可中断的存量集群时再考虑。
