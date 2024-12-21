# Sink

**Sink** is a command-line tool designed to help you generate structured, language model-ready prompts from your codebase. By analyzing your repository, filtering files based on patterns or `gitignore` rules, stripping comments, adding line numbers, and even estimating token usage, Sink streamlines the process of turning your code into rich, descriptive AI prompts.

## Features

- **Language-Agnostic**: Works with a variety of programming languages, leveraging extensible syntax mappings.
- **Flexible File Selection**: Include or exclude files and/or directories based on glob patterns (including double stars `**`), `gitignore` rules, or case sensitivity.
- **File Watching**: Monitor a directory (and its subdirectories) for changes and automatically regenerate the output when files are added, modified, or removed.
- **Configurable via YAML**: A `sink-config.yaml` file allows for easy configuration of defaults, such as encoding, price estimation, and filter patterns.

## Installation

You will need [Go](https://go.dev/dl/) installed (version 1.22+ recommended).

### Step 1: Build the binary

```sh
go build -o ./out/sink ./cmd/sink
```

This will produce a `sink` binary in the `./out` directory.

### Step 2: Make the binary accessible system-wide

```sh
sudo mv ./out/sink /usr/local/bin/sink
```

Now you can run `sink` from anywhere on your system.

## Usage

### Basic generation:

```sh
sink generate . -o output.md
```

This command:
- Scans the current directory (`.`)
- Generates a Markdown file (`output.md`) that includes your code files
- Uses filters and configurations from `sink-config.yaml`

### Filtering Files:

```sh
sink generate . -o output.md -f "*.go" --tokens
```

This command:
- Includes only `.go` files
- Counts and displays token usage in the output

You can also use double-star (`**`) patterns for recursive matching:
```sh
sink generate . -o output.md -f "myproj/**/*.py"
```
This includes all Python files under the `myproj` directory (no matter how many nested subdirectories exist).

### Watching for changes:

```sh
sink watch . -o output.md
```

This command:
- Monitors the current directory (`.`) for file changes
- Automatically regenerates the Markdown output (`output.md`) whenever files are created, modified, or removed
- Applies the same filtering rules and configurations from `sink-config.yaml`

You can also specify additional flags, for example:
```sh
sink watch . -o output.md -f "*.go,*.md"
```

- `-f "*.go,*.md"` includes only Go and Markdown files

Press **Ctrl+C** to stop watching.

## Configuration

Sink looks for a `sink-config.yaml` file for default configurations. In this file, you can specify:

- Filters and exclude patterns
- Template paths for custom Markdown formatting
- Other options like token encodings, providers, and models for price estimation

See the [example config](./examples/sink-config.yaml) for more details.

## Acknowledgements

Sink was inspired by the work done on [code2prompt](https://github.com/raphaelmansuy/code2prompt).
