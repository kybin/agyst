# agyst

A lightweight wrapper for `agy` to customize things `agy` itself doesn't let you configure.

It makes the agent feel calmer and more static so it's less distracting.

Built with an AI agent.

## Install

```bash
go install github.com/kybin/agyst@latest
```

Make sure Go's `bin` directory (`$(go env GOPATH)/bin`, usually `~/go/bin`) is in your `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

## Use `agyst` when running `agy`

Add this line to your `~/.bashrc` or `~/.zshrc`:

```bash
alias agy='agyst'
```

Then reload your shell:

```bash
source ~/.bashrc
```
