module github.com/nmichlo/norfair-go/examples/simple

go 1.24.0

toolchain go1.24.9

require (
	github.com/nmichlo/norfair-go v0.0.0
	gonum.org/v1/gonum v0.16.0
)

require (
	github.com/arthurkushman/go-hungarian v0.0.0-20210331201642-2b0c3bc2fb3f // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/schollz/progressbar/v3 v3.18.0 // indirect
	gocv.io/x/gocv v0.42.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.36.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

// Use local norfair-go for development
replace github.com/nmichlo/norfair-go => ../..
