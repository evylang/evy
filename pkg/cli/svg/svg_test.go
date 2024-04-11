//go:build !tinygo

package svg

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"evylang.dev/evy/pkg/assert"
)

var testStrokeWidth = 2.0

func TestGolden(t *testing.T) {
	testCases := map[string]SVG{
		"empty": {
			XMLNS:   "http://www.w3.org/2000/svg",
			ViewBox: "0 0 200 120",
		},
		"group": {
			XMLNS:    "http://www.w3.org/2000/svg",
			ViewBox:  "0 0 200 120",
			Elements: []any{Group{}},
		},
		"group-styled": {
			XMLNS:   "http://www.w3.org/2000/svg",
			ViewBox: "0 0 200 120",
			Elements: []any{Group{
				Attr: Attr{
					Fill:            "blue",
					Stroke:          "red",
					StrokeWidth:     &testStrokeWidth,
					StrokeLinecap:   "butt",
					StorkeDashArray: "3 5",
				},
			}},
		},
		"group-text-styled": {
			XMLNS:   "http://www.w3.org/2000/svg",
			ViewBox: "0 0 200 120",
			Elements: []any{Group{
				Attr: Attr{
					Fill: "blue",
				},
				TextAttr: TextAttr{
					TextAnchor: "middle",
					Baseline:   "middle",
				},
			}},
		},
		"text-simple": {
			XMLNS:   "http://www.w3.org/2000/svg",
			ViewBox: "0 0 200 120",
			Elements: []any{&Text{
				Attr: Attr{
					Fill: "blue",
				},
				X:     40,
				Y:     60,
				Value: "Hello, World!",
			}},
		},
		"circle": {
			XMLNS:   "http://www.w3.org/2000/svg",
			ViewBox: "0 0 200 120",
			Elements: []any{&Circle{
				Attr: Attr{
					Fill: "blue",
				},
				CX: 40,
				CY: 60,
				R:  20,
			}},
		},
		"ellipse-simple": {
			XMLNS:   "http://www.w3.org/2000/svg",
			ViewBox: "0 0 400 400",
			Elements: []any{&Ellipse{
				Attr: Attr{
					Fill: "blue",
				},
				CX: 100,
				CY: 200,
				RX: 50,
				RY: 70,
			}},
		},
		"nested": {
			XMLNS:   "http://www.w3.org/2000/svg",
			ViewBox: "0 0 200 120",
			Elements: []any{
				&Group{
					Attr: Attr{
						Fill: "red",
					},
					Elements: []any{
						&Circle{
							Attr: Attr{
								Fill: "blue",
							},
							CX: 40,
							CY: 60,
							R:  20,
						},
					},
				},
			},
		},
	}
	for name, data := range testCases {
		t.Run(name, func(t *testing.T) {
			wantFilename := filepath.Join("testdata", name+".svg")
			wantData, err := os.ReadFile(wantFilename)
			assert.NoError(t, err)
			want := string(wantData)
			want = strings.TrimSuffix(want, "\n")

			gotFilename := filepath.Join(t.TempDir(), name+"svg")
			gotFile, err := os.Create(gotFilename)
			assert.NoError(t, err)
			encoder := xml.NewEncoder(gotFile)
			encoder.Indent("", "  ")
			err = encoder.Encode(data)
			assert.NoError(t, err)
			err = gotFile.Close()
			assert.NoError(t, err)
			gotData, err := os.ReadFile(gotFilename)
			assert.NoError(t, err)
			got := string(gotData)

			assert.Equal(t, want, got)
		})
	}
}
