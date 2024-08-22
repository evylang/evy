module evylang.dev/evy/learn

go 1.22.0

toolchain go1.22.6

require (
	evylang.dev/evy v0.1.160
	github.com/alecthomas/kong v0.9.0
	golang.org/x/tools v0.24.0
	gopkg.in/yaml.v3 v3.0.1
	rsc.io/markdown v0.0.0-20240717201619-868a055c40ae
)

require golang.org/x/text v0.17.0 // indirect

replace evylang.dev/evy => ..
