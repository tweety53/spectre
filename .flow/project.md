# spectre — flow project configuration

## default landing route

merge and push

## lint

`gofmt -w .` is the auto-fix, run before the list. `gofmt -l` exits 0 whatever it lists, so the
list's line fails through `grep` when a file is unformatted.

```bash
! gofmt -l . | grep .
go vet ./...
```

## test

```bash
go test ./...
```
