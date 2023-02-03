package instance

import "strings"

type Instance struct {
	CatIDP  int    `json:"cat_idp"`
	Country string `json:"country"`
	Geo     []struct {
		lat  float32
		long float32
	} `json:"geo"`
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Profiles []Profile `json:"profiles"`
}

type Instances []Instance

// Filter filters a list of instances
func (i *Instances) Filter(search string) *Instances {
	x := Instances{}
	for _, i := range *i {
		l1 := strings.ToLower(i.Name)
		l2 := strings.ToLower(search)
		if strings.Contains(l1, l2) {
			x = append(x, i)
		}
	}
	return &x
}
