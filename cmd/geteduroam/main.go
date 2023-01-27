package main

import (
	"log"
	"fmt"
	"github.com/geteduroam/linux/internal/discovery"
	"github.com/geteduroam/linux/internal/eap"
	"github.com/geteduroam/linux/internal/instance"
	"os"
	"strings"
	"strconv"
)

func FilterByName(search string, instances *[]instance.Instance) *[]instance.Instance {
	x := []instance.Instance{}
	for _, i := range *instances {
		l1 := strings.ToLower(i.Name)
		l2 := strings.ToLower(search)
		if strings.Contains(l1, l2) {
			x = append(x, i)
		}
	}
	return &x
}

func ask(prompt string, validator func(input string) bool) string {
	for {
		var x string
		fmt.Print(prompt)
		fmt.Scanln(&x)

		if validator(x) {
			return x
		}
	}
}

func filteredOrganizations(orgs *[]instance.Instance) (f *[]instance.Instance) {
	for {
		x := ask("Please enter your organization (e.g. SURF): ", func(x string) bool {
			if len(x) == 0 {
				fmt.Fprintln(os.Stderr, "Your organization cannot be empty")
				return false
			}
			return true
		})
		f = FilterByName(x, orgs)
		if f != nil && len(*f) > 0 {
			break
		}
		fmt.Fprintf(os.Stderr, "No organizations found with search term: %v. Please try again\n", x)
	}
	return f
}

func validateOrg(input string, n int) bool {
	r, err := strconv.ParseInt(input, 10, 32)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid choice. Please enter a number")
		return false
	}
	if r <= 0 || r > int64(n) {
		fmt.Fprintf(os.Stderr, "Invalid choice range. Please enter an input between: %v and %v\n", 1, n)
		return false
	}
	return true
}

func getOrganization(orgs *[]instance.Instance) *instance.Instance {
	f := *filteredOrganizations(orgs)
	fmt.Println("Found the following matches: ")
	for n, c := range f {
		fmt.Printf("[%d] %s\n", n+1, c.Name)
	}
	input := ask("Please enter a choice: ", func(input string) bool {
		return validateOrg(input, len(f))
	})
	r, err := strconv.ParseInt(input, 10, 32)
	// This can't happen because we already validated that this can be parsed
	if err != nil {
		panic(err)
	}
	return &f[r-1]
}

func main() {
	c := discovery.NewCache()
	i, err := c.Instances()
	if err != nil {
		log.Fatalf("failed to get instances from discovery: %v", err)
	}

	chosen := getOrganization(i)
	config, err := chosen.Profiles[0].EAP()
	if err != nil {
		log.Fatalf("Could not obtain eap config: %v", err)
	}
	p, err := eap.Parse(config)
	if err != nil {
		log.Fatalf("Error with EAP: %v", err)
	}
	m, err := p.AuthenticationType()
	if err != nil {
		log.Fatalf("error getting authentication: %v", err)
	}
	fmt.Println("Got authentication:", m)
}
