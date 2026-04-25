package install

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// netrcTransport is an http.RoundTripper that reads credentials from a netrc
// file and adds an Authorization header for matching hosts.
type netrcTransport struct {
	base      http.RoundTripper
	netrcPath string
}

// RoundTrip implements http.RoundTripper. It looks up the request host in the
// netrc file and adds basic auth if credentials are found.
func (t *netrcTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	login, password, err := lookupNetrc(t.netrcPath, req.URL.Hostname())
	if err == nil && login != "" {
		req = req.Clone(req.Context())
		req.SetBasicAuth(login, password)
	}
	return t.base.RoundTrip(req)
}

// lookupNetrc parses a netrc file and returns the login and password for the
// given machine. Returns empty strings if no match is found.
func lookupNetrc(path, machine string) (login, password string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("opening netrc file %q: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var inMachine bool
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		tokens := strings.Fields(line)
		for i := 0; i < len(tokens); i++ {
			switch tokens[i] {
			case "machine":
				i++
				if i < len(tokens) {
					inMachine = tokens[i] == machine
				}
			case "default":
				if login == "" {
					inMachine = true
				}
			case "login":
				i++
				if i < len(tokens) && inMachine {
					login = tokens[i]
				}
			case "password":
				i++
				if i < len(tokens) && inMachine {
					password = tokens[i]
				}
			}
		}

		if inMachine && login != "" && password != "" {
			return login, password, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("reading netrc file: %w", err)
	}
	return login, password, nil
}
