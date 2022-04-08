package launcher

import (
	"encoding/xml"
	"github.com/brawaru/marct/sdtypes"
)

type Log4JMessage struct {
	XMLName xml.Name `xml:"Message"`
	Content string   `xml:",cdata"`
}

type Log4JEvent struct {
	XMLName   xml.Name                `xml:"Event"`
	Logger    string                  `xml:"logger,attr"`
	Timestamp sdtypes.EpochTimeMillis `xml:"timestamp,attr"`
	Level     string                  `xml:"level,attr"`
	Thread    string                  `xml:"thread,attr"`
	Message   Log4JMessage            `xml:"Message"`
}
