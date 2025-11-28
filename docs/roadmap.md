# 🗺️ Project Roadmap (项目路线图)

> **Project Goal**: To build a high-performance, enterprise-grade EVM event monitor capable of handling chain reorgs and high concurrency.
>
> **项目目标**: 构建一个高性能、企业级的 EVM 事件监控系统，具备处理区块回滚（Reorg）和高并发处理能力。

---

## 📅 Phase 1: Infrastructure & Foundation (基础设施)
**Status**: ✅ Completed (已完成)

- [x] **Project Initialization**: Setup Standard Go Project Layout. (建立标准目录结构)
- [x] **Configuration Management**: Integrate `Viper` for multi-environment support (Dev/Prod). (集成 Viper 配置管理)
- [x] **Logging System**: Implement structured logging with `Zap`. (集成 Zap 结构化日志)
- [x] **Database Design**: Design `BlockTrace` and `TransferEvent` models with Gorm. (数据库模型设计)

---

## 📅 Phase 2: Core Scanner (核心扫描器)
**Status**: 🔄 In Progress (进行中)
**ETA**: 2 Days

- [ ] **RPC Client**: Encapsulate `ethclient` with connection keep-alive/retry logic. (封装 RPC 客户端)
- [ ] **Block Iterator**: Implement logic to fetch block numbers and iterate from `Start` to `Current`. (实现区块遍历器)
- [ ] **Log Fetching**: Use `FilterLogs` to fetch raw event logs from the chain. (抓取原始链上日志)

---

## 📅 Phase 3: Parser & Persistence (解析与存储)
**Status**: ⏳ Pending (待启动)
**ETA**: 3 Days

- [ ] **ABI Binding**: Generate Go bindings for ERC-20 smart contracts. (生成合约 ABI 代码)
- [ ] **Data Parsing**: Parse Hex logs into human-readable transfer events (From, To, Amount). (解析数据)
- [ ] **Persistence Layer**: Save valid transactions to MySQL using Gorm. (数据入库)
- [ ] **State Management**: Record the last scanned block to allow resume-from-break. (记录扫描位点，支持断点续传)

---

## 📅 Phase 4: Architecture Upgrade (架构升级) 🔥
**Status**: ⏳ Pending
**ETA**: 5 Days

- [ ] **Reorg Handling**: Implement block hash comparison to detect and handle chain forks. (实现区块回滚/分叉检测)
- [ ] **Pipeline Concurrency**: Refactor to Producer-Consumer model using Go Channels. (重构为流水线并发模型)
- [ ] **Graceful Shutdown**: Ensure no data loss during service restart. (优雅退出机制)

---

## 📅 Phase 5: DevOps & Delivery (部署与交付)
**Status**: ⏳ Pending
**ETA**: 2 Days

- [ ] **Dockerization**: Write `Dockerfile` and `docker-compose.yml`. (容器化)
- [ ] **Documentation**: Complete `README.md` with architecture diagrams. (完善文档)
- [ ] **Monitoring**: (Optional) Integrate Prometheus metrics. (可选：集成监控指标)

---

---

## 📅 Phase 6: Kubernetes Orchestration (K8s 容器编排) ☸️
**Status**: ⏳ Planned (规划中)
**Goal**: Deploy the system into a local K8s cluster to simulate enterprise production environments.

- [ ] **Image Registry**: Push Docker images to Docker Hub or Aliyun Registry. (镜像推送)
- [ ] **K8s Config Management**: Migrate `config.yaml` to **ConfigMap** and `Secrets`. (配置迁移)
- [ ] **App Deployment**: Write `Deployment.yaml` for FlashMonitor with 3 replicas. (多副本部署)
- [ ] **Stateful Workloads**: Deploy MySQL & Redis using **StatefulSet** and **PVC** (Persistent Volume Claim). (有状态服务部署)
- [ ] **Service Discovery**: Expose Adminer via **Service (NodePort/LoadBalancer)**. (服务暴露)
- [ ] **Health Checks**: Configure Liveness and Readiness probes for the Go application. (健康检查)

## 🛠 Tech Stack (技术栈)
* **Language**: Golang 1.20+
* **Blockchain**: Go-Ethereum (Geth)
* **Database**: MySQL 8.0
* **Cache**: Redis 7.x
* **ORM**: Gorm
* **Config**: Viper
* **Logging**: Zap
* **Deployment**: Docker Compose & Kubernetes (K8s)