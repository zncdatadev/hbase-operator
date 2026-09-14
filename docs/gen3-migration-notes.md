# Gen 3 迁移说明（BaseCluster → GenericReconciler）

本次重构将 hbase-operator 从 Gen 2b（operator-go v0.12.6 `BaseCluster` + 每角色子包）迁移到
Gen 3（**operator-go v0.13.0**：`GenericReconciler` + `RoleProvider`/`RoleGroupResolver` + `BaseRoleGroupHandler` + per-CR `ExtensionRegistry`）。
方法论：YAML 平价优先 —— 既有 e2e（`test/e2e/`）是行为契约；同时补入 Gen 3 生命周期、
外部依赖 watch、状态与 Vector 的回归断言。渲染资源与迁移前一致，显式列外的差异见下文
intentional-diff 清单。

## 架构映射

<!-- markdownlint-disable MD013 -->

| 迁移前（Gen 2b） | 迁移后（Gen 3） |
| --- | --- |
| `internal/controller/hbasecluster_controller.go` + `cluster/` | `reconciler.GenericReconciler`（框架） |
| `master/`、`regionserver/`、`restserver/` 每角色包 | CR `GetSpec()` 桥接出的 `Roles` map + 单一 `HbaseRoleGroupHandler` |
| `common/statefulset.go`（手建 STS） | 框架构建，产品只**声明**：`DeclareRoles` 返回的 `RoleDeclaration`（入口脚本进 `Command`、探针、端口、主容器名、`ConfigDefaults` 里的默认亲和）+ `buildCtx.VolumeProviders`（HDFS 配置卷、Kerberos CSI 卷）。全部在框架 `Build()` **之前**生效，故用户 podOverrides 仍最高优先 |
| `common/configmap.go`（手建 ConfigMap） | `ResolveRoleGroup`（`RoleGroupResolver`，带 ctx/client，可直接读 zk discovery ConfigMap，返回 `Contribution` 作为最低合并层）+ `hbase-env.sh` 整文件 setIfAbsent |
| `common/service.go`（metrics Service） | handler `buildMetricsService` → `RoleGroupResources.MetricsService`（形状与 observability e2e 断言逐字段一致） |
| `authz/krb5.go` | `internal/controller/kerberos.go`（settings/卷注解逐字节平价） |
| `authz/oidc.go`（手建 oauth2-proxy 容器） | 框架 `sidecar.OAuth2ProxySidecarProvider` + `OidcCookieSecretExtension` |
| 旧代码内的镜像默认值 | CR `GetSpec()` 镜像适配器（repo/productVersion/kubedoopVersion 默认值不变） |
| 每 role group 日志渲染 | `RoleDeclaration.LogProducers`（log4j，`log4j.properties`，console pattern 不变） |
| （无） | 角色级 PDB 由框架从 `roleConfig.podDisruptionBudget` 构建（与 pdb e2e 断言一致：`hbase-<role>`） |
| 外部对象靠周期性 reconcile 间接刷新 | controller-runtime 字段索引 + `SetupWithManagerOptions.Watches`：被引用的 ZooKeeper/HDFS/Vector ConfigMap、OIDC Secret 和 cluster-scoped AuthenticationClass 变化会直接 enqueue 对应 HbaseCluster |

<!-- markdownlint-enable MD013 -->

## Intentional-diff 清单（渲染资源与迁移前的已知差异）

1. **新增每 role group 的 `<name>-headless` Service**（框架为 STS 网络身份构建）。同时
   STS `spec.serviceName` 指向它；对存量集群该字段不可变——框架 apply 路径会保留旧值并发
   `ImmutableFieldIgnored` 事件，不阻塞升级。
2. **标签/selector 集合变化**：框架 selector 为 `instance + component + managed-by=operator-go +
   role-group marker`。若旧 STS selector 的 `managed-by` 值不同，存量 STS 的 pod template 更新会被
   API server 拒绝（selector 不可变且必须匹配 template）。**这是 breaking release**：存量集群升级
   需手动 `kubectl delete sts --cascade=orphan` 后由新 operator 重建。发布说明必须包含此步骤。
   （e2e 每次全新建集群,不受影响。）
