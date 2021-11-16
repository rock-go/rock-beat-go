package event

import (
	"encoding/json"
	"encoding/xml"
	"github.com/rock-go/rock-beat-go/windows/event/watch"
	"github.com/rock-go/rock/audit"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/node"
	"strings"
	"time"
)


type logonEv struct {
	Time         time.Time          `xml:"-"         json:"time"`
	EventID      uint64             `xml:"-"         json:"event_id"`
	TaskText string    `xml:"-"         json:"task_text"`
	EvData   EventData `xml:"EventData" json:"data"`
	PC       string    `xml:"-"         json:"pc"`
	UserName     string             `xml:"-"         json:"username"`
	Remote       string             `xml:"-"         json:"remote"`
	Port         string             `xml:"-"         json:"port"`
	SubStatus    string             `xml:"-"         json:"sub_status"`
	Msg          string             `xml:"-"         json:"msg"`
	Keywords     string             `xml:"-"         json:"keywords"`
	Send         bool               `xml:"-"         json:"send"`
}

func (lev *logonEv) handle(ev *audit.Event) {
	data , err := json.Marshal(lev)
	if err == nil {
		node.Put(winLoginBucket, lua.S2B(time.Now().Format(time.RFC3339Nano)) , data)
	} else {
		ev.Set(audit.E(err))
	}
}

func (wv *winEv) logon(evt *watch.WinLogEvent) {
	var ev *audit.Event
	defer func() { if ev != nil { ev.Time(evt.Created).Put() } }()


	switch evt.EventId {
	case 4625: //账户登录成功
		ev = audit.NewEvent("win-logon" ,
			audit.Subject("账户登录失败"),
			audit.From(wv.vm.CodeVM()),
			audit.Time(evt.Created))

	case 4624: //账户登陆失败
		ev = audit.NewEvent("win-logon" ,
			audit.Subject("账户登录成功"),
			audit.From(wv.vm.CodeVM()),
			audit.Time(evt.Created))

	case 4634:
		ev = audit.NewEvent("win-logon" ,
			audit.Subject("账户注销"),
			audit.From(wv.vm.CodeVM()),
			audit.Time(evt.Created))

	default:
		//todo

		return
	}

 	msg := strings.ReplaceAll(evt.Msg , "\r\n" , "\n")
	msg = strings.ReplaceAll(msg , "\n\n" , "\n")
	msg = strings.ReplaceAll(msg , "\t\t" , " ")
	ev.Set(audit.Msg(msg))

	lev := logonEv{
		Time:      evt.Created,
		EventID:   evt.EventId,
		TaskText:  evt.TaskText,
		Keywords:  evt.Keywords,
		Msg:       msg,
		Send:      false,
	}

	err := xml.Unmarshal( auxlib.S2B(evt.XmlText) , &lev)
	if err != nil {
		ev.Set(audit.E(err))
		return
	}

	for _ , item := range lev.EvData.Data {
		switch item.Name {
		case "TargetUserName":
			ev.Set(audit.User(item.Text))
			lev.UserName = item.Text
		case "TargetDomainName":
			lev.PC = item.Text
		case "SubStatus":
			lev.SubStatus = item.Text
		case "IpAddress":
			ev.Set(audit.RAddr(item.Text))
			lev.Remote = item.Text
		case "IpPort":
			ev.Set(audit.RPort(auxlib.ToInt(item.Text)))
			lev.Port = item.Text
		}
	}

	lev.handle(ev)
	return
}