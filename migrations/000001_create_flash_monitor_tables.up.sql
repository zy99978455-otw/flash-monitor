-- 1. 创建区块足迹表 (Block Trace)
CREATE TABLE IF NOT EXISTS block_traces (
    id BIGSERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL UNIQUE,
    block_hash VARCHAR(66) NOT NULL,
    parent_hash VARCHAR(66) NOT NULL,
    scan_time TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- 2. 创建转账事件表 (Transfer Event)
CREATE TABLE IF NOT EXISTS transfer_events (
   id BIGSERIAL PRIMARY KEY,
   tx_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    block_number BIGINT NOT NULL,
    block_hash VARCHAR(66) NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    -- Web3 的 uint256 极大，PostgreSQL 的 NUMERIC 类型完美支持无限精度
    amount NUMERIC NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
    );

-- 添加联合唯一索引，防止同一笔交易的同一个 log 被重复插入
ALTER TABLE transfer_events ADD CONSTRAINT tx_log_unique UNIQUE (tx_hash, log_index);

-- 为经常查询的字段添加普通索引，加速 API 读取
CREATE INDEX IF NOT EXISTS idx_transfer_events_block_number ON transfer_events(block_number);
CREATE INDEX IF NOT EXISTS idx_transfer_events_from_address ON transfer_events(from_address);
CREATE INDEX IF NOT EXISTS idx_transfer_events_to_address ON transfer_events(to_address);