module github.com/lucas-clemente/quic-go

go 1.12

require (
	github.com/cheekybits/genny v1.0.0
	github.com/cloudflare/sidh v0.0.0-20181111220428-fc8e6378752b // indirect
	github.com/golang/mock v1.3.1
	github.com/marten-seemann/qtls v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190311183353-d8887717615a
	golang.org/x/sys v0.0.0-20190228124157-a34e9553db1e // indirect
)

replace github.com/marten-seemann/qtls => github.com/marten-seemann/qtls-deprecated v0.0.0-20190207043627-591c71538704
