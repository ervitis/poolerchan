# PoolerChan

Simple worker pool library for golang

## Getting Started

### Prerequisites

Requirements for the software and other tools to build, test and push

- [Go1.21](https://go.dev/doc/install)

### Installing

You can use the generator like this:

```bash
go get github.com/ervitis/poolerchan
```

You can see some examples in the folder `_examples`

## Running the tests

```bash
make tests
```

## Built With

- Go1.21

## Contributing

Please read [CONTRIBUTING.md](./.github/CONTRIBUTING.md) for details on our code
of conduct, and the process for submitting pull requests to us.

## Versioning

We use [Semantic Versioning](http://semver.org/) for versioning. For the versions
available, see the [tags on this
repository](https://github.com/PurpleBooth/a-good-readme-template/tags).

## Authors

- [@ervitis](https://github.com/ervitis)

## License

This project is licensed under the [Apache 2.0](LICENSE)

## Tasks

### Tasks done

- N workers
- N tasks
- Logger
- Cleanup queues
- Check context

### Pending tasks

- Handle errors
  - Ignore some kind of errors -> Nope possible

### Dev notes

- Using gotoolchain for dependency managements
- Use errgroup https://pkg.go.dev/golang.org/x/sync/errgroup -> Nope, it will stop on error
