package instance

import (
	"strings"

	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/utils"
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
	return &x
}
