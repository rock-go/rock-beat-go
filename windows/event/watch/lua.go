package watch

import (
	"encoding/xml"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/json"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/node"
	"strings"
	"time"
	"unicode"
)

func (xd *XmlEvent) Json(L *lua.LState) int {
	L.Push(lua.B2L(xd.Bytes()))
	return 1
}

func (xd *XmlEvent) Index(L *lua.LState, key string) lua.LValue {
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

	for _, item := range xd.EvData.Data {
		if item.Name == key {
			return lua.S2L(item.Text)
		}
	}
	return lua.LNil
}

func (evt *WinLogEvent) Bytes() []byte {
	buff := json.NewEncoder()
	buff.Tab("")
	buff.KV("addr", node.LoadAddr())
	buff.KV("node_id", node.ID())
	buff.KV("provider_name", evt.ProviderName)
	buff.KV("event_id", evt.EventId)
	buff.KV("qualifiers", evt.Qualifiers)
	buff.KV("level", evt.Level)
	buff.KV("task", evt.Task)
	buff.KV("op_code", evt.Opcode)
	buff.KV("create_time", evt.Created)
	buff.KV("record_id", evt.RecordId)
	buff.KV("process_id", evt.ProcessId)
	buff.KV("thread_id", evt.ThreadId)
	buff.KV("channel", evt.Channel)
	buff.KV("computer", evt.ComputerName)
	buff.KV("version", evt.Version)
	buff.KV("render_field_error", evt.RenderedFieldsErr)

	//格式化
	txt := strings.ReplaceAll(evt.Msg, "\r", "")
	txt = strings.ReplaceAll(txt, "\n", " ")
	txt = strings.ReplaceAll(txt, "\t", "")
	buff.KV("msg", txt)

	buff.KV("level_text", evt.LevelText)
	buff.KV("task_text", evt.TaskText)
	buff.KV("op_code_text", evt.OpcodeText)
	buff.KV("keywords", evt.Keywords)
	buff.KV("channel_text", evt.ChannelText)
	buff.KV("provider_text", evt.ProviderText)
	buff.KV("id_text", evt.IdText)
	buff.KV("publish_error", evt.PublisherHandleErr)
	buff.KV("bookmark", strings.ReplaceAll(evt.Bookmark, "\r\n", ""))
	buff.KV("subscribe", evt.SubscribedChannel)

	text := strings.TrimFunc(evt.XmlText, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})

	buff.KV("xml_txt", text)
	buff.KV("xml_error", evt.XmlErr)
	buff.End("}")
	return buff.Bytes()
}

func (evt *WinLogEvent) Json(L *lua.LState) int {
	L.Push(lua.B2L(evt.Bytes()))
	return 1
}

func (evt *WinLogEvent) EvData(L *lua.LState) lua.LValue {
	var xd XmlEvent
	err := xml.Unmarshal(auxlib.S2B(evt.XmlText), &xd)
	if err != nil {
		logger.Errorf("%v", err)
		return lua.LNil
	}

	return L.NewAnyData(&xd)
}

func (evt *WinLogEvent) Index(L *lua.LState, key string) lua.LValue {
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
		return lua.S2L(evt.ComputerName)
	case "version":
		return lua.LNumber(evt.Version)
	case "render_field_err":
		return lua.S2L(evt.RenderedFieldsErr.Error())

	case "message":
		txt := strings.ReplaceAll(evt.Msg, "\r\n", "\n")
		txt = strings.ReplaceAll(txt, "\n\n", "\n")
		txt = strings.ReplaceAll(txt, "\t\t", " ")
		return lua.S2L(txt)

	case "level_text":
		return lua.S2L(evt.LevelText)
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
