# Workspace

This is a quick review of a Go workspace.

A Go workspace is a directory hierarchy with three directories at its root:

1. `src`: contains Go source files organized into packages (one package per directory),
2. `pkg`: contains package objects, and
3. `bin`: contains executable commands.

The `GOPATH` environment variable lists the places to look for Go workspaces. It is a colon-separated list of paths on Unix and a semicolon-separated list on Windows. If the environment variable is unset, Go uses a default path that depends on the operating system:

- On Unix systems, it is `$HOME/go`.
- On Windows, it is `%USERPROFILE%\go`.

Here's an example of what a workspace might look like:

```
bin/
    hello                          # command executable
    outyet                         # command executable
pkg/
    linux_amd64/
        github.com/golang/example/
            stringutil.a           # package object
src/
    github.com/golang/example/
        .git/                      # Git repository metadata
	hello/
	    hello.go               # command source
	outyet/
	    main.go                # command source
	    main_test.go           # test source
	stringutil/
	    reverse.go             # package source
	    reverse_test.go        # test source
```

In this workspace, the `hello` command is stored in the `hello` directory beneath `github.com/golang/example/`, corresponding to the import path of the command. The `hello` directory contains the `hello.go` source file, which declares a main package with two functions.

Note: As of Go 1.11, the Go team introduced a new concept called 'Go Modules' which is an experimental opt-in feature in Go 1.11 and 1.12, and is planned to be always on in Go 1.13. It allows for the deprecation of the GOPATH, and for work to be done outside of the GOPATH.

## Modules

Go Modules is a dependency management system that was introduced in Go 1.11. It allows for versioning and package distribution and is the official dependency management solution for the Go language.

Here are some key points about Go Modules:

1. **Versioning**: Go Modules allows you to specify the versions of your dependencies. Each version is associated with a specific commit in the dependency's repository. This ensures that all developers working on a project use the same versions of dependencies, which makes builds reproducible.

2. **Dependency resolution**: Go Modules automatically resolves dependencies for your project. When you import a package, Go Modules downloads the necessary version of the dependency and adds it to your `go.mod` file.

3. **Vendoring**: Go Modules supports vendoring, which is the process of making a copy of your dependencies in your project's repository. This ensures that your project is not affected by changes in its dependencies, and makes it possible to build your project without needing to download any dependencies.

4. **GOPATH independence**: Before Go Modules, Go developers had to put their code in the `GOPATH` directory. With Go Modules, you can put your code in any directory.

Here's an example of how to use Go Modules:

1. Initialize a new module:

    ```bash
    go mod init github.com/my/repo
    ```

    This creates a `go.mod` file that describes your module.

2. Add dependencies:

    Just import the packages you need in your code. Go will automatically add them to your `go.mod` file when you build or test your code.

3. Build your code:

    ```bash
    go build
    ```

    This will download the necessary dependencies and build your code.

4. Test your code:

    ```bash
    go test
    ```

    This will download the necessary dependencies for testing and run your tests.

5. Update dependencies:

    ```bash
    go get -u
    ```

    This will update all dependencies to their latest versions.

6. Check your dependencies:

    ```bash
    go list -m all
    ```

    This will list all current module versions being used.

## Multiple interdependent modules

The `go.work` file is a new feature introduced in Go 1.18 as part of the workspaces proposal. It allows you to specify a set of local modules that should be used for building and testing, even if there are other versions of those modules elsewhere in your `GOPATH` or in the module cache.

**The `go.work` file is useful when you're working on multiple interdependent modules. Instead of needing to update and tag each module every time you make a change, you can include all the modules in a `go.work` file and Go will use the local versions of those modules.**

Here's an example of what a `go.work` file might look like:

```go
go 1.18

use (
    ./foo
    ./bar
)
```

In this example, the `go.work` file specifies that the `foo` and `bar` modules, which are in directories relative to the `go.work` file, should be used.

When you run a Go command, it looks for a `go.work` file in the current directory or any parent directory. If it finds a `go.work` file, it uses the modules specified in the `go.work` file. If it doesn't find a `go.work` file, it uses the normal module resolution process.

Please note that as of the time of writing, the `go.work` feature is still experimental and its design may change in future versions of Go.

## Internal Packages

The internal keyword in Go is a special directory name that restricts the accessibility of the packages inside it.
Packages inside an internal directory can only be imported and used by the code that is in the same parent directory.

**The internal package is mainly used for internal implementation details that are shared across multiple packages within the parent package.**

In the case of net/internal, it contains implementation details and helper functions that are used by other packages within the net package, but are not intended to be directly used by programs that import the net package. This is a way to hide implementation details and prevent them from becoming part of the package's public API.
