module evylang.dev/evy/learn

go 1.23.0

require (
	evylang.dev/evy v0.1.207
	github.com/alecthomas/kong v1.7.0
	golang.org/x/tools v0.29.0
	gopkg.in/yaml.v3 v3.0.1
	rsc.io/markdown v0.0.0-20241212154241-6bf72452917f
)

require golang.org/x/text v0.21.0 // indirect

replace evylang.dev/evy => ..
