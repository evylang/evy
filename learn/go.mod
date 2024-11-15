module evylang.dev/evy/learn

go 1.23.0

require (
	evylang.dev/evy v0.1.199
	github.com/alecthomas/kong v1.4.0
	golang.org/x/tools v0.27.0
	gopkg.in/yaml.v3 v3.0.1
	rsc.io/markdown v0.0.0-20240717201619-868a055c40ae
)

require golang.org/x/text v0.20.0 // indirect

replace evylang.dev/evy => ..
