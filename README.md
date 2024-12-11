# Sink

**Sink** is a command-line tool designed to help you generate structured, language model-ready prompts from your codebase. By analyzing your repository, filtering files based on patterns or `gitignore` rules, stripping comments, adding line numbers, and even estimating token usage, Sink streamlines the process of turning your code into rich, descriptive AI prompts.

## Features

- **Language-Agnostic**: Works with a variety of programming languages, leveraging extensible syntax mappings.
- **Flexible File Selection**: Include or exclude files and/or directories based on glob patterns, `gitignore` rules, or case sensitivity.
- **Configurable via YAML**: A `sink-config.yaml` file allows for easy configuration of defaults, such as encoding, price estimation, and filter patterns.

## Installation

You will need [Go](https://go.dev/dl/) installed (version 1.22+ recommended).

```sh
go build -o ./out/sink ./cmd/sink
```

This will produce a `sink` binary in the `./out` directory.

## Usage

**Basic generation:**

```sh
./out/sink generate . -o output.md
```

This command:
- Scans the current directory (`.`)
- Generates a Markdown file (`output.md`) that includes your code files
- Uses filters and configurations from `sink-config.yaml`

**Filtering Files:**

```sh
./out/sink generate . -o output.md -f "*.go" --tokens
```

This command:
- Includes only `.go` files
- Counts and displays token usage in the output

## Configuration

Sink looks for a `sink-config.yaml` file for default configurations. In this file, you can specify:

- Filters and exclude patterns
- Template paths for custom Markdown formatting

See the [example config](./examples/sink-config.yaml) for more details.



## Acknowledgements

Sink was inspired by the work done on [code2prompt](https://github.com/raphaelmansuy/code2prompt).
