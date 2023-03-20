# didactic_compiler

DTU Master thesis project which compiles a Go like language (the source) into
RISC-V assembly (the target).

## Requirements

The project is built with Go and therefore, requires [Go](https://go.dev/) to
build.

To test the outputted RISC-V assembly use
[RARS](https://github.com/TheThirdOne/rars).

## Build

To build be at the root of the project and invoke:

```bash
go build
```

this will produce a executable binary at the root of the project called
`didactic_compiler`.

## Test

To run the unit tests be at the root of the project and invoke:

```bash
go test ./...
```

## Usage

The compiler can take one argument which is a source file and will output to
standard out RISC-V assembly.

In the directory `testdata/` are some example source files.

---

If the compiler is called without an argument, then it enters an interactive
mode, and one can enter source code which is compiled on the fly.
