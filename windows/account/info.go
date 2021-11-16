package account

type Info struct {
	Accounts []*Account
	Groups   []*Group
}

type As []*Account
type Gs []*Group

func GetInfo() *Info {
	return GetAll()
}
