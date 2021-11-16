package event

import "encoding/xml"

type XmlEvent struct {
	XMLName xml.Name `xml:"Event"`
	Text    string   `xml:",chardata"`
	Xmlns   string   `xml:"xmlns,attr"`
	System  struct {
		Text     string `xml:",chardata"`
		Provider struct {
			Text string `xml:",chardata"`
			Name string `xml:"Name,attr"`
			Guid string `xml:"Guid,attr"`
		} `xml:"Provider"`
		EventID     string `xml:"EventID"`
		Version     string `xml:"Version"`
		Level       string `xml:"Level"`
		Task        string `xml:"Task"`
		Opcode      string `xml:"Opcode"`
		Keywords    string `xml:"Keywords"`
		TimeCreated struct {
			Text       string `xml:",chardata"`
			SystemTime string `xml:"SystemTime,attr"`
		} `xml:"TimeCreated"`
		EventRecordID string `xml:"EventRecordID"`
		Correlation   struct {
			Text       string `xml:",chardata"`
			ActivityID string `xml:"ActivityID,attr"`
		} `xml:"Correlation"`
		Execution struct {
			Text      string `xml:",chardata"`
			ProcessID string `xml:"ProcessID,attr"`
			ThreadID  string `xml:"ThreadID,attr"`
		} `xml:"Execution"`
		Channel  string `xml:"Channel"`
		Computer string `xml:"Computer"`
		Security string `xml:"Security"`
	} `xml:"System"`
	EvData EventData `xml:"EventData"`
}

type EventDataKV struct {
	Text string `xml:",chardata"`
	Name string `xml:"Name,attr"`
}

type EventData struct {
	Text string        `xml:",chardata"`
	Data []EventDataKV `xml:"Data"`
}