# Almira

An enterprise AI agent built in Go, one step at a time.

Currently: a Cobra CLI that sends one prompt to an OpenAI-compatible Chat
Completions endpoint and prints the answer.

Planned: agent loops, tools, ephemeral container sessions using a replaceable
runtime (containerd by default), and private tenant, organization, division,
and user scopes. Container execution and access isolation are not implemented yet.

See [concept and next steps](docs/concept.md) for the design direction.

## Run

Requires Go 1.26+, `OPENAI_API_KEY`, and a model available to your account.
Set credentials in your environment; `.env` files are not loaded automatically.

```sh
go run . --prompt "Hello, Almira" --model gpt-4.1-mini
```

`OPENAI_MODEL` supplies the default model. `OPENAI_BASE_URL` optionally selects
an OpenAI-compatible endpoint. Use `-p` / `-m` for short flags, or `--help`.

## Development

```sh
go test -timeout 30s ./...
go vet ./...
go build -o bin/almira .
```

Tests use local mock servers and require no API credentials.
