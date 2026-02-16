package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/geteduroam/linux-app/internal/utils"
	"golang.org/x/text/language"
)

// LocalizedString is a string localized in the language
type LocalizedString struct {
	// Display is the string that is displayed
	Display string `json:"display"`
	// Lang is the language for the display
	Lang string `json:"lang"`
}

// LocalizedStrings is a list of localized strings
type LocalizedStrings []LocalizedString

// Corpus returns the localized strings joined in one single search string
func (ls LocalizedStrings) Corpus() string {
	var corpus strings.Builder
	for _, v := range ls {
		corpus.WriteString(v.Display)
	}
	return corpus.String()
}

var systemLanguage = language.English

func setSystemLanguage() {
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	first := strings.Split(lang, ".")[0]
	tag, err := language.Parse(first)
	if err != nil {
		// TODO: log invalid language
		return
	}
	systemLanguage = tag
}

// Get gets a string based on the system language
func (ls LocalizedStrings) Get() string {
	// first get the non-empty values
	var disp string
	var conf language.Confidence
	m := language.NewMatcher([]language.Tag{systemLanguage})
	for _, val := range ls {
		// no display yet
		if disp == "" {
			disp = val.Display
			// we don't continue here as we still need to store the confidence
		}
		if val.Lang == "" {
			continue
		}
		t, err := language.Parse(val.Lang)
		// tag is invalid, just continue with the next option
		if err != nil {
			continue
		}

		// the confidence that this matches
		// is higher than the current confidence
		_, _, got := m.Match(t)
		if got > conf {
			disp = val.Display
			conf = got
		}
	}
	return disp
}

// Provider is the info for a single eduroam/getgovroam etc provider
type Provider struct {
	// ID is the identifier of the provider
	ID string `json:"id"`
	// Country is where the provider is from
	Country string `json:"country"`
	// Name is the name of the provider
	Name LocalizedStrings `json:"name"`
	// Profiles is the list of profiles for the provider
	Profiles []Profile `json:"profiles"`
}

// Providers is the list of providers
type Providers []Provider

// SortNames sorts two localized strings
func SortNames(a LocalizedStrings, b LocalizedStrings, search string) int {
	la := strings.ToLower(a.Corpus())
	lb := strings.ToLower(b.Corpus())
	bd := strings.Compare(la, lb)
	// compute the base difference which is based on alphabetical order
	// if no search is defined return the base difference
	if search == "" {
		return bd
	}
	lower := strings.ToLower(search)
	escaped := regexp.QuoteMeta(lower)
	match := regexp.MustCompile(fmt.Sprintf("(^|[\\P{L}])%s[\\P{L}]", escaped))
	mi := match.MatchString(la)
	mj := match.MatchString(lb)
	if mi == mj {
		// tiebreak on alphabetical order
		return bd
	} else if mi {
		return -1
	}
	return 1
}

// ByName is the struct that implements by name sorting
type ByName struct {
	// Providers is the list of providers
	Providers Providers
	// Search is the search string
	Search string
}

// Len returns the length of the ByName sort
func (s ByName) Len() int { return len(s.Providers) }

// Swap swaps two providers
func (s ByName) Swap(i, j int) { s.Providers[i], s.Providers[j] = s.Providers[j], s.Providers[i] }

// Less sorts the providers
func (s ByName) Less(i, j int) bool {
	diff := SortNames(s.Providers[i].Name, s.Providers[j].Name, s.Search)
	// if i is less than j, diff returns less than 0
	return diff < 0
}

// FilterSingle searches inside the corpus
func FilterSingle(name LocalizedStrings, search string) bool {
	l1, err1 := utils.RemoveDiacritics(strings.ToLower(name.Corpus()))
	l2, err2 := utils.RemoveDiacritics(strings.ToLower(search))
	if err1 != nil || err2 != nil {
		return false
	}
	if !strings.Contains(l1, l2) {
		return false
	}
	return true
}

// FilterSort filters and sorts a list of providers
// The sorting is done in reverse as this is used in the CLI where the most relevant providers should be shown at the bottom
func (i *Providers) FilterSort(search string) *Providers {
	x := ByName{
		Providers: Providers{},
		Search:    search,
	}
	for _, i := range *i {
		if FilterSingle(i.Name, search) {
			x.Providers = append(x.Providers, i)
		}
	}
	sort.Sort(sort.Reverse(ByName(x)))
	return &x.Providers
}

// Custom gets provider info using a custom URL
func Custom(ctx context.Context, query string) (*Provider, error) {
	client := http.Client{Timeout: 10 * time.Second}
	// parse URL and add scheme
	u, err := url.Parse(query)
	if err != nil {
		return nil, err
	}

	u.Scheme = "https"

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Add("Accept", "application/eap-config")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	b, err := readResponse(resp)
	if err != nil {
		return nil, err
	}
	p := &Provider{
		ID:   "custom_provider",
		Name: LocalizedStrings{{Display: "Custom Provider", Lang: "en"}},
	}
	prof := Profile{
		ID:   "custom_profile",
		Name: LocalizedStrings{{Display: "Custom Profile", Lang: "en"}},
	}
	ct := resp.Header.Get("Content-Type")
	pt := ""
	switch ct {
	case "application/json":
		pt = "letswifi"
	case "application/eap-config":
		pt = "eap-config"
	default:
		return nil, fmt.Errorf("unknown content type: %v", ct)
	}
	prof.Type = pt
	prof.CachedResponse = b
	p.Profiles = []Profile{prof}
	return p, nil
}

func init() {
	setSystemLanguage()
}
