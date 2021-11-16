package account

import (
	"github.com/StackExchange/wmi"
	"github.com/rock-go/rock/json"
	"github.com/rock-go/rock/logger"
	"github.com/rock-go/rock/lua"
	"time"
)

type Group struct {
	Caption      string    `json:"caption"`
	Description  string    `json:"description"`
	Domain       string    `json:"domain"`
	InstallDate  time.Time `json:"install_date"`
	LocalAccount bool      `json:"local_account"`
	Name         string    `json:"name"`
	Sid          string    `json:"sid"`
	SidType      uint8     `json:"sid_type"`
	Status       string    `json:"status"`
}

type Account struct {
	AccountType        uint32    `json:"account_type"`
	Caption            string    `json:"caption"`
	Description        string    `json:"description"`
	Disabled           bool      `json:"disabled"`
	Domain             string    `json:"domain"`
	FullName           string    `json:"full_name"`
	InstallDate        time.Time `json:"install_date"`
	LocalAccount       bool      `json:"local_account"`
	Lockout            bool      `json:"lockout"`
	Name               string    `json:"name"`
	PasswordChangeable bool      `json:"password_changeable"`
	PasswordExpires    bool      `json:"password_expires"`
	PasswordRequired   bool      `json:"password_required"`
	SID                string    `json:"sid"`
	SIDType            uint8     `json:"sid_type"`
	Status             string    `json:"status"`
}

var (
	WQLAccount          = "SELECT * FROM Win32_UserAccount"
	WQLGroup            = "SELECT * FROM Win32_Account"

	account_bucket_name = []byte("windows_metric_account")
)

// GetAll 获取用户和用户组信息
func GetAll() *Info {
	accounts := GetAccounts()
	groups := GetGroups()
	return &Info{
		Accounts: accounts,
		Groups:   groups,
	}
}

func GetAccounts() As {
	var ret As
	err := wmi.Query(WQLAccount, &ret)
	if err != nil {
		logger.Errorf("metric account got account fail error %v" , err)
		return ret
	}
	return ret
}

func GetGroups() Gs {
	var ret Gs
	err := wmi.Query(WQLGroup, &ret)
	if err != nil {
		return nil
	}

	return ret
}

func (a As) Byte() []byte {
	buf := json.NewBuffer()
	buf.Arr("")

	for _ , item := range a {
		buf.Tab("")
		buf.KI("account_type"        ,  int(item.AccountType))
		buf.KV("caption"             ,  item.Caption)
		buf.KV("description"         ,  item.Description)
		buf.KB("disabled"            ,  item.Disabled)
		buf.KV("domain"              ,  item.Domain)
		buf.KV("full_name"           ,  item.FullName)
		buf.KT("install_date"        ,  item.InstallDate)
		buf.KB("local_account"       ,  item.LocalAccount)
		buf.KB("lockout"             ,  item.Lockout)
		buf.KV("name"                ,  item.Name)
		buf.KB("password_changeable" ,  item.PasswordChangeable)
		buf.KB("password_expires"    ,  item.PasswordExpires)
		buf.KB("password_required"   ,  item.PasswordRequired)
		buf.KV("sid"                 ,  item.SID)
		buf.KI("sid_type"            ,  int(item.SIDType))
		buf.KV("status"              ,  item.Status)
		buf.End("},")
	}

	buf.End("]")

	return buf.Bytes()
}

func (a As) String() string {
	return lua.B2S(a.Byte())
}

func (g Gs) Byte() []byte {
	buf := json.NewBuffer()
	buf.Arr("")
	for _ , item := range g {
		buf.Tab("")
		buf.KV("caption"      , item.Caption)
		buf.KV("description"  , item.Description)
		buf.KV("domain"       , item.Domain)
		buf.KT("install_date" , item.InstallDate)
		buf.KB("local_account", item.LocalAccount)
		buf.KV("name"         , item.Name)
		buf.KV("sid"          , item.Sid)
		buf.KI("sid_type"     , int(item.SidType))
		buf.KV("status"       , item.Status)
		buf.End("},")
	}

	buf.End("]")
	return buf.Bytes()
}

func (g Gs) String() string {
	return lua.B2S(g.Byte())
}