3. **oauth2-proxy 形态**：容器名 `oidc` → `oauth2-proxy`；普通容器 → native sidecar
   （init container + restartPolicy Always，带 readiness/liveness `/ping` 探针）；镜像从产品镜像内
   置二进制换为上游 `quay.io/oauth2-proxy/oauth2-proxy`；cookie secret 从 CR UID 派生（可伪造，
   不安全）改为一次性生成的 `<cluster>-oidc-cookie` Secret（key `COOKIE_SECRET`）。
   用户凭据 Secret 契约不变（CLIENT_ID/CLIENT_SECRET）。

   ⚠️ **这条抬升了最低 Kubernetes 版本**：native sidecar 上的探针要 k8s **1.29+**（1.33 GA）。
   在更老的集群上，启用 OIDC 的集群其**每个** Pod 都会被 API server 拒绝
   （`spec.initContainers[0].livenessProbe: Forbidden: may not be set for init containers`），
   StatefulSet 建了但永远创建不出 Pod、也不会 CrashLoop，只在 STS 的 `FailedCreate` 事件里可见。
   本次迁移已把 `Makefile` 的 `KIND_K8S_VERSION` 默认值从陈旧的 `1.26.15` 对齐到 CI 的 `1.35.0`
   （原默认值与 CI 矩阵脱节，本地 `make setup-chainsaw-cluster` 必然撞上此问题）。
   若必须支持 <1.29 的集群，可用框架的 `SidecarConfig.Probes{DisableReadiness, DisableLiveness}`
   关掉探针 —— 代价是失去"代理挂了把 Pod 移出 Service"的保护。
4. **Vector agent 门控**：旧行为是 `vectorAggregatorConfigMapName` 非空即注入 vector 容器；
   新行为遵循框架三重门（`logging.enableVectorAgent` + 声明的 producer + vector.yaml 来源）。
   滚动日志文件名 `hbase.log4j.xml` → `<role>.log4j.xml`（框架约定）。observability e2e 会启动
   master 的 native Vector sidecar，核验生成的 `vector.yaml`、log4j 文件目标、二进制与 9598 metrics。
5. **ServiceAccount**：pods 由 default SA 改为框架**派生**的 `hbasecluster-<cluster>`
   （`reconciler.ServiceAccountResourceName(kind, cluster)`，超长时带 sha256 后缀）。
   operator-go #616 移除了 `ServiceAccountName`/`ServiceAccountNameFunc`：名字不再可配置，因为框架
   创建它、controller-own 它、随 CR 回收它、并把工作负载的 Role 绑到它，没有任何东西需要按产品
   自选的名字寻址它。Kind 进入名字是因为单个 namespace 内 CR 名不唯一。
   本仓库先前设的 `hbase-<cluster>` 已删除；从该中间状态升级会滚动重启一次并遗留旧 SA 待手工清理。
6. **config 卷**：卷名 `hbase-config` → `config`（框架保留名），挂载只读；挂载路径不变
   （`/kubedoop/mount/config/hbase`）。`hdfs-config` 卷不变。
7. **hbase-site.xml / ssl-*.xml 由 XML adapter 渲染**：键值不变（用户 `configOverrides` 仍最高优先），
   XML 排版可能与旧 `xml.XMLConfiguration` 输出有格式差异。
8. **容器 env 顺序**：env 改经 envOverrides 合并管道（排序输出），值不变。
9. **SecurityContext**：框架默认注入 1001/0/1001 + 硬化集；Kind 上的全量 Chainsaw 已通过
   HBase 实际启动验证，未观察到 HDFS 数据属主问题。
10. **状态子资源**：`status` 增加 `roleGroups` 账本与 `observedGeneration`；conditions 语义改为框架
    条件集（Available/Progressing/Degraded/Paused/ServiceHealthy/ReconcileComplete 等）。
11. **kerberos discovery config**（`GetDiscoveryConfig`）在旧代码中未被引用，未迁移。
12. **修正 metrics Service 路由**：三个角色仍暴露 HBase 原生 UI/Prometheus 端口
    16010/16030/8085，但 `targetPort` 统一指向实际存在的 `ui-http`，不再误指 master 中不存在、
    另两角色中为 9100 的 `metrics` 端口。
13. **修正 HBase 角色 JVM 变量**：`HBASE_<ROLE>_OPTS` 使用 HBase 实际识别的大写角色名
    (`HBASE_MASTER_OPTS` / `HBASE_REGIONSERVER_OPTS` / `HBASE_RESTSERVER_OPTS`)。

## 验证状态

