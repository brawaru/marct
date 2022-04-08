package sdtypes

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"
)

type EpochXCoder struct {
	precision time.Duration
}

func (x *EpochXCoder) DecodeInt(i int64) time.Time {
	return time.Unix(0, i*x.precision.Nanoseconds())
}

func (x *EpochXCoder) DecodeJSON(data []byte) (t time.Time, err error) {
	var n int64
	if err = json.Unmarshal(data, &n); err != nil {
		err = fmt.Errorf("decode %v: %w", data, err)
	} else {
		t = x.DecodeInt(n)
	}

	return
}

func (x *EpochXCoder) DecodeXMLElement(elem *xml.Decoder, start xml.StartElement) (t time.Time, err error) {
	var n int64
	if e := elem.DecodeElement(&n, &start); e == nil {
		t = x.DecodeInt(n)
	} else {
		err = fmt.Errorf("error decoding XML element: %v", e)
	}
	return
}

func (x *EpochXCoder) DecodeString(s string) (t time.Time, err error) {
	var n int64
	if n, err = strconv.ParseInt(s, 10, 64); err == nil {
		t = x.DecodeInt(n)
	} else {
		err = fmt.Errorf("decode %q as epoch timestamp: %w", s, err)
	}
	return
}

func (x EpochXCoder) DecodeXMLAttr(attr xml.Attr) (t time.Time, err error) {
	t, err = x.DecodeString(attr.Value)
	if err != nil {
		err = fmt.Errorf("decode XML attr %q with value %q: %w", attr.Name, attr.Value, err)
	}
	return
}

func (x *EpochXCoder) EncodeInt(t time.Time) int64 {
	return t.UnixNano() / x.precision.Nanoseconds()
}

func (x *EpochXCoder) EncodeString(t time.Time) string {
	return strconv.FormatInt(x.EncodeInt(t), 10)
}

func (x *EpochXCoder) EncodeJSON(t time.Time) ([]byte, error) {
	return json.Marshal(x.EncodeInt(t))
}

func (x *EpochXCoder) EncodeXMLAttr(t time.Time, name xml.Name) xml.Attr {
	return xml.Attr{
		Name:  name,
		Value: x.EncodeString(t),
	}
}

func (x *EpochXCoder) EncodeXML(t time.Time, elem *xml.Encoder, start xml.StartElement) error {
	return elem.EncodeElement(x.EncodeInt(t), start)
}

// EpochTime represents a time.Time that is encoded and decoded as a Unix timestamp in seconds.
type EpochTime struct {
	time.Time
}

func (e *EpochTime) MarshalJSON() ([]byte, error) {
	coder := EpochXCoder{time.Second}
	return coder.EncodeJSON(e.Time)
}

func (e *EpochTime) MarshalXML(elem *xml.Encoder, start xml.StartElement) error {
	coder := EpochXCoder{time.Second}
	return coder.EncodeXML(e.Time, elem, start)
}

func (e *EpochTime) MarshalXMLAttr(name xml.Name) xml.Attr {
	coder := EpochXCoder{time.Second}
	return coder.EncodeXMLAttr(e.Time, name)
}

func (e *EpochTime) UnmarshalJSON(data []byte) error {
	coder := EpochXCoder{time.Second}
	t, err := coder.DecodeJSON(data)
	if err != nil {
		return err
	}
	*e = EpochTime{t}
	return nil
}

func (e *EpochTime) UnmarshalXML(elem *xml.Decoder, start xml.StartElement) error {
	coder := EpochXCoder{time.Second}
	t, err := coder.DecodeXMLElement(elem, start)
	if err != nil {
		return err
	}
	*e = EpochTime{t}
	return nil
}

func (e *EpochTime) UnmarshalXMLAttr(attr xml.Attr) error {
	coder := EpochXCoder{time.Second}
	t, err := coder.DecodeXMLAttr(attr)
	if err != nil {
		return err
	}
	*e = EpochTime{t}
	return nil
}

// EpochTimeMillis represents a time.Time that is encoded and decoded as a Unix timestamp in milliseconds.
type EpochTimeMillis struct {
	time.Time
}

func (e *EpochTimeMillis) MarshalJSON() ([]byte, error) {
	coder := EpochXCoder{time.Millisecond}
	return coder.EncodeJSON(e.Time)
}

func (e *EpochTimeMillis) MarshalXML(elem *xml.Encoder, start xml.StartElement) error {
	coder := EpochXCoder{time.Millisecond}
	return coder.EncodeXML(e.Time, elem, start)
}

func (e *EpochTimeMillis) MarshalXMLAttr(name xml.Name) xml.Attr {
	coder := EpochXCoder{time.Millisecond}
	return coder.EncodeXMLAttr(e.Time, name)
}

func (e *EpochTimeMillis) UnmarshalJSON(data []byte) error {
	coder := EpochXCoder{time.Millisecond}
	t, err := coder.DecodeJSON(data)
	if err != nil {
		return err
	}
	*e = EpochTimeMillis{t}
	return nil
}

func (e *EpochTimeMillis) UnmarshalXML(elem *xml.Decoder, start xml.StartElement) error {
	coder := EpochXCoder{time.Millisecond}
	t, err := coder.DecodeXMLElement(elem, start)
	if err != nil {
		return err
	}
	*e = EpochTimeMillis{t}
	return nil
}

func (e *EpochTimeMillis) UnmarshalXMLAttr(attr xml.Attr) error {
	coder := EpochXCoder{time.Millisecond}
	t, err := coder.DecodeXMLAttr(attr)
	if err != nil {
		return err
	}
	*e = EpochTimeMillis{t}
	return nil
}

// EpochTimeNanos represents a time.Time that is encoded and decoded as a Unix timestamp in nanoseconds.
type EpochTimeNanos struct {
	time.Time
}

func (e *EpochTimeNanos) MarshalJSON() ([]byte, error) {
	coder := EpochXCoder{time.Nanosecond}
	return coder.EncodeJSON(e.Time)
}

func (e *EpochTimeNanos) MarshalXML(elem *xml.Encoder, start xml.StartElement) error {
	coder := EpochXCoder{time.Nanosecond}
	return coder.EncodeXML(e.Time, elem, start)
}

func (e *EpochTimeNanos) MarshalXMLAttr(name xml.Name) xml.Attr {
	coder := EpochXCoder{time.Nanosecond}
	return coder.EncodeXMLAttr(e.Time, name)
}

func (e *EpochTimeNanos) UnmarshalJSON(data []byte) error {
	coder := EpochXCoder{time.Nanosecond}
	t, err := coder.DecodeJSON(data)
	if err != nil {
		return err
	}
	*e = EpochTimeNanos{t}
	return nil
}

func (e *EpochTimeNanos) UnmarshalXML(elem *xml.Decoder, start xml.StartElement) error {
	coder := EpochXCoder{time.Nanosecond}
	t, err := coder.DecodeXMLElement(elem, start)
	if err != nil {
		return err
	}
	*e = EpochTimeNanos{t}
	return nil
}

func (e *EpochTimeNanos) UnmarshalXMLAttr(attr xml.Attr) error {
	coder := EpochXCoder{time.Nanosecond}
	t, err := coder.DecodeXMLAttr(attr)
	if err != nil {
		return err
	}
	*e = EpochTimeNanos{t}
	return nil
}
