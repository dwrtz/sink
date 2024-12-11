# sink

Sink is a tool for generating AI prompts from codebases.

## Building

```sh
go build -o ./out/sink ./cmd/sink
```

## Usage

```sh
./out/sink generate . -o output.md -f "*.go" --tokens
```
