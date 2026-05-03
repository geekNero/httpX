### Further tasks that can be done in http client
Here’s a focused set of feature ideas based on what’s already in your codebase (custom HTTP/1.1 parsing, chunked responses, proxying). I grouped them by payoff and learning
  value.


##### trie-based URL router, a parser for multipart form data, or a basic rate-limiting middleware

  Protocol & Parsing

  1. Support Connection: keep-alive and multiple requests per TCP connection (pipelining off initially).
  2. Parse Transfer-Encoding: chunked requests (you already emit chunked responses).
  3. Add header folding guardrails and size limits (max header bytes, max line length).
  4. Validate Content-Length vs actual body bytes; handle Content-Length: 0.
  5. Implement Expect: 100-continue for large uploads.

  Response & Status Coverage

  6. Implement HEAD and OPTIONS methods.
  7. Add 404, 405, 413, 414, 431, 500 and friends, plus Allow header for 405.
  8. Correct Trailer header spelling and trailer line endings in WriteChunkedBodyDone.
  9. Add Date header and basic Server header.

  Routing & Handlers

  10. Simple router with path params and method matching.
  11. Static file server with safe path traversal checks.
  12. Query string parsing and form decoding.
  13. Middleware chain (logging, recovery, request IDs, CORS).

  Concurrency & Robustness

  14. Per-connection read/write deadlines (idle timeout, header timeout, body timeout).
  15. Connection limits / backpressure (max in-flight conns, max requests per conn).
  16. Graceful shutdown that drains active connections.

  HTTP Compliance & Interop

  17. Proper Host/absolute-form support for proxy requests.
  18. Support Range requests for partial content.
  19. gzip/deflate response compression when client accepts it.

  Observability

  20. Structured access logs (method, path, status, bytes, latency).
  21. Simple metrics endpoint (in-flight, total requests, errors, p95).
  22. Trace IDs propagated via headers.

  Testing & Tooling

  23. Property tests for header parsing and request line parsing.
  24. Golden tests for raw HTTP input/output fixtures.
  25. Benchmarks for parser and chunked writer.

  If you want, tell me which direction you care about most (protocol correctness, routing/features, or robustness), and I’ll map that into a small, concrete roadmap and propose
  specific changes in files like internal/request/request.go, internal/response/response.go, and internal/server/server.go.
