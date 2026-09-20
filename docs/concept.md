# Almira: concept and continuation notes

Recorded 2026-09-20. This document preserves the product direction and planning
conversation. Planned behavior below is not implemented unless explicitly stated.

## Product direction

Almira is an enterprise AI agent built in Go. Build it in small, explainable
increments so the code and design can be understood as they evolve.

Requirements established so far:

- Multi-agent by default, with agent configuration stored in a database rather
  than configuration files.
- Support multiple tenants, organizations, divisions, and users. Their private
  data must remain isolated; hierarchy alone does not grant access.
- Execution processes are ephemeral. Session working files belong inside
  disposable containers, not host workspace directories.
- Users can continue a chat and load a prior conversation. Ephemeral execution
  therefore does not mean deleting the conversation history.
- Containerization and orchestration are replaceable. The default execution
  adapter will use containerd directly, without a Docker dependency.
- Start with the OpenAI API format, with boundaries that allow other model
  providers later.
- Apply DDD principles using familiar Go conventions: root `main.go`, Cobra,
  feature modules with `entity`, `usecase`, `controller`, and explicit wiring in
  `provider.go`. Keep the README slim; detailed design belongs here.

## Separate persistence from execution

Use precise terms rather than treating every lifetime as one "session":

| Concept | Intended lifetime |
| --- | --- |
| Agent/team definitions and instruction content | Persistent in DB |
| Conversation and its ownership/history | Persistent and resumable |
| Execution run | Disposable process/container |
| Agent run | One agent's work within an execution |
| Working files | Disposable unless explicitly saved as artifacts |

The database is the durable source of configuration and conversation state.
Containers hold temporary execution state. Persistence must not depend on a
container surviving or a host workspace being retained. Containerd itself still
uses host storage for images and snapshots; ephemeral means cleanup of disposable
execution state, not the absence of physical storage.

## Instructions and configuration

Agent execution needs instruction content equivalent to `AGENTS.md` and `SOUL.md`.
With DB-backed configuration, these names describe instruction roles rather than
requiring configuration files as the source of truth:

- `AGENTS.md`: operating instructions, tool conventions, and working rules.
- `SOUL.md`: agent identity, tone, and interaction style.

Proposed design: resolve an authorized, versioned instruction bundle at startup.
Inject it into memory initially; optionally materialize read-only copies inside
the container when file access is needed. Do not discover these instructions by
searching a host working directory.

The repository's own `AGENTS.md` guides developers working on Almira. It is
separate from the instruction content configured for agents using the product.

Instructions cannot grant permissions. Go application code and the runtime must
enforce access, tool capabilities, and sandbox restrictions. Editing instructions
during a run must not silently change the durable configuration.

## Multi-agent model

Proposed domain concepts:

- **Agent definition:** name, instructions, model settings, tool permissions,
  and permitted delegation targets.
- **Team definition:** entry agent and participating agents.
- **Conversation:** persistent interaction, authorized ownership, and a reference
  to the configuration used for that interaction.
- **Execution run:** disposable execution that processes conversation work.
- **Agent run:** an individual agent's task and execution state.

Proposed behavior: the entry agent delegates explicit tasks to specialists and
receives results. Agents have separate contexts; delegation must not implicitly
share all conversation data. Enforce delegation permissions, iteration limits,
and concurrency limits outside model instructions. Team membership does not
bypass tenant, organization, division, or user isolation.

A first example can use one entry agent and one specialist. Multi-agent support
should be built into the model without requiring every request to fan out.

## Continuing a conversation

Proposed continuation flow:

1. Authorize access to the requested conversation.
2. Load its recorded history and configuration snapshot.
3. Start fresh execution with the relevant agent contexts.
4. Process the next user message and any permitted delegations.
5. Persist results and dispose of execution state.

Persist the records needed to reconstruct context: user/assistant messages,
delegations, agent messages, and tool calls/results, subject to access and
visibility rules. Durable conversation storage is intentional; copying private
payloads into diagnostic logs is not.

Proposed default: pin a conversation to a configuration version so continuing
it does not silently change agent behavior. Applying newer instructions should
be explicit. Exact snapshot/version semantics still need agreement.

Continuing a completed chat and recovering interrupted work are separate
features. Recovery needs checkpoints and safeguards against repeating tool side
effects. Loading history must not automatically replay previously executed tools.
Restoring a conversation does not restore the old container filesystem; durable
artifacts and restoration rules are separate future work.

## Implemented today

- Root `main.go` and Cobra CLI: one prompt, one model response.
- Prompt validation in `internal/agent/entity`.
- `Ask` use case and consumer-owned `TextGenerator` port in
  `internal/agent/usecase`.
- Command controller under `internal/agent/controller/command`.
- Explicit module wiring in `internal/agent/provider.go`.
- Standard-library HTTP adapter under `internal/provider/openai` using the
  OpenAI Chat Completions format.
- Environment-based provider settings for this development CLI. This is temporary
  bootstrap configuration, not the planned DB-backed agent configuration system.
- Mock-server tests covering commands, use cases, and provider behavior.

Not implemented: database storage, conversation resume, agent/team definitions,
conversation or autonomous agent loops, delegation, tools, containerd execution,
or enterprise access control. The current CLI is not an isolated multi-user
service. No live provider test has been performed in this development session.

## Suggested next increment

The original learning sequence was provider call, agent loop, then tools. The
new persistence requirements add a small foundation before the loop:

1. Agree on the minimal agent/team, conversation, message, and run model, plus
   ownership and configuration-version semantics.
2. Define consumer-owned repository ports and choose a database implementation
   for configuration and conversations. Do not leave the final configuration
   system file-backed.
3. Load an instruction bundle and persist/restore conversation messages around
   an in-memory conversation loop.
4. Add bounded delegation using an entry agent and one specialist.
5. Add tool contracts and execution limits; implement sandboxing before enabling
   shell commands or file execution.
6. Add direct containerd execution and lifecycle cleanup, then interruption
   recovery and durable artifacts as separate increments.

This ordering is a proposal, not authorization to implement all stages at once.
Scope and authorization constraints must shape storage from the beginning and
be enforced before exposing the service to multiple users.

## Decisions still open

- Database engine, schema, migration tooling, and repository implementation.
- Ownership and explicit sharing rules across the scope hierarchy.
- How identity is authenticated and authorization is supplied to each use case.
- Instruction precedence, team configuration format, and version migration UX.
- Persistence granularity, retention/deletion policies, and agent-message visibility.
- Delegation scheduling and budgets; whether agents share a container or receive
  separate containers.
- Checkpoint/recovery semantics and durable artifact storage.

## Continuing development on another device

Read this document and the repository `AGENTS.md`, then inspect the existing
single-call path before changing it. Continue in one reviewable increment at a
time and explain the request flow. No production database or runtime has been
chosen; do not infer one from the current CLI.

For Go changes, run `gofmt`, `go test -timeout 30s ./...`, and `go vet ./...`.
Tests use local mocks and do not require OpenAI credentials. Supply your own
provider credentials through the environment for a live CLI run; never commit
credentials or private conversation data.
