package launcher

import (
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLog4JParser(t *testing.T) {
	x := `<log4j:Event logger="MyLogger" timestamp="1648475574576" level="INFO" thread="main">
	<log4j:Message><![CDATA[This is a test]]></log4j:Message>
</log4j:Event>`

	// FIXME: we do not handle additional properties, though we probably should for completeness?

	var r Log4JEvent
	err := xml.Unmarshal([]byte(x), &r)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Equal(t, "MyLogger", r.Logger) {
		return
	}

	if !assert.Equal(t, "INFO", r.Level) {
		return
	}

	if !assert.Equal(t, "main", r.Thread) {
		return
	}

	if !assert.Equal(t, "This is a test", r.Message.Content) {
		return
	}

	if !assert.Equal(t, int64(1648475574576), r.Timestamp.UnixMilli()) {
		return
	}

	b, err := xml.Marshal(r)
	if !assert.NoError(t, err) {
		return
	}

	t.Log(string(b))
}
