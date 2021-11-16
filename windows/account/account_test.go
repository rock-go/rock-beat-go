package account

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetAccounts(t *testing.T) {
	d := GetAll()

	data, _ := json.Marshal(d)
	fmt.Printf("%s", data)
}

type Test struct {
	Name string  `json:"name"`
	Pct  float64 `json:"pct"`
}

func TestMarshal(t *testing.T) {
	test := Test{}
	test.Name = "test"
	d := 1
	f := 1
	test.Pct = div(float64(d), float64(f))
	//math.NaN()
	data, err := json.Marshal(test)
	if err != nil {
		fmt.Printf("error:%v\n", err)
		return
	}

	fmt.Printf("%s\n", data)
}

func div(d, f float64) float64 {
	return 0 / (d - f)
}
