package launcher

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

type Resolution struct {
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
}

func (r *Resolution) MarshalJSON() ([]byte, error) {
	if r == nil {
		return json.Marshal(nil)
	}

	m := map[string]int{}

	if r.Width > 0 {
		m["width"] = r.Width
	}

	if r.Height > 0 {
		m["height"] = r.Height
	}

	return json.Marshal(m)
}

func (r *Resolution) UnmarshalJSON(data []byte) error {
	var m map[string]int

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	if h, ok := m["height"]; ok {
		r.Height = h
	} else {
		r.Height = 0
	}

	if w, ok := m["width"]; ok {
		r.Width = w
	} else {
		r.Width = 0
	}

	return nil
}

type AspectRatio struct {
	Width  int
	Height int
}

// CalculateMissingSides calculates sizes for missing sides of resolution given the aspect ratio.
// If both sides are missing, then returns the initial value.
func (r *Resolution) CalculateMissingSides(ratio AspectRatio) (res Resolution) {
	res.Width = r.Width
	res.Height = r.Height

	if r != nil {
		if r.Width > 0 {
			res.Height = ratio.Height * (r.Width / ratio.Width)
		} else if r.Height > 0 {
			res.Width = ratio.Width * (r.Height / ratio.Height)
		}
	}

	return
}

func (r *Resolution) Max(c Resolution) (res Resolution) {
	if r != nil {
		res.Width = int(math.Max(float64(r.Width), float64(c.Width)))
		res.Height = int(math.Max(float64(r.Height), float64(c.Height)))
	}

	return
}

func (r *Resolution) Min(c Resolution) (res Resolution) {
	if r != nil {
		res.Width = int(math.Min(float64(r.Width), float64(c.Width)))
		res.Height = int(math.Min(float64(r.Height), float64(c.Height)))
	}

	return
}

// ParseResolution parses resolution in format WIDTHxHEIGHT or HEIGHTp, which returns LauncherProfileResolution with
// width and height fields set accordingly. Height is optional and can be omitted.
//
// For value "auto" this function will return nothing (neither resolution, nor error).
func ParseResolution(v string) (*Resolution, error) {
	if v == "auto" {
		return nil, nil
	}

	var w int
	var h int

	if strings.HasSuffix(v, "p") {
		p := strings.TrimSuffix(v, "p")
		if i, err := strconv.Atoi(p); err != nil {
			return nil, &ResolutionParseError{
				Input: v,
				Err: &IllegalDimensionValueError{
					Value:     p,
					Dimension: "height",
					Err:       err,
				},
			}
		} else {
			h = i
		}
	} else {
		d := strings.Split(v, "x")
		l := len(d)

		switch {
		case l > 2:
			return nil, &ResolutionParseError{
				Input: v,
				Err:   &IllegalDimensionsNumberError{l},
			}
		case l == 2:
			if i, err := strconv.Atoi(d[1]); err != nil {
				return nil, &ResolutionParseError{
					Input: v,
					Err: &IllegalDimensionValueError{
						Value:     d[1],
						Dimension: "height",
						Err:       err,
					},
				}
			} else {
				h = i
			}
			fallthrough
		case l == 1:
			if i, err := strconv.Atoi(d[0]); err == nil {
				w = i
			} else {
				return nil, &ResolutionParseError{
					Input: v,
					Err: &IllegalDimensionValueError{
						Value:     d[0],
						Dimension: "width",
						Err:       err,
					},
				}
			}
		case l == 0:
			return nil, &ResolutionParseError{
				Input: v,
				Err:   &IllegalDimensionsNumberError{-1},
			}
		}
	}

	return &Resolution{
		Width:  w, // FIXME: maybe you should not use pointers for this one bruv
		Height: h,
	}, nil
}
