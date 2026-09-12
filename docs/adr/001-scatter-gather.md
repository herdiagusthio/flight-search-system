# ADR 001: Scatter-Gather Pattern for Flight Aggregation

**Date:** 2026-09-12
**Status:** Accepted

## Context
The system must query multiple external flight providers. These providers have varying latencies, reliability, and data formats. A sequential request pattern would result in unacceptable response times for the user.

## Decision
We implemented a **Scatter-Gather** pattern utilizing Go's concurrency primitives:
- **Scatter**: Each provider is queried in a separate goroutine.
- **Gather**: Results are collected via a buffered channel.
- **Timeouts**: A dual-timeout strategy is used: a global timeout for the entire request and a per-provider timeout to prevent a single slow API from blocking the response.
- **Partial Success**: The system is designed to return a valid response as long as at least one provider succeeds, rather than failing the entire request.

## Trade-offs
- **Pros**: Maximum possible throughput, minimal response latency, and high resilience to partial outages.
- **Cons**: Increased complexity in goroutine management and the need for a robust panic-recovery mechanism per worker.

## Final Verdict
Given the unreliable nature of third-party flight APIs, the Scatter-Gather pattern is the only way to guarantee a responsive and reliable user experience.
