package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/geteduroam/linux-app/internal/discovery"
	"github.com/geteduroam/linux-app/internal/handler"
	"github.com/geteduroam/linux-app/internal/instance"
	"github.com/geteduroam/linux-app/internal/network"
)

// askSecret is a tweak of thee 'ask' function that uses golang.org/x/term to read a secret securely
// The prompt is the text to show e.g. "enter something: "
// Validator is the function that checks if the secret is valid
func askSecret(prompt string, validator func(input string) bool) string {
	for {
		fmt.Print(prompt)
		pwd, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read password: %v", err)
			continue
		}
		fmt.Println()
		// get the password as a string
		pwdS := string(pwd)
		if validator(pwdS) {
			return pwdS
		}
	}
}

// ask asks the user for an input
// The prompt is the text to show e.g. "enter something: "
// Validator is the function that checks if the input is valid
// It loops until a valid input is given
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

// filteredOrganizations gets the instances as filtered by the user
func filteredOrganizations(orgs *instance.Instances, q string) (f *instance.Instances) {
	for {
		x := ask(q, func(x string) bool {
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

// validateRange validates if the input is in the range of 1-n (inclusive)
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

// organization gets an organization/instance from the user
func organization(orgs *instance.Instances) *instance.Instance {
	_, h, err := term.GetSize(0)
	if err != nil {
		fmt.Println("Could not get height")
		h = 10
	}
	f := orgs
	f = filteredOrganizations(f, "Please enter your organization (e.g. SURF): ")
	for {
		if len(*f) > h {
			for _, c := range *f {
				fmt.Printf("%s\n", c.Name)
			}
			fmt.Println("\nList is long...")
			f = filteredOrganizations(f, "Please refine your search: ")
		} else {
			break
		}
	}
	fmt.Println("\nFound the following matches: ")
	for n, c := range *f {
		fmt.Printf("[%d] %s\n", n+1, c.Name)
	}
	input := ask("\nPlease enter a choice for the organisation: ", func(input string) bool {
		return validateRange(input, len(*f))
	})
	r, err := strconv.ParseInt(input, 10, 32)
	// This can't happen because we already validated that this can be parsed
	if err != nil {
		panic(err)
	}
	return &(*f)[r-1]
}

// profile gets a profile for a list of profiles by asking the user one if there are multiple
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

// askUsername asks the user for the username
// p is the prefix for which the username must start
// s is the suffix for which the username must end
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

// askPassword asks the user for a password
func askPassword() string {
	validator := func(input string) bool {
		if input == "" {
			fmt.Fprintln(os.Stderr, "Please enter a password that is not empty")
			return false
		}
		return true
	}

	password1 := ""
	password2 := ""

	for next := true; next; next = password1 != password2 {
		password1 = askSecret("Please enter your password: ", validator)
		password2 = askSecret("Please confirm your password: ", validator)

		if password1 != password2 {
			fmt.Fprintln(os.Stderr, "\nPasswords do not match, try again")
		}
	}

	return password1
}

// askCredentials asks the user for credentials
// It returns the username and password
func askCredentials(c network.Credentials, pi network.ProviderInfo) (string, string) {
	fmt.Println("\nOrganization info:")
	fmt.Println(" Title:", pi.Name)
	fmt.Println(" Description:", pi.Description)
	if pi.Helpdesk.Email != "" {
		fmt.Println(" Helpdesk e-mail:", pi.Helpdesk.Email)
	}
	if pi.Helpdesk.Phone != "" {
		fmt.Println(" Helpdesk phone number:", pi.Helpdesk.Phone)
	}
	if pi.Helpdesk.Web != "" {
		fmt.Println(" Helpdesk URL:", pi.Helpdesk.Web)
	}

	username := c.Username
	password := c.Password
	if c.Username == "" {
		username = askUsername(c.Prefix, c.Suffix)
	}
	if c.Password == "" {
		password = askPassword()
	}
	return username, password
}

// askCertificate asks the user for a certificate
// This is used in the TLS/OAuth flow
func askCertificate(_ string, _ network.ProviderInfo) string {
	panic("todo")
}

// file does the flow when the file has been obtained
func file(metadata []byte) (err error) {
	h := handler.Handlers{
		CredentialsH: askCredentials,
		CertificateH: askCertificate,
	}

	// Configure the network further.
	// The handlers will take care of the rest
	return h.Configure(metadata)
}

// direct does the handling for the direct flow
func direct(p *instance.Profile) {
	config, err := p.EAPDirect()
	if err != nil {
		log.Fatalf("Could not obtain eap config: %v", err)
	}

	err = file(config)
	if err != nil {
		log.Fatalf("Failed to configure the connection using the metadata: %v", err)
	}
}

// redirect does the handling for the redirect flow
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

// oauth does the handling for the OAuth flow
func oauth(p *instance.Profile) {
	config, err := p.EAPOAuth()
	if err != nil {
		log.Fatalf("Could not obtain eap config with OAuth: %v", err)
	}

	err = file(config)
	if err != nil {
		log.Fatalf("Failed to configure the connection using the OAuth metadata: %v", err)
	}
}

func doLocal(filename string) {
	    b, err := os.ReadFile(filename)
	    if err != nil {
		    log.Fatalf("Failed to read local file: %v", err)
	    }
	    err = file(b)
	    if err != nil {
		    log.Fatalf("Failed to configure the connection using the metadata: %v", err)
	    }
}

func doDiscovery() {
	c := discovery.NewCache()
	i, err := c.Instances()
	if err != nil {
		log.Fatalf("failed to get instances from discovery: %v", err)
	}

	chosen := organization(i)
	p := profile(chosen.Profiles)

	// TODO: This switch statement should probably be moved to the profile code
	// By providing an "EAP" method on profile
	switch p.Flow() {
	case instance.DirectFlow:
		direct(p)
	case instance.RedirectFlow:
		redirect(p)
		return
	case instance.OAuthFlow:
		oauth(p)
	}
}

// findVersion gets the version in the following order:
// - Gets a release version if it detects it is a release
// - Gets the commit using debug info
// - Returns a default
func findVersion() string {
	// TODO: Support a release version too
	if dbg, ok := debug.ReadBuildInfo(); ok {
		for _, s := range dbg.Settings {
			if s.Key == "vcs.revision" {
				return "Git checkout " + s.Value
			}
		}
	}
	return "0.0 (unknown)"
}

const usage = `Usage of %s:
  -h, --help			Prints this help information
  --version			Prints version information
  -l <file>, --local=<file>	The path to a local EAP metadata file
`

func main() {
	var help bool
	var version bool
	var local string
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&version, "version", false, "Show version")
	flag.StringVar(&local, "local", "", "The path to a local EAP metadata file")
	flag.StringVar(&local, "l", "", "The path to a local EAP metadata file")
	flag.Usage = func() { fmt.Printf(usage, filepath.Base(os.Args[0])) }
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	if version {
		fmt.Println("Version:", findVersion())
		return
	}
	if local != "" {
		doLocal(local)
	} else {
		doDiscovery()
	}
	fmt.Println("\nYour eduroam connection has been added to NetworkManager with the name eduroam (from Geteduroam)")
}
