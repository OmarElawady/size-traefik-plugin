module github.com/OmarElawady/size-traefik-plugin

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.24.0
	github.com/threefoldtech/zbus v0.1.5
	github.com/threefoldtech/zos v0.5.2
)

replace github.com/threefoldtech/zos => ../zos
