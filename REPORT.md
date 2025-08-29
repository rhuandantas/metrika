# Architecture Decision Log

## 1. Persistence Layer

**Decision:**
Use SQLite for local persistence with batched writes and in-memory caching.

**Rationale:**
SQLite is lightweight, easy to set up, and sufficient for local storage needs. Batched writes reduce I/O overhead, and in-memory caching improves read performance.

**Libraries/Tech:**
- SQLite
---

## 2. Data Handling

**Decision:**
Use streaming JSON writers and chunked writes for large datasets. Initially used JSONL, but switched to Lumberjack for log rotation and compression.

**Rationale:**
Writing large JSON files at once is inefficient and resource-intensive. Streaming and chunking improve performance and reliability. Lumberjack enables automatic log rotation and compression, addressing file size and management concerns.

**Libraries/Tech:**
- [Lumberjack](https://github.com/natefinch/lumberjack)
- [Zerolog](https://github.com/rs/zerolog)
---

## 3. Periodic Checkpointing

**Decision:**  
Implement periodic checkpointing to persist progress and enable recovery after crashes.

**Rationale:**  
Regularly saving state prevents data loss and supports fault tolerance.

**Libraries/Tech:**  
- Go standard library (`time.Ticker`, goroutines)
- Custom metrics repository
---

## 4. Concurrency and Safety

**Decision:**  
Use mutexes to protect shared state and avoid race conditions.

**Rationale:**  
Concurrent access to in-memory metrics requires synchronization for correctness.

**Libraries/Tech:**  
- Go standard library (`sync.Mutex`, `sync.RWMutex`)
---

## 5. Testing Frameworks

**Decision:**  
Adopt Ginkgo and Gomega for BDD-style testing, and GoMock for mocking interfaces.

**Rationale:**  
These libraries provide expressive, maintainable tests and support mocking dependencies.

**Libraries/Tech:**  
- [Ginkgo](https://github.com/onsi/ginkgo)
- [Gomega](https://github.com/onsi/gomega)
- [GoMock](https://github.com/golang/mock/gomock)
---

## 6. Logging

**Decision:**  
Use Zerolog for structured, performant logging.

**Rationale:**  
Zerolog offers fast, leveled logging with JSON output.

**Libraries/Tech:**  
- [Zerolog](https://github.com/rs/zerolog)
---

## 7. API Client

**Decision:**  
Abstract SmartBlox API interactions behind a client interface.

**Rationale:**  
Encapsulation enables easier testing and future changes.

**Libraries/Tech:**  
- Custom `smartblox.Client` interface
---

# Opportunities
- Use a robust cache like Redis for better strategies, performance and scalability.
- Implement a more sophisticated batching mechanism that adapts to load and optimizes batch sizes dynamically.
- Send metrics to another service or queue for asynchronous processing, decoupling ingestion from downstream metric handling and enhancing scalability.
- Introduce configuration files for flexible runtime settings
- Support pluggable storage backends (e.g., PostgreSQL, cloud databases)
- Implement a mechanism to control the number of requests to the SmartBlox external API to avoid hitting rate limits (e.g., request throttling, token bucket, or leaky bucket algorithms)
- Add support for distributed tracing to track requests across services
- Better error handling and retry logic for transient failures when interacting with the SmartBlox API or database