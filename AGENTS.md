# Working on Almira

- Build in small, explainable increments. The learning sequence is provider call,
  agent loop, then tools, followed by disposable execution and enterprise access.
- Use Go. Favor the standard library while it keeps the implementation readable.
- Keep the executable entry point in root `main.go`. Define Cobra commands in
  `cmd/`, using constructors and `RunE`; keep business rules in the domain and
  application packages. Propagate the command context for cancellation.
- Apply domain-driven design: organize by bounded context; keep domain rules and
  vocabulary independent of transport, providers, and container runtimes.
- Keep use cases in application packages. Define ports where they are consumed.
  Implement provider and runtime details in infrastructure adapters. Wire concrete
  implementations in the entry point.
- Introduce entities, aggregates, repositories, and events only when real behavior
  requires them. Avoid empty abstractions and speculative frameworks.
- Start with OpenAI-format Chat Completions. Keep its JSON schema private to the
  adapter so other providers can be added without rewriting domain behavior.
- Session data and generated files must not use host workspace directories.
  The future runtime owns disposable container state and cleanup.
- Containerization and orchestration must be replaceable. The first runtime is
  direct containerd, with no Docker dependency.
- Tenant, organization, division, and user scope must be explicit when introduced.
  Isolation is deny-by-default; hierarchy does not imply content access.
- Never persist or log API credentials, prompts, or complete provider error bodies.
- For Go changes, run `gofmt`, `go test -timeout 30s ./...`, and `go vet ./...`.
  Use local HTTP servers for provider contract tests; live API tests are separate.
- Explain the request flow, why each boundary exists, and which guarantees are
  implemented versus planned. Keep the README current.
