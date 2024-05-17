module evylang.dev/evy/learn

go 1.22

require (
	evylang.dev/evy v0.1.92
	github.com/alecthomas/kong v0.9.0
	gopkg.in/yaml.v3 v3.0.1
	rsc.io/markdown v0.0.0-20240117044121-669d2fdf1650
)

require (
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.20.0 // indirect
)

replace evylang.dev/evy => ..

// Remove when https://github.com/rsc/markdown/pull/18 is merged.
replace rsc.io/markdown => evylang.dev/markdown v0.0.0-20240503034508-36e9fda2871b