- [x] `make generate && manifests && helm-crd-sync` —— **CRD diff 为纯新增**：结构变化只有 status 的
      `roleGroups` 与 `observedGeneration`；其余为上游文档扩写与 6 条新 `pattern` 校验。60 行删除全部是
      role config 块内的 `default:`（gracefulShutdownTimeout/日志级别/storage/pdb），即上游 #544 修复
      "继承块内不得有 schema 默认值"，由 `api/v1alpha1/crd_schema_test.go` 的守卫断言。
- [x] `make lint`（0 issues）、`make test`、`go build`、`go vet` 全绿
- [x] RBAC 纯新增：serviceaccounts / events / persistentvolumeclaims / pods（Gen 3 框架所需）
- [x] **渲染平价单测** `internal/controller/handler_test.go`：入口脚本（三角色 subcommand、config 拷贝、
      信号处理）、探针端口、config/hdfs-config 卷、默认亲和、metrics Service 逐字段、Kerberos
      （ssl-*.xml、hbase-site 键、CSI 卷注解）、OIDC sidecar（upstream 端口、cookie 引用而非内联）
- [x] **产品级 GenericReconciler 集成测试**：可选角色；Vector 成功链与缺失 aggregator 的 fail-close；
      pause 不改资源、unpause 恢复漂移修复、stop 缩到 0、resume 恢复副本；status roleGroups/conditions
- [x] **外部依赖 watch 单测**：ConfigMap/Secret 按 namespace 映射，AuthenticationClass 跨 namespace
      映射；重复/空引用去重
- [x] 生产 Go 代码量：`internal/` 3384 → 1372 行（-59%；包含新增的外部依赖 watch）
- [x] chainsaw 全量（default/kerberos/oidc/pdb/observability）—— Kind Kubernetes 1.35.0 上
      HBase 2.6.1 与 2.6.2 均通过；observability 同时验证 Vector 0.47.0、三角色原生 metrics 与
      Prometheus targets，且 Prometheus release/RBAC 按测试 namespace 隔离并显式清理
- [x] go.mod 固定到正式发布版 **operator-go v0.13.0**（不再是本地 replace 或 pseudo-version）

## 迁移中发现并在本轮修复的既有缺陷

1. **metrics Service 曾指向错误命名端口**：Service 的 `targetPort: metrics`，但 master 角色
   的容器端口只有 `master`(16000) 和 `ui-http`(16010)，没有名为 `metrics` 的端口 → 该 Service 无法路由。
   regionserver/restserver 虽有 `metrics`(9100) 端口，但 Service 声明 `port: 16030/8085 → targetPort:
   metrics(9100)`，与注解自相矛盾。本轮已改为 `targetPort: ui-http`，并同步 unit/Chainsaw 断言。
2. **`HBASE_<role>_OPTS` 曾使用小写角色名**（如 `HBASE_master_OPTS`），HBase 不读取该变量。
   本轮改为大写角色名，并由 Kerberos e2e 同时断言大写变量存在、小写变量不存在。

## 仍需注意的既有行为

1. **旧默认反亲和从未生效**：旧代码用 `app.kubernetes.io/name: hbase` 做 matchLabels，但旧 pod 实际带的是
   `name: hbasecluster`（由 GVK Kind 小写推导），选择器永不匹配。Gen 3 下 `name` 标签变为 `hbase`
   （handler 的 ProductName），反亲和**开始真正生效** —— 单节点 kind 集群上多副本角色可能因此变得
   分散不下去，e2e 若出现 Pending 需优先怀疑此处。

## operator-go 框架反馈（已由 framework-steward 独立核实）

**已在框架内、本次迁移已改用框架 API（原以为是缺口，实为重复造轮子）**：

- `reconciler.EnsureGeneratedSecret`（#583）已实现"生成一次、永不重写、但补齐丢失的键"，
  且比本仓库最初手写的扩展更强：手写版只判断 Secret 是否存在，若 Secret 存在而 `COOKIE_SECRET`
  键丢失（部分恢复、手工编辑），oauth2-proxy 的 `Validate` 会每轮失败且**永不自愈**。已替换。
- `builder.MetricsServiceBuilder` 已存在，`buildMetricsService` 已改为调用它。与手写版的唯二差异
  （`Type: ClusterIP`、`Protocol: TCP`）都是 API server 默认值，live object 零差异，
  observability e2e 断言不受影响。
