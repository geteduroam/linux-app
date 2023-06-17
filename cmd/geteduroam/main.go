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

	"github.com/geteduroam/linux-app/internal/discovery"
	"github.com/geteduroam/linux-app/internal/handler"
	"github.com/geteduroam/linux-app/internal/instance"
	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/utils"
	"github.com/ktr0731/go-fuzzyfinder"
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

func main() {
	c := discovery.NewCache()
	i, err := c.Instances()
	if err != nil {
		log.Fatalf("failed to get instances from discovery: %v", err)
	}

	instances := *i

	gotIdx, err := fuzzyfinder.Find(
		instances,
		func(idx int) string {
			raw := instances[idx].Name
			conv, err := utils.RemoveDiacritics(raw)
			if err != nil {
				return raw
			}
			return conv
		},
		fuzzyfinder.WithPreviewWindow(func(idx, w, h int) string {
			if idx == -1 {
				return ""
			}
			// get the profile names
			var pn strings.Builder
			pn.WriteString("Profiles: \n")
			for _, p := range instances[idx].Profiles {
				pn.WriteString("- " + p.Name + "\n")
			}
			return pn.String()
		}),
		fuzzyfinder.WithHeader("Please enter your organization (e.g. SURF)"),
	)
	if err != nil {
		log.Fatalf("Error when searching for an organization: %v", err)
	}
	chosen := instances[gotIdx]
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
	fmt.Println("\nYour eduroam connection has been added to NetworkManager with the name eduroam (from Geteduroam)")
}
