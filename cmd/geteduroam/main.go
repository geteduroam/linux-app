package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/jwijenbergh/geteduroam-linux/internal/configure"
	"github.com/jwijenbergh/geteduroam-linux/internal/discovery"
	"github.com/jwijenbergh/geteduroam-linux/internal/instance"
)

func askSecret(prompt string, validator func(input string) bool) string {
	for {
		fmt.Print(prompt)
		pwd, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read password: %v", err)
			continue
		}
		pwdS := string(pwd)
		if validator(pwdS) {
			return pwdS
		}
	}
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

func filteredOrganizations(orgs *instance.Instances) (f *instance.Instances) {
	for {
		x := ask("Please enter your organization (e.g. SURF): ", func(x string) bool {
			if len(x) == 0 {
				fmt.Fprintln(os.Stderr, "Your organization cannot be empty")
				return false
			}
			return true
		})
		f = orgs.Filter(x)
		if f != nil && len(*f) > 0 {
			break
		}
		fmt.Fprintf(os.Stderr, "No organizations found with search term: %v. Please try again\n", x)
	}
	return f
}

func validateRange(input string, n int) bool {
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

func organization(orgs *instance.Instances) *instance.Instance {
	f := *filteredOrganizations(orgs)
	fmt.Println("Found the following matches: ")
	for n, c := range f {
		fmt.Printf("[%d] %s\n", n+1, c.Name)
	}
	input := ask("Please enter a choice for the organisation: ", func(input string) bool {
		return validateRange(input, len(f))
	})
	r, err := strconv.ParseInt(input, 10, 32)
	// This can't happen because we already validated that this can be parsed
	if err != nil {
		panic(err)
	}
	return &f[r-1]
}

func profile(profiles []instance.Profile) *instance.Profile {
	// Only one profile, return it immediately
	if len(profiles) == 1 {
		return &profiles[0]
	}
	// Multiple profiles found, we need to get the right one
	fmt.Println("Found the following profiles: ")
	for n, c := range profiles {
		fmt.Printf("[%d] %s\n", n+1, c.Name)
	}
	input := ask("Please enter a choice for the profile: ", func(input string) bool {
		return validateRange(input, len(profiles))
	})
	r, err := strconv.ParseInt(input, 10, 32)
	// This can't happen because we already validated that this can be parsed
	if err != nil {
		panic(err)
	}
	return &profiles[r-1]
}

func askUsername(p string, s string) string {
	prompt := "Please enter your username"
	if p != "" {
		prompt += fmt.Sprintf(", beginning with: '%s'", p)
	}
	if s != "" {
		if p != "" {
			prompt += "and"
		}
		prompt += fmt.Sprintf(" ending with: '%s'", s)
	}
	prompt += ": "
	username := ask(prompt, func(input string) bool {
		if input == "" {
			fmt.Fprintln(os.Stderr, "Please enter a username that is not empty")
			return false
		}
		if !strings.HasPrefix(input, p) {
			fmt.Fprintf(os.Stderr, "Your username does not begin with: '%s'\n", p)
			return false
		}
		if !strings.HasSuffix(input, s) {
			fmt.Fprintf(os.Stderr, "Your username does not end with: '%s'\n", s)
			return false
		}
		return true
	})

	return username
}

func askPassword() string {
	password := askSecret("Please enter your password: ", func(input string) bool {
		if input == "" {
			fmt.Fprintln(os.Stderr, "Please enter a password that is not empty")
			return false
		}
		return true
	})

	return password
}

func askCertificate(name string, desc string) string {
	panic("todo")
}

func direct(p *instance.Profile) {
	config, err := p.EAPDirect()
	if err != nil {
		log.Fatalf("Could not obtain eap config: %v", err)
	}

	c := configure.Configure{
		UsernameH:    askUsername,
		PasswordH:    askPassword,
		CertificateH: askCertificate,
	}
	err = c.Configure(config)
	if err != nil {
		fmt.Println("failed to configure", err)
	}
}

func redirect(p *instance.Profile) {
	r, err := p.RedirectURI()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to complete the flow, no redirect URI is available")
		return
	}
	err = exec.Command("xdg-open", r).Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to complete the flow, cannot open browser with error: %v\n", err)
		return
	}
	fmt.Println("Opened your browser, please continue the process there")
}

func main() {
	c := discovery.NewCache()
	i, err := c.Instances()
	if err != nil {
		log.Fatalf("failed to get instances from discovery: %v", err)
	}

	chosen := organization(i)
	p := profile(chosen.Profiles)

	switch p.Flow() {
	case instance.DirectFlow:
		direct(p)
	case instance.RedirectFlow:
		redirect(p)
	default:
		fmt.Fprint(os.Stderr, "\nWe do not support OAuth just yet")
	}
	fmt.Println("\nYour eduroam connection has been added to NetworkManager with the name eduroam (from Geteduroam)")
}