- `security.SecretProvisioner` 的注解面**完全覆盖** hbase 所需（`WithKerberosServiceNames` →
  `kerberosServiceNames`，`WithPassword` → `tlsPKCS12Password`）。本次仍保留手写卷以求字节平价：
  切换需处理 scope 顺序、storage 1Mi vs 默认 10Mi、挂载根、ReadOnly 四处差异，且框架
  `KerberosVolume` 会额外写 `secrets.kubedoop.dev/format: kerberos`，**需先与 secret-operator
  确认该注解缺省时的行为**，故留作后续独立变更。

**已被上游修复（升级到 operator-go `v0.13.0` 时确认；该发布与 `e8a9495` 代码一致，仅多一个 CHANGELOG 提交）**：

- **`ProductConfig` 无法读取集群** —— 上游 #591 直接把这条修了：`ProductConfig` 整体被
  `RoleGroupResolver`（带 ctx/client/error）取代。本仓库的 `injectZookeeperSiteKeys` 旁路已删除，
  zk 三键回到正规位置（`ResolveRoleGroup`），且框架保证其贡献落在用户 overrides **之下**。
  上游 commit message 里写明"ZERO operators used the hook while two of them hand-wrote the same
  workaround, byte for byte including its doc comment" —— 本仓库正是其中之一。

**待提 issue（成立）**：

1. **`AffinityBuilder` 应上提到框架（HIGH）** —— 去空白后**逐字节相同**的四份副本存在于
   hdfs / zookeeper / superset / hbase，本次迁移又抄了第五遍（`internal/controller/affinity.go`）。
   默认策略（instance 亲和 20 + instance/component 反亲和 70）在四个仓库中同构，产品差异只有标签值。
   框架侧只有 `DecodeAffinity`（用户写的 affinity），构造侧从未上提。建议
   `builder.NewAffinityBuilder` + `DefaultRoleAffinity(productName, clusterName, roleName)`，
   并在 `BaseRoleGroupHandler` 提供"config/podOverrides 都没给才生效"的兜底钩子（避免每个产品
   自己写 `if podSpec.Affinity == nil`，本仓库 handler.go 正是这么做的）。四个 operator 各删约 123 行。
2. **`ProductConfig` 文档承诺了签名做不到的事（HIGH docs / MEDIUM API）** —— 同一段 doc comment 内
   既说"may derive from live cluster state — e.g. a ZooKeeper connection string built from the actual
   resources"，又说"It is a pure function of the CR"。签名无 ctx/client。需求面比原估计**窄**：
   兄弟 operator 多为按引用透传（kafka/dolphinscheduler/nifi/hdfs 都不在 operator 端读取 discovery
   内容），目前只有 hbase 需要拆分变换。修复优先级：先改文档（删掉那个例子并指明正确做法是在
   `BuildResources` 写 `buildCtx.MergedConfig`），API 增强（`ProductConfigWithContext`）次之。
   框架内已有同形状先例：`vector.DiscoverAggregatorAddress` + `resolveVectorAggregatorAddress`。
3. **缺少 Gen 2b → Gen 3 迁移文档（HIGH）** —— 本文档"intentional diff #2"经核实成立且**不是 hbase
   特有**：v0.12.6 的 `managed-by` 值为 `kubedoop.dev`（`builder/object.go`），main 为 `operator-go`
   （`base_role_group_handler.go` `managedByValue`），且该标签位于**不可变的** `.spec.selector` 内 →
   每个带存量集群迁移的 operator 都会被 API server 硬拒。失败形态是两段式且第一段静默：apply 路径保留
   live selector 并发 `ImmutableFieldIgnored` 事件，随后 STS 更新以 "selector does not match template
   labels" 被拒。框架 `docs/` 下无任何迁移文档，而 org 规则是"文档对代码具有权威性" → 缺失文档是一等
   缺陷。建议新增 `docs/migration-gen2-to-gen3.md`，本文件的 intentional-diff 清单可作初稿。

**本次迁移自身被指出并已修复的缺陷**：最初 `customizeStatefulSet` 在 `BuildResources` **返回之后**
改 StatefulSet，而 podOverrides 的 strategic merge 发生在框架 `Build()` **内部** → 无条件赋值的
`Command`/`Args` 会静默覆盖用户 podOverrides（精确的优先级倒置），追加卷则绕过了
`PodOverrideViolations()` 校验（用户若声明同名卷 → 重名 volume 被 API server 硬拒）。
已改用框架 #585 引入的 `MainContainerCustomizer`（在合并**前**运行）与 `buildCtx.VolumeProviders`。
