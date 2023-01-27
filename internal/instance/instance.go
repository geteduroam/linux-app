package instance

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
