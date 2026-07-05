# Contributing to minigui

Thanks for your interest. minigui is a tiny immediate-mode GUI for Ebitengine,
built to be shared across small programs that need basic input without a heavy
widget framework. "Tiny" is not a phase it will grow out of; it is the design.

## Immediate mode is the contract

Every frame: build an `Input`, `Begin`, run the widgets, `End`, `Render`.
Widgets read plain input and append draw commands; state lives in the caller.

- No retained widget tree, no hidden per-widget state machines, no callback
  soup. If a widget needs memory across frames, it is kept small, explicit,
  and owned by the context.
- `Input` stays a plain struct. That is what makes widget logic testable
  without a window, and headless testability is a feature we do not give back.
- New widgets come with headless tests (drive `Input`, assert state and draw
  commands), not screenshots.

## YAGNI: a widget needs a real caller

The bar for a new widget or option is a real program that needs it; the family
apps (NeoFrame, kutta, the linefire editors) are where that demand comes from.
No speculative widgets, no config knob nobody sets, no theming hook without a
theme that uses it. The shortest change that solves the problem wins.

If a widget can be composed from existing ones in the caller, that is the
answer; not everything deserves to live in the library.

## Dependencies are frozen

Ebitengine and [native](https://github.com/crgimenes/native) (cgo-free
clipboard); everything else is the standard library. No new dependencies, and
nothing that brings cgo with it: `CGO_ENABLED=0` builds must stay green.
Fonts and styling stay pluggable (faces and theming are injection points);
the library does not embed assets beyond the minimal default.

## Public API stability

minigui is a public library used by several programs. Exported API does not get
removed; if something turns out wrong, document it and supersede it. Signature
changes are still possible pre-1.0, but they ripple into every family app, so
they need a reason.

## Code style

`gofmt`, US English everywhere. No inline `if` init; assign on its own line,
then `if`. No `else` after a terminal branch; return early. Prefer `switch`
over long `else if` chains. Comments explain why, not what. For user-facing
text handling, think in runes, not bytes.

## Before you open a PR

```sh
go fix ./...
gofmt -l .        # must print nothing
go vet ./...
golangci-lint run ./...
go test -timeout 30s -count 1 ./...
```

The widget tests run headless, so the suite passes on a machine with no
display; keep it that way. If a change is visual (layout, theming, text
rendering), run one of the family apps and look at it; the screen is the final
reviewer.

## Proposing a change

The default branch is `trunk`. Fixes and small improvements can go straight to
a PR. For a new widget or an API change, open an issue first and name the real
program that needs it; that conversation is short and saves long ones later.
