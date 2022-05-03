package launcher

import (
	"encoding/xml"
)

type Log4JEventConsumer func(event Log4JEvent)

type Log4JWriter struct {
	Consumer Log4JEventConsumer
}

// Log4JWriter implements io.Writer interface where each write is parsed as XML of Log4JEvent and transformed using the provided Formatter.
func (w *Log4JWriter) Write(data []byte) (n int, err error) {
	var r Log4JEvent
	err = xml.Unmarshal(data, &r)
	if err != nil {
		return
	}
	w.Consumer(r)
	return len(data), nil
}
