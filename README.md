# Agent OS Indexer

独立的知识库后台索引进程。Knowledge 是数据库结构和事务的唯一维护方；Indexer 消费固定版本的 `indexing` Go 包，不包含业务 SQL、迁移、上传、检索或管理路由。

## 调度与恢复

每 3 秒扫描，补齐批次有界，每轮最多读取 10 个待处理 ID，并逐文件处理。待处理 ID 只是提示，真正互斥及状态重读由 Knowledge 协议负责。文件处理用一条连接持有 advisory lock，另一条执行查询，连接池至少 2、默认 3；不能不经重设计就把并发提高到连接池上限。

LightRAG 上传效果未知时等待人工核对；已有 track ID 只轮询，不重新上传。Milvus 重试使用稳定分块 ID。处理始终使用文件留存的固定配置，不跟随 Registry 最新版本变化。供应商或数据库失败不会终止调度器，后续轮次继续恢复；取消停止后续文件处理。

## 独立构建与运行

需要 Go 1.27.2，数据域使用 PostgreSQL。

```powershell
$env:GOWORK='off'; $env:GOPROXY='off'
./scripts/Build.ps1
go test -mod=vendor -count=1 ./...
go vet -mod=vendor ./...
```

也可 `go build -mod=vendor ./cmd/indexer` 或使用本仓 Dockerfile。固定依赖和提交的 `vendor` 不要求相邻 Knowledge 源码。

[配置示例](deploy/service.example.json) 使用独立的应用角色及 `search_path=knowledge`，默认端口 8097，空 Registry/Runtime 地址保留回环 8096/8094。多机部署显式填写服务地址。就绪检查只验证本进程所需表，Runtime 接收实例心跳。

## 数据权限与版本兼容

必须由 **Knowledge 的迁移程序**一次授予 Knowledge 与 Indexer 两个账号权限；Indexer 不附带也不执行迁移。新账号只需 documents/files/chunks SELECT 与 files/chunks UPDATE，不需上传、Registry、Runtime、artifact 或 DDL 权限。

若旧单仓 Indexer 仍是回滚目标，应保留它 readiness 所需的 `knowledge.artifacts`、`knowledge.artifact_refs` SELECT。撤销旧权限意味着结束旧程序回滚窗口，不能同时声称兼容。

发布时固定 Indexer 提交、Knowledge 模块版本/校验和及 [发布契约](deploy/release-contract.json)。独立更新和回滚必须保留旧资料及未完成导入；不做反向数据库迁移。

## 验证状态

单元测试覆盖调度失败和取消。Knowledge 的真实 PostgreSQL 测试覆盖两个受限角色、上传中断、未知效果和 Milvus 稳定 ID，外部提供者为模拟服务。Windows 独立升级/回滚及真实旧/新 Indexer 接力已验证，但这些不证明真实 RAG 后端可用、机器故障高可用或最终容量。

Actions 日常关闭。[历史记录](docs/history/service-notes-20261014.en.md) 保留最初旧 readiness 权限失败和后续验收边界。
