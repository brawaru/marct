package sdtypes

import (
	"encoding/json"
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUNIXJSON(t *testing.T) {
	j := `{"timestamp":1649351651325}`
	type result struct {
		Timestamp EpochTimeMillis `json:"timestamp"`
	}
	var r result
	err := json.Unmarshal([]byte(j), &r)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, int64(1649351651325), r.Timestamp.UnixMilli()) {
		return
	}
}

func TestUNIXXMLAttr(t *testing.T) {
	x := `<Result timestamp="1649351651325"/>`
	type result struct {
		Timestamp EpochTimeMillis `xml:"timestamp,attr"`
	}
	var r result
	err := xml.Unmarshal([]byte(x), &r)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, int64(1649351651325), r.Timestamp.UnixMilli()) {
		return
	}
}

func TestUNIXXML(t *testing.T) {
	x := `
<Result><Timestamp>1649351651325</Timestamp></Result>
`
	type result struct {
		Timestamp EpochTimeMillis `xml:"Timestamp"`
	}
	var r result
	err := xml.Unmarshal([]byte(x), &r)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, int64(1649351651325), r.Timestamp.UnixMilli()) {
		return
	}
}

func TestUNIXMillisJSON(t *testing.T) {
	j := `{"timestamp":1649351651}`
	type result struct {
		Timestamp EpochTimeMillis `json:"timestamp"`
	}
	var r result
	err := json.Unmarshal([]byte(j), &r)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, int64(1649351651), r.Timestamp.UnixMilli()) {
		return
	}
}

func TestUNIXMillisXML(t *testing.T) {
	x := `
<Result><Timestamp>1649351651</Timestamp></Result>
`
	type result struct {
		Timestamp EpochTimeMillis `xml:"Timestamp"`
	}
	var r result
	err := xml.Unmarshal([]byte(x), &r)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, int64(1649351651), r.Timestamp.UnixMilli()) {
		return
	}
}
