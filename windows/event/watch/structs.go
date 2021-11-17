// +build windows

package watch

import (
	"encoding/xml"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/json"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/node"
	"strings"
	"sync"
	"time"
)

type EventDataKV struct {
	Text string `xml:",chardata"`
	Name string `xml:"Name,attr"`
}

type EventData struct {
	Text string        `xml:",chardata"`
	Data []EventDataKV `xml:"Data"`
}

type XmlEvent struct {
	XMLName xml.Name `xml:"Event"`
	Text    string   `xml:",chardata"`
	Xmlns   string   `xml:"xmlns,attr"`
	System  struct {
		Correlation   struct {
			Text       string `xml:",chardata"`
			ActivityID string `xml:"ActivityID,attr"`
		} `xml:"Correlation"`
		Security string `xml:"Security"`
	} `xml:"System"`
	EvData EventData `xml:"EventData"`
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
	ProviderText       string    `lua:"provider_text"`
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

func (xd *XmlEvent) Bytes() []byte {
	buff := json.NewBuffer()
	buff.Tab("")
	buff.KV("xml_space"      , xd.XMLName.Space)
	buff.KV("xml_local"      , xd.XMLName.Local)
	buff.KV("xmlns"          , xd.Xmlns)
	buff.KV("text"           , xd.Text)

	buff.KV("event_text"     , xd.EvData.Text)
	buff.Arr("event_data")
	for _ , item := range xd.EvData.Data {
		buff.KV(item.Name , item.Text)
	}
	buff.End("]}")

	return buff.Bytes()
}

func (xd *XmlEvent) Json(L *lua.LState) int {
	L.Push(lua.B2L(xd.Bytes()))
	return 1
}

func (xd *XmlEvent) Get(L *lua.LState , key string) lua.LValue {
	switch key {
	case "xml_space":
		return lua.S2L(xd.XMLName.Space)
	case "xml_local":
		return lua.S2L(xd.XMLName.Local)
	case "xmlns":
		return lua.S2L(xd.Xmlns)
	case "text":
		return lua.S2L(xd.Text)
	case "event_text":
		return lua.S2L(xd.EvData.Text)

	case "Json":
		return L.NewFunction(xd.Json)
	default:
		//todo
	}

	for _ , item := range xd.EvData.Data {
		if item.Name == key {
			return lua.S2L(item.Text)
		}
	}
	return lua.LNil
}

func (evt *WinLogEvent) Bytes() []byte {
	buff := json.NewBuffer()
	buff.Tab("")
	buff.KV("addr" , node.LoadAddr())
	buff.KV("node_id" , node.ID())
	buff.KV("provider_name" , evt.ProviderName)
	buff.KV("event_id" , evt.EventId)
	buff.KV("qualifiers" , evt.Qualifiers)
	buff.KV("level" , evt.Level)
	buff.KV("task" , evt.Task)
	buff.KV("op_code" , evt.Opcode)
	buff.KV("create_time" , evt.Created)
	buff.KV("record_id" , evt.RecordId)
	buff.KV("process_id" , evt.ProcessId)
	buff.KV("thread_id" , evt.ThreadId)
	buff.KV("channel" , evt.Channel)
	buff.KV("computer" , evt.ComputerName)
	buff.KV("version" , evt.Version)
	buff.KV("render_field_error" , evt.RenderedFieldsErr)

	//格式化
	txt := strings.ReplaceAll(evt.Msg , "\r" , "")
	txt = strings.ReplaceAll(txt , "\n" , " ")
	txt = strings.ReplaceAll(txt, "\t" , "")
	buff.KV("msg" , txt)

	buff.KV("level_text" , evt.LevelText)
	buff.KV("task_text" , evt.TaskText)
	buff.KV("op_code_text" , evt.OpcodeText)
	buff.KV("keywords" , evt.Keywords)
	buff.KV("channel_text" , evt.ChannelText)
	buff.KV("provider_text" , evt.ProviderText)
	buff.KV("id_text" , evt.IdText)
	buff.KV("publish_error" , evt.PublisherHandleErr)
	buff.KV("bookmark" ,strings.ReplaceAll(evt.Bookmark , "\r\n" , ""))
	buff.KV("subscribe" , evt.SubscribedChannel)
	buff.KV("xml_txt" , evt.XmlText)
	buff.KV("xml_error" , evt.XmlErr)
	buff.End("}")
	return buff.Bytes()
}

func (evt *WinLogEvent) Json(L *lua.LState) int {
	L.Push(lua.B2L(evt.Bytes()))
	return 1
}

func (evt *WinLogEvent) EvData(L *lua.LState) lua.LValue {
	var xd XmlEvent
	err := xml.Unmarshal(auxlib.S2B(evt.XmlText) , &xd)
	if err != nil {
		logger.Errorf("%v", err)
		return lua.LNil
	}

	return L.NewAnyData(&xd)
}

func (evt *WinLogEvent) Get(L *lua.LState , key string ) lua.LValue {
	switch key {
	case "xml":
		return lua.S2L(evt.XmlText)
	case "provider_name":
		return lua.S2L(evt.ProviderName)
	case "event_id":
		return lua.LNumber(evt.EventId)
	case "task":
		return lua.S2L(evt.TaskText)
	case "op_code":
		return lua.LNumber(evt.Opcode)
	case "create_time":
		return lua.S2L(evt.Created.Format(time.RFC3339Nano))
	case "record_id":
		return lua.LNumber(evt.RecordId)
	case "process_id":
		return lua.LNumber(evt.ProcessId)
	case "thread_id":
		return lua.LNumber(evt.ThreadId)
	case "channel":
		return lua.S2L(evt.Channel)
	case "computer":
		return  lua.S2L(evt.ComputerName)
	case "version":
		return lua.LNumber(evt.Version)
	case "render_field_err":
		return lua.S2L(evt.RenderedFieldsErr.Error())

	case "msg":
		txt := strings.ReplaceAll(evt.Msg , "\r\n" , "\n")
		txt = strings.ReplaceAll(txt, "\n\n" , "\n")
		txt = strings.ReplaceAll(txt, "\t\t" , " ")
		return  lua.S2L(txt)

	case "level_text":
		return  lua.S2L(evt.LevelText)
	case "task_text":
		return lua.S2L(evt.TaskText)
	case "op_code_text":
		return lua.S2L(evt.OpcodeText)
	case "keywords":
		return lua.S2L(evt.Keywords)
	case "channel_text":
		return lua.S2L(evt.ChannelText)
	case "id_text":
		return lua.S2L(evt.IdText)
	case "publish_err":
		return lua.S2L(evt.PublisherHandleErr.Error())
	case "bookmark":
		return lua.S2L(evt.Bookmark)
	case "subscribe":
		return lua.S2L(evt.SubscribedChannel)

	case "exdata":
		return evt.EvData(L)

	case "Json":
		return L.NewFunction(evt.Json)
	default:
		return lua.LNil
	}
}

func (evt *WinLogEvent) ToLValue(L *lua.LState) lua.LValue {
	return L.NewAnyData(evt)
}