// +build windows

package watch

import (
	"encoding/json"
	"encoding/xml"
	"sync"
	"time"
)

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
	EventData struct {
		Text string `xml:",chardata"`
		Data []struct {
			Text string `xml:",chardata"`
			Name string `xml:"Name,attr"`
		} `xml:"Data"`
	} `xml:"EventData"`
}

// Stores the common fields from a log event
type WinLogEvent struct {
	// XML
	XmlText string `lua:"xml_text"`
	XmlErr  error  `lua:"xml_err"`

	// From EvtRender
	ProviderName      string    `lua:"provider_name"`
	EventId           uint64    `lua:"event_id"`
	Qualifiers        uint64    `lua:"qualifiers"`
	Level             uint64    `lua:"level"`
	Task              uint64    `lua:"task"`
	Opcode            uint64    `lua:"opcode"`
	Created           time.Time `lua:"created"`
	RecordId          uint64    `lua:"record_id"`
	ProcessId         uint64    `lua:"process_id"`
	ThreadId          uint64    `lua:"thread_id"`
	Channel           string    `lua:"channel"`
	ComputerName      string    `lua:"computer_name"`
	Version           uint64    `lua:"version"`
	RenderedFieldsErr error     `lua:"rendered_fields_err"`

	// From EvtFormatMessage
	Msg                string   `lua:"msg"`
	LevelText          string   `lua:"level_text"`
	TaskText           string   `lua:"task_text"`
	OpcodeText         string   `lua:"opcode_text"`
	Keywords           string   `lua:"keywords"`
	ChannelText        string   `lua:"channel_text"`
	ProviderText       string   `lua:"provider_text"`
	IdText             string   `lua:"id_text"`
	PublisherHandleErr error    `lua:"publisher_handle_err"`

	// Serialied XML bookmark to
	// restart at this event
	Bookmark string             `lua:"bookmark"`

	// Subscribed channel from which the event was retrieved,
	// which may be different than the event's channel
	SubscribedChannel string    `lua:"subscribed_channel"`
}

type channelWatcher struct {
	subscription ListenerHandle
	callback     *LogEventCallbackWrapper
	bookmark     BookmarkHandle
}

// Watches one or more event log channels
// and publishes events and errors to Go
// channels
type WinLogWatcher struct {
	errChan   chan error
	eventChan chan *WinLogEvent

	renderContext SysRenderContext
	watches       map[string]*channelWatcher
	watchMutex    sync.Mutex
	shutdown      chan interface{}

	// Optionally render localized fields. EvtFormatMessage() is slow, so
	// skipping these fields provides a big speedup.
	RenderKeywords bool
	RenderMessage  bool
	RenderLevel    bool
	RenderTask     bool
	RenderProvider bool
	RenderOpcode   bool
	RenderChannel  bool
	RenderId       bool
}

type SysRenderContext uint64
type ListenerHandle uint64
type PublisherHandle uint64
type EventHandle uint64
type BookmarkHandle uint64

type LogEventCallback interface {
	PublishError(error)
	PublishEvent(EventHandle, string)
}

type LogEventCallbackWrapper struct {
	callback          LogEventCallback
	subscribedChannel string
}

func (evt *WinLogEvent) Marshal() ([]byte , error) {
	return json.Marshal(*evt)
}

var null = []byte("")
func (evt *WinLogEvent) Json() []byte {
	data , err := evt.Marshal()
	if err != nil {
		return null
	}
	return data
}