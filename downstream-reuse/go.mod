module github.com/d0w/openshell-bff-examples/downstream-reuse

go 1.26.5

require github.com/d0w/openshell-bff-examples/upstream v0.0.0

require github.com/go-chi/chi/v5 v5.3.1 // indirect

// upstream hasn't tagged a release yet; point at the local checkout.
// Once upstream cuts e.g. v0.1.0, drop this line and `go get` the real tag.
replace github.com/d0w/openshell-bff-examples/upstream => ../upstream
