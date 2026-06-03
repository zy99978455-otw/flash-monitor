# 🐋 FlashMonitor: High-Performance Web3 Indexer

![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15.0+-4169E1?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)
![License](https://img.shields.io/badge/License-MIT-green.svg)

<img width="2560" height="1440" alt="image" src="https://github.com/user-attachments/assets/a04a40e3-99c0-4ef1-8ca3-6fa7e82a7e27" />

FlashMonitor is a lightweight, production-grade EVM on-chain data indexing and distribution engine written in Go. Embracing the "Let's go further" geek philosophy, the system focuses on solving core pain points in Web3 data indexing: **data consistency (chain reorgs)**, **long-lived connection performance**, and **Single Point of Failure (SPOF)**.

## 🎯 Core Features

### 1. Millisecond-Level Global Graceful Shutdown
Say goodbye to dangling connections and dirty data. By taking over the lifecycle via a global `context.Context`, the engine perfectly executes the following within **5 milliseconds** of receiving an OS termination signal (SIGINT/SIGTERM):
* Intercepts and safely rolls back ongoing PostgreSQL atomic transactions.
* Severs and cleans up all active SSE frontend long-lived connections.
* Blocks and cancels underlying RPC network requests, achieving a zero-memory-leak safe exit.

### 2. Chain Reorg Immunity Mechanism
Mainnet forks shouldn't turn into data disasters. The engine senses chain state rollbacks in real-time through a `BlockHash` traceability comparison algorithm. Coupled with database atomic transaction isolation, it achieves automatic discovery and safe eviction of "phantom data," ensuring the absolute purity of business-layer data.

### 3. Multi-Node RPC Fallback Manager
Break the Single Point of Failure (SPOF) curse. Built with a `NodeManager` connection pool based on read-write locks (`sync.RWMutex`) and adaptive block height comparison, it implements real-time heartbeat monitoring of node health. When the primary node is rate-limited (HTTP 429) or unresponsive, it seamlessly routes to backup nodes and executes exponential backoff retries.

### 4. Lock-Free Real-Time SSE Broker
Ditching bulky WebSockets, this system utilizes a Server-Sent Events (SSE) broadcast center built on Go's native `Channel` and `select` mechanisms. It leverages the non-blocking send characteristic of channels to automatically fuse (disconnect) slow clients, supporting ultra-fast, one-way pushing for tens of thousands of concurrent connections on a single machine.

---

## 🚀 Quick Start

Thanks to complete Docker container orchestration, you can launch the system with one click without installing any database or Go environment locally. The system is aggressively memory-tuned (GOMEMLIMIT constraints and Postgres buffer tuning) for lightweight cloud servers with 1GB RAM (e.g., RackNerd VPS).

### 1. Environment Setup
Clone the repository and prepare environment variables:
```bash
git clone [https://github.com/YourUsername/flash-monitor.git](https://github.com/YourUsername/flash-monitor.git)
cd flash-monitor
cp .env.example .env
Please insert your actual ETH_RPC_MAIN address in the .env file
```

### 2. One-Click Launch
```bash
docker compose up -d --build
```

### 3. Real-Time Observation
* Backend Logs: docker compose logs -f api
* Frontend Whale Dashboard: Access http://localhost:4010 (or your server's IP) via browser to connect to the SSE real-time stream.

## Roadmap
To maintain the extreme lightweight and pure nature of V1, the following features have architectural placeholders and are planned for future versions:

[x] V1.0: ETH mainnet monitoring, chain reorg defense, lock-free SSE push, Docker memory tuning.

[ ] V2.0: Introduce Redis for historical data caching and distributed cluster locks to support multi-node horizontal scaling.

[ ] V2.1: Support dynamic configuration for concurrent monitoring of multiple EVM-compatible chains (BSC / Base / Arbitrum).

[ ] V3.0: Integrate the Prometheus + Grafana ecosystem for visual monitoring of system metrics (RPC latency, memory reclamation rate, sync watermarks).