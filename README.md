README

# gorefresh - the missing `go run --watch`

[![Go Report Card](https://goreportcard.com/badge/github.com/draganm/gorefresh)](https://goreportcard.com/report/github.com/draganm/gorefresh)
[![GoDoc](https://pkg.go.dev/badge/github.com/draganm/gorefresh)](https://pkg.go.dev/github.com/draganm/gorefresh)
![License](https://img.shields.io/github/license/draganm/gorefresh)

Are you tired of running go run every time you make a change to your Go code? Do you want to run your Go code as soon as you save it? gorefresh is a simple tool that watches your Go files and runs them whenever they change.
I built this because I was tired of configuring [reflex](https://github.com/cespare/reflex) or [air](https://github.com/cosmtrek/air), especially when using embedded files. I wanted a simple tool that would just work.

## 🌟 Key Features

- Automatically discovers dependencies (Go, cgo and embedded files) based on the main module directory.
- Builds and restarts only if the dependencies have changed (see [gosha](https://github.com/draganm/gosha)).
- Supports program arguments.

## 📘 Usage

```bash
gorefresh <main-module-directory> -- [ program arguments ]
```

## 📥 Installation

Install the CLI tool using `go get`:

```bash
go install github.com/draganm/gorefresh
```

## Future work

- Support for build arguments
- Restart only if the build outputs change

## 👥 Contributing

Contributions are welcome! Feel free to submit issues for bug reports, feature requests, or even pull requests.

## 📜 License

This project is licensed under the MIT License.
