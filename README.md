# 📦 sndbx

**`sndbx`** is a CLI utility for spinning up quick Docker-based sandboxes either from a local Dockerfile or a remote image. It simplifies container management for rapid testing, development, or experimentation with custom Docker environments.

## 🚀 Features

- 🔨 Build a Docker image from a local `Dockerfile`
- 📥 Pull and run containers from public Docker images
- 🧹 Automatically remove containers and images after use
- 🔌 Expose custom ports
- 📁 Mount current working directory as `/app` inside the container
- 🧪 Simple interactive shell access to the container environment

---

## 📦 Installation

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Go](https://golang.org/dl/) 1.18 or above

### Build and Install

```bash
make install
```

This will:

- Build the binary to `bin/sndbx`
- Copy it to `/bin/sndbx` for global CLI access

To uninstall:

```bash
make uninstall
```

---

## 🔧 Usage

```bash
sndbx init [flags]
```

### Available Flags

| Flag           | Alias | Description                            | Default        |
|----------------|-------|----------------------------------------|----------------|
| `--build`      | `-b`  | Build from a local Dockerfile          | `false`        |
| `--context`    | `-c`  | Path to Dockerfile or Image name       | `""`           |
| `--remove`     | `--rm`| Remove the sandbox after exit          | `false`        |
| `--ports`      | `-p`  | List of ports to expose                | `[]`           |

---

## 🧪 Examples

### 1. Run sandbox from an image (e.g., Ubuntu)

```bash
sndbx init --context ubuntu:latest
```

### 2. Build and run from local Dockerfile

```bash
sndbx init --build --context Dockerfile
```

### 3. Automatically remove the container and image after session

```bash
sndbx init --build --context Dockerfile --remove
```

### 4. Expose ports

```bash
sndbx init --context ubuntu:latest --ports 8080,3000
```

---

## 🛠️ Development

### Run Locally

```bash
go run cmd/main.go
```

### Build Binary

```bash
make build
```

### Run Tests

```bash
make test
```

### Dev Sandbox Test

```bash
make dev-test
```

### Reset Dev Environment

```bash
make dev-reset-env
```

---

## 📁 Project Structure

```
.
├── cmd/
│   └── main.go          # CLI entry point
├── internal/
│   ├── sandbox/         # Sandbox lifecycle management
│   ├── check_context/   # Context helper for Dockerfile check
│   └── utils/           # Common utilities (e.g., stream handling)
├── test/                # Test scenarios
├── Makefile             # Common commands
```

---

## 📦 Available Base Images

`sndbx` supports any image available on Docker Hub. Some common options:

- `ubuntu:latest`
- `debian:latest`
- `alpine:latest`
- `centos:latest`
- `fedora:latest`

---

## 🧠 Internals

- Uses Docker Go SDK for managing containers and images.
- Automatically mounts the current working directory as `/app` inside the container.
- Uses host user IDs for permission consistency.
- Gracefully cleans up temporary containers/images if `--remove` is passed.

---

## 🫶 Contributing

Pull requests and feedback are welcome! Feel free to open issues for bugs or feature requests.
