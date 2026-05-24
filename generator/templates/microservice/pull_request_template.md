## Summary

<!-- What does this PR do? Why? -->

## Type of change

- [ ] Bug fix
- [ ] New feature
- [ ] Refactoring (no functional change)
- [ ] Documentation

## Checklist

- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] New code follows the architecture in `ARCHITECTURE.md`
- [ ] No business logic in handlers
- [ ] No direct postgres imports in services or handlers
- [ ] Errors mapped correctly (repo errors -> service errors -> HTTP status)
