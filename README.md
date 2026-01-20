# go-specification

Для тех, кто будет работать с go, советую установить gvm (Go Version Manager)
Внутри проекта `go applymod`

# generate overlay

`go run framework/cli/main.go project --plugin dynamic_checks ./samples/projects/simple/ ./overlays`

# run without checks

`go run samples/projects/simple/main.go`

# run with enabled checks

`go run --overlay overlays/gmpa_overlay.json samples/projects/simple/main.go`