# Working on Almira

- Read `docs/concept.md` for the product direction, current implementation, and open decisions before planning the next increment.

- Build in small, explainable increments. The learning sequence is provider call,
  agent loop, then tools, followed by disposable execution and enterprise access.
- Use Go. Favor the standard library while it keeps the implementation readable.
- Keep the executable entry point in root `main.go`. Assemble Cobra commands in `cmd/` and define feature commands under `internal/<context>/controller/command/`. Use constructors and `RunE`; keep business rules in `entity` and `usecase`. Propagate the command context for cancellation.
- Apply domain-driven design: organize by bounded context; keep domain rules and
  vocabulary independent of transport, providers, and container runtimes.
- Follow Altair conventions: `entity`, `usecase`, `controller`, and `provider.go`.
  Keep use cases in `usecase` packages. Define ports where they are consumed.
  Implement model providers under `internal/provider/` and runtime details in adapters. Wire concrete implementations in each module's `provider.go`.
- Introduce entities, aggregates, repositories, and events only when real behavior
  requires them. Avoid empty abstractions and speculative frameworks.
- Start with OpenAI-format Chat Completions. Keep its JSON schema private to the
  adapter so other providers can be added without rewriting domain behavior.
- Persist authorized conversation history and agent/team configuration in the DB; execution processes and working files are disposable. Do not use host workspace directories for session persistence.
  The future runtime owns disposable container state and cleanup.
- Containerization and orchestration must be replaceable. The first runtime is
  direct containerd, with no Docker dependency.
- Tenant, organization, division, and user scope must be explicit when introduced.
  Isolation is deny-by-default; hierarchy does not imply content access.
- Never commit API credentials or private conversation data, or log credentials, prompts, or complete provider error bodies. Persist conversation content only through authorized conversation storage; the current single-call CLI has no such storage.
- For Go changes, run `gofmt`, `go test -timeout 30s ./...`, and `go vet ./...`.
  Use local HTTP servers for provider contract tests; live API tests are separate.
- Explain the request flow, why each boundary exists, and which guarantees are
  implemented versus planned. Keep the README current.
