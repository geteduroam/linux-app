package instance

import (
	"sort"
	"strings"

	"github.com/geteduroam/linux-app/internal/utils"
)

type geo struct {
	Lat  float32 `json:"lat"`
	Long float32 `json:"long"`
}

type Instance struct {
	CatIDP   int       `json:"cat_idp"`
	Country  string    `json:"country"`
	Geo      []geo     `json:"geo"`
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Profiles []Profile `json:"profiles"`
}

type Instances []Instance

type ByName []Instance

func (s ByName) Len() int		{ return len(s) }
func (s ByName) Swap(i, j int)		{ s[i], s[j] = s[j], s[i] }
func (s ByName) Less(i, j int) bool {
	// Do we want to involve Profiles{}.Name in the sort
	// And if so, how?
	// For now we sort reverse as an example
	return s[i].Name > s[j].Name
}

// Filter filters a list of instances
func (i *Instances) Filter(search string) *Instances {
	x := Instances{}
	for _, i := range *i {
		l1, err1 := utils.RemoveDiacritics(strings.ToLower(i.Name))
		l2, err2 := utils.RemoveDiacritics(strings.ToLower(search))
		if err1 != nil || err2 != nil {
			continue
		}
		if strings.Contains(l1, l2) {
			x = append(x, i)
		}
	}
	sort.Sort(ByName(x))
	return &x
}
