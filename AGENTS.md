# Fugo — agent guide

**v0.1.0** · Local Server-Driven UI framework for desktop: app logic, state, and routing in
**Go**; a precompiled **Flutter** binary renders over gRPC bidirectional streaming (TCP on
Windows, UDS elsewhere). Go is the single source of truth.

The engine, widget API (`fg/`), transport, supervisor, CLI, and Flutter client are all
**implemented and run end-to-end**. For the full architecture, data flow, and conventions see
[`CLAUDE.md`](./CLAUDE.md) — it is the canonical, up-to-date guide; this file is a short index.

## Layout

- `fg/` — declarative widget API (`fg.Text`, `fg.Button`, `fg.Container`, …; prefix-free
  constructors returning `*fg.XxxWidget`) plus the `Theme` system (`fg.DarkTheme`/`LightTheme`,
  `fg.UseTheme`).
- `style/` — styling primitives (`Color`, `EdgeInsets`, `TextStyle`, `Border`).
- `engine/` — `Diff` (keyed patches), `Reconciler`, `Scheduler` (16ms/60fps coalescing).
- `transport/` — gRPC server (UDS/TCP), health, keepalive.
- `supervisor/` — spawns/monitors the Flutter subprocess.
- `cmd/fugo/` — CLI (`init`, `run` + `--watch`, `build`, `doctor`); `cmd/fugo-spike/` — demo.
- `flutter_client/` — Dart render client.

## Commands

```sh
make test          # go test ./... -count=1 -race -shuffle=on -v
make lint          # golangci-lint run --timeout 10m ./...
make vet           # go vet ./...
make build         # build bin/fugo with version/commit/date ldflags
make install       # install Lefthook git hooks
make install-tools # (re)install gofumpt, staticcheck, golangci-lint built with local Go
make run           # build spike + spawn Flutter client (OS-aware)
make proto         # regenerate Go + Dart protobuf from transport/proto/fugo/v1/fugo.proto
```

## Conventions

- **Formatter**: `gofumpt` (not `gofmt`). Pre-commit auto-formats and re-stages.
- **Linters**: 80+ via `.golangci.yml`; `staticcheck` also enforced. If golangci-lint refuses to
  run with a Go-version mismatch, run `make install-tools`.
- **Wire format**: standard `google.golang.org/protobuf` — per-widget `*Props` marshaled as
  nested protobuf inside `WidgetNode.props`. **Not** FlatBuffers/vtprotobuf (README/SPEC/ROADMAP
  describe FlatBuffers as a design aspiration; trust the code).
- **Adding a widget** touches four places: the proto (`WidgetType` + `*Props`), `make proto`, the
  Go widget in `fg/`, and the Dart builder in `flutter_client/lib/registry.dart`.
- **Release**: use `make release` — never hand-edit `VERSION`/`CHANGELOG.md`.

## Git hooks (Lefthook)

**Pre-commit** (parallel, auto-fixes staged files): golangci-lint --fast --fix → go vet →
gofumpt -w (re-staged) → go mod tidy.

**Pre-push** (sequential, full suite): version/CHANGELOG sync → golangci-lint → go vet →
staticcheck → go build → go test -race -shuffle=on → go mod tidy.

## CI (push/PR to `main`)

Parallel jobs: golangci-lint + staticcheck, go vet, go build, go test -race -shuffle=on, gofumpt
format check.
