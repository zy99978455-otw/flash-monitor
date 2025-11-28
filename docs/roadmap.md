# 🗺️ Project Roadmap (项目路线图)

> **Project Goal**: To build a high-performance, enterprise-grade EVM event monitor capable of handling chain reorgs, high concurrency, and distributed processing.
>
> **项目目标**: 构建一个高性能、企业级的 EVM 事件监控系统，具备处理区块回滚（Reorg）、高并发处理、微服务解耦以及自动化运维能力。

---

## 📅 Phase 1: Infrastructure & Foundation (基础设施)
**Status**: ✅ Completed (已完成)

- [x] **Project Initialization**: Setup Standard Go Project Layout.
- [x] **Configuration Management**: Integrate `Viper` for multi-environment support.
- [x] **Logging System**: Implement structured logging with `Zap`.
- [x] **Database Design**: Design `BlockTrace` and `TransferEvent` models with Gorm.

---

## 📅 Phase 2: Core Scanner (核心扫描器)
**Status**: ✅ Completed (已完成)

- [x] **RPC Client**: Encapsulate `ethclient` with connection keep-alive.
- [x] **Block Iterator**: Implement block iteration logic from Start to End.
- [x] **Log Fetching**: Fetch raw logs via `FilterLogs`.

---

## 📅 Phase 3: Parser & Persistence (解析与存储)
**Status**: 🔄 In Progress (进行中)
**ETA**: 3 Days

- [ ] **Data Parsing**: Parse Hex logs into human-readable transfer events (From, To, Amount).
- [ ] **Persistence Layer**: Save valid transactions to MySQL using Gorm.
- [ ] **State Management**: Record the last scanned block to allow resume-from-break.

---

## 📅 Phase 4: Architecture Upgrade (架构升级 - 核心卖点) 🔥
**Status**: ⏳ Planned
**ETA**: 5 Days

- [ ] **Reorg Handling**: Implement block hash comparison to detect and handle chain forks (LIFO rollback).
- [ ] **Pipeline Concurrency**: Refactor to Producer-Consumer model using Go Channels.
- [ ] **Graceful Shutdown**: Ensure no data loss during service restart.

---

## 📅 Phase 5: DevOps & Delivery (部署与交付)
**Status**: ⏳ Planned

- [ ] **Dockerization**: Write `Dockerfile` and `docker-compose.yml`.
- [ ] **Documentation**: Complete `README.md` with architecture diagrams.
- [ ] **Monitoring**: Integrate Prometheus metrics (Scan Speed, RPC Latency).

---

## 📅 Phase 6: Kubernetes Orchestration (K8s 容器编排) ☸️
**Status**: ⏳ Planned

- [ ] **Deployment**: Define K8s Deployment & Service resources.
- [ ] **Config Management**: Migrate `config.yaml` to ConfigMap & Secrets.
- [ ] **Stateful Workloads**: Deploy MySQL & Redis using StatefulSet and PVC.
- [ ] **Health Checks**: Configure Liveness and Readiness probes.

---

## 📅 Phase 7: Microservices Evolution (微服务演进 - Kafka) 🚀
**Status**: ⏳ Planned (Advanced)
**Goal**: Decouple Scanner and Parser using Message Queue.

- [ ] **Kafka Integration**: Deploy Kafka & Zookeeper (or Kraft) via Docker.
- [ ] **Producer**: Scanner pushes raw logs to Kafka topic `chain-events`.
- [ ] **Consumer Group**: Parser consumes from Kafka (supports horizontal scaling).
- [ ] **Traffic Shaping**: Handle traffic spikes during historical sync.

---

## 📅 Phase 8: Database Migration (PostgreSQL 迁移) 🐘
**Status**: ⏳ Planned (Refactoring)
**Goal**: Migrate storage engine to support advanced JSON queries and high precision.

- [ ] **Driver Switch**: Switch Gorm driver from MySQL to PostgreSQL.
- [ ] **Data Migration**: Migrate existing data to PG.
- [ ] **JSONB Optimization**: Refactor `TransferEvent` to use PG `JSONB` for flexible storage.

---

## 📅 Phase 9: CI/CD & Observability (自动化与可观测性) 🛡️
**Status**: ⏳ Planned (Reliability)
**Goal**: Establish a safety net with automated testing and distributed tracing.

- [ ] **CI Pipeline**: Setup GitHub Actions for Linting (golangci-lint) and Unit Tests.
- [ ] **CD Pipeline**: Auto-build Docker image and push to Registry on git push.
- [ ] **Distributed Tracing**: Integrate OpenTelemetry (OTEL) + Jaeger to trace requests across Microservices.

---

## 🛠 Tech Stack (技术栈)
* **Language**: Golang 1.20+
* **Blockchain**: Go-Ethereum (Geth)
* **Database**: MySQL 8.0 -> **PostgreSQL 15** (Phase 8)
* **Message Queue**: **Kafka** (Phase 7)
* **Cache**: Redis 7.x
* **ORM**: Gorm
* **Config**: Viper
* **Observability**: Prometheus, Grafana, Jaeger (Phase 9)
* **Infrastructure**: Docker Compose & Kubernetes (K8s)