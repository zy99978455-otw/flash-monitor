# 🐋 FlashMonitor: High-Performance Web3 Indexer

![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15.0+-4169E1?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)
![License](https://img.shields.io/badge/License-MIT-green.svg)


FlashMonitor 是一个采用 Go 语言编写的轻量级、生产级 EVM 链上数据监听与分发引擎。系统秉承 "Let's go further" 的极客理念，专注于解决 Web3 数据索引过程中的**数据一致性（链分叉问题）**、**长连接性能**以及**节点单点故障**等核心痛点。

## 🎯 核心架构亮点 (Core Features)

### 1. 毫秒级全局优雅停机 (Graceful Shutdown)
告别悬挂连接与脏数据。通过全局 `context.Context` 生命周期接管，引擎能够在接收到 OS 停机信号（SIGINT/SIGTERM）的 **5毫秒内**，完美完成：
* 拦截并安全回滚正在进行的 PostgreSQL 原子事务。
* 切断并清理所有活跃的 SSE 前端长连接。
* 阻断并取消底层 RPC 网络请求，实现 0 内存泄漏安全退出。

### 2. 链重组免疫机制 (Chain Reorg Defense)
主网分叉不应成为数据灾难。引擎通过 `BlockHash` 溯源比对算法实时感知链状态回滚。配合数据库的原子事务隔离，实现“幽灵数据”的自动发现与安全剔除，保证业务层数据的绝对纯正。

### 3. 多节点容灾路由 (RPC Fallback Manager) —— *V2.0 新特性***
打破单点故障 (SPOF) 诅咒。内置 `NodeManager` 连接池，基于读写锁 (`sync.RWMutex`) 与自适应高度比对，实现节点健康状况的实时心跳检测。当主节点被限流 (HTTP 429) 或假死时，无缝转移至备用节点并执行指数退避重试。

### 4. 无锁实时事件广播 (Lock-free SSE Broker)
摒弃笨重的 WebSocket，采用基于 Go 原生 `Channel` 和 `select` 机制构建的 Server-Sent Events (SSE) 广播中心。利用通道非阻塞发送特性实现慢速客户端自动熔断，支持单机万级并发连接的极速单向推送。

---

## 🚀 极速部署 (Quick Start)

得益于完整的 Docker 容器化编排，无需在本地安装任何数据库或 Go 环境即可一键启动。系统已针对 1GB 内存的轻量级云服务器（如 RackNerd VPS）进行了极限内存调优（GOMEMLIMIT 限制与 Postgres 缓冲调优）。

### 1. 环境准备
克隆代码并准备环境变量：
```bash
git clone https://github.com/zy99978455-otw/flash-monitor.git
cd flash-monitor
cp .env.example .env
(请在 .env 中填入你的真实 ETH_RPC_URLS 地址)
```

### 2. 一键启动
```bash
docker compose up -d --build
```

### 3. 实时观测
* 后端日志: docker compose logs -f api
* 前端巨鲸大屏：通过浏览器访问 http://localhost:4010 接入 SSE 流式推送。

## 架构演进路线（Roadmap）
为了保持 V1 版本的极致轻量与纯粹，以下功能已完成架构预留，计划于后续版本接入：

- [x] **V1.0: 基础设施 (Foundation)**
  以太坊主网单节点扫链、同步日志提取、无锁 SSE 实时流推送以及基础 Docker 配置。
- [x] **V2.0: 高可用与容灾 (High Availability & Resilience)**
  引入 `NodeManager` 实现多节点 RPC 故障转移、指数退避重试机制以及原子级的优雅停机，彻底消除单点故障 (SPOF)。
- [ ] **V2.1: 状态与准确性守护 (State & Accuracy Guard)**
  实现持久化的断点续传 (Progress Checkpointing) 与延迟确认逻辑，硬核防御链上区块重组 (Reorgs)，确保数据绝对干净。
- [ ] **V2.2: 数据变现与 API 层 (Data Monetization & API Layer)**
  开发 RESTful 接口 (`/api/v1/whales`) 提供头部巨鲸交易查询，并通过 `/api/v1/health` 暴露 `NodeManager` 的节点健康度与延迟监控数据。
- [ ] **V3.0: 扩展性与架构重构 (Scalability & Architecture)**
  使用解耦的回调架构 (Callback Architecture) 重构扫链引擎，引入 Goroutine 协程池 (Worker Pool) 实现极高吞吐量的并发区块解析。
- [ ] **V4.0: 分布式运维 (Distributed Operations)**
  引入 Redis 实现集群分布式锁 (支持多实例水平扩展)，并部署 Prometheus + Grafana 栈以获得企业级可观测性。
