module github.com/nmichlo/norfair-go/examples/simple

go 1.24.0

toolchain go1.24.9

require (
	github.com/nmichlo/norfair-go v0.0.0
	gonum.org/v1/gonum v0.8.1
)

// Use local norfair-go for development
replace github.com/nmichlo/norfair-go => ../..
