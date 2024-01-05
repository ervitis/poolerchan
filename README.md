# PoolerChan

Simple worker pool library for golang

### Tasks done

- N workers
- N tasks
- Logger
- Cleanup queues

### Pending tasks

- Handle errors
  - Ignore some kind of errors
- Check context

### Dev notes

- Using gotoolchain for dependency managements
- Use errgroup https://pkg.go.dev/golang.org/x/sync/errgroup -> Nope, it will stop on error
