module evylang.dev/evy/learn

go 1.22

require (
	evylang.dev/evy v0.1.130
	github.com/alecthomas/kong v0.9.0
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d
	gopkg.in/yaml.v3 v3.0.1
	rsc.io/markdown v0.0.0-20240603215554-74725d8a840a
)

require golang.org/x/text v0.16.0 // indirect

replace evylang.dev/evy => ..
