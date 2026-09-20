# Almira

An enterprise AI agent, built in Go one small step at a time.

## Step 1: one prompt, one answer

This increment calls an OpenAI-format Chat Completions endpoint and prints the
answer. The CLI uses Cobra, with a root `main.go` entry point. The OpenAI adapter
uses Go's standard library, so you can read the actual HTTP request without
learning an SDK first.

The CLI is a development harness. Agent loops, tools, tenant authorization, and
container execution are future increments. This is not yet a deployed enterprise
service or an isolated session runtime.

## Run

Requires Go 1.26 or later, an OpenAI API key, and access to a model that supports
Chat Completions. Set the key through your environment or secret manager. Almira
does not read a `.env` file or accept keys as command-line arguments.

PowerShell 7 (the key prompt is hidden and does not put the key in shell history):

```powershell
$env:OPENAI_API_KEY = Read-Host 'OpenAI API key' -MaskInput
$env:OPENAI_MODEL = 'gpt-4.1-mini'
go run . --prompt 'Introduce yourself as Almira in one sentence.'
```

Bash:

```bash
read -r -s -p 'OpenAI API key: ' OPENAI_API_KEY
export OPENAI_API_KEY
export OPENAI_MODEL=gpt-4.1-mini
go run . --prompt 'Introduce yourself as Almira in one sentence.'
```

The model is deliberately an explicit setting. The example uses
[GPT-4.1 mini](https://developers.openai.com/api/docs/models/gpt-4.1-mini);
choose a model your account can access. `--model` (or `-m`) overrides `OPENAI_MODEL`.
Use `--prompt` (or `-p`) for input, and `go run . --help` for command help.
Cobra uses double dashes for long flags; the original `-prompt` and `-model`
spellings are replaced by `--prompt` and `--model`.

| Setting | Purpose |
| --- | --- |
| `OPENAI_API_KEY` | Required credential, held in process memory |
| `OPENAI_MODEL` | Model ID; can instead be supplied with `--model` |
| `OPENAI_BASE_URL` | Optional API prefix; defaults to `https://api.openai.com/v1` |

For an OpenAI-compatible endpoint, set its API prefix (for example,
`https://provider.example/v1`). Almira appends `/chat/completions`. Use that
provider's credential and model ID. Compatibility must be checked per provider;
other protocols will need their own adapters. HTTPS is required except for HTTP
on loopback during local development.

Requests time out after 60 seconds. Ctrl+C cancels an in-flight request. Errors
go to stderr and produce a nonzero exit status. A truncated answer, refusal, or
tool request is an error in this first text-only increment. There are no automatic
retries, so a failed request cannot silently cause another model call.

## Read the code in this order

1. `internal/agent/domain/prompt.go`: a `Prompt` is immutable text that must contain
   something besides whitespace. Formatting is preserved. This is a value object,
   not an entity: it has no identity or lifecycle of its own.
2. `internal/agent/application/ask.go`: `Ask.Execute` creates the prompt and asks a
   `TextGenerator` for an answer. The interface is a port owned by the application.
3. `internal/agent/infrastructure/openai/client.go`: the adapter translates that
   call into HTTP and private JSON types, then translates the response back into
   text. Provider details stay here.
4. `cmd/root.go`: constructs the Cobra root command, defines flags, selects the
   adapter, connects it to `Ask`, and prints the result. `RunE` returns errors to
   the entry point; `command.Context()` carries cancellation to the provider.
5. `main.go`: creates the interrupt-aware context, executes the command, and sets
   the exit status. This stays small as we add more commands.

```text
main.go -> Cobra command -> Ask -> TextGenerator port -> OpenAI adapter -> API
                            |
                            +-> Prompt value object
```

The application imports the domain. Infrastructure imports the domain and
implicitly implements the application interface. Only the CLI chooses a concrete
provider. There is no need for an entity, aggregate, repository, or database yet.
This is the initial slice of the agent bounded context; more domain behavior will
appear when we introduce sessions and tools.

## Privacy and ephemeral execution

This executable creates no conversation files or workspace directories. Prompts
and answers live in memory; the answer is written to stdout. Shell history,
terminal capture, or explicit output redirection are controlled by the caller.

The request sets `store: false`. This disables optional completion storage; it
does not promise zero provider-side retention. See OpenAI's
[Chat Completions reference](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create)
and [data controls](https://developers.openai.com/api/docs/guides/your-data).
Provider error bodies are not printed because they can contain private input or
credentials. Response bodies are limited to 1 MiB, and redirects are rejected.

The eventual session runtime will use disposable containers, with containerd as
the first adapter and no Docker dependency. Session files belong in the container's
writable filesystem, without host workspace bind mounts. A later runtime increment
must own cleanup of tasks, containers, and writable snapshots, including crash
recovery. Containerd itself uses host storage for images and snapshots; ephemeral
means the session state is discarded, not that storage never exists on the host.

## Verify without API credentials

```sh
go test -timeout 30s ./...
go vet ./...
go build -o bin/almira .
```

Tests use local HTTP servers and a fake provider. They cover the complete
application-to-adapter path, request headers and JSON, explicit `store: false`,
input validation, cancellation, redirects, provider failures, and incomplete or
oversized responses. Command tests also cover help without credentials, flag
validation, model overrides, output routing, and context propagation. They do not
call OpenAI or consume API credits.

## Next increments

1. **Done:** one provider call behind a small application port.
2. **Agent loop:** introduce in-memory session state, messages, termination rules,
   and explicit iteration limits.
3. **Tools:** introduce tool calls, results, permissions, and a tool registry.
4. **Execution:** add a runtime port and direct containerd adapter for disposable
   sessions; keep orchestration choices outside the agent domain.
5. **Enterprise access:** implement tenant, organization, division, and user
   boundaries before exposing the service to multiple users.

The target privacy rule is deny-by-default isolation at every scope. Parent
membership or an administrator role must not silently grant access to child or
sibling session content. Exact membership and sharing rules will be modeled when
we implement access control. No tenant isolation is implemented in this CLI.
