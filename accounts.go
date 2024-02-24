package main

import (
	"encoding/json"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

// useWebfinger indicates whether Oddmu looks up the profile pages of fediverse accounts. To enable this, set the
// environment variable ODDMU_WEBFINGER to "1".
var useWebfinger = false

// accountStore controlls access to the usernames. Make sure to lock and unlock as appropriate.
type accountStore struct {
	sync.RWMutex

	// uris is a map, mapping account names likes "@alex@alexschroeder.ch" to URIs like
	// "https://social.alexschroeder.ch/@alex".
	uris map[string]string
}

// accounts holds the global mapping of accounts to profile URIs.
var accounts accountStore

// This is called once at startup and therefore does not need to be locked. On every restart, this map starts empty and
// is slowly repopulated as pages are visited.
func init() {
	if os.Getenv("ODDMU_WEBFINGER") == "1" {
		accounts.uris = make(map[string]string)
		useWebfinger = true
	}
}

// accountLink links a social media accountLink like @accountLink@domain to a profile page like https://domain/user/accountLink. Any
// accountLink seen for the first time uses a best guess profile URI. It is also looked up using webfinger, in parallel. See
// lookUpAccountUri. If the lookup succeeds, the best guess is replaced with the new URI so on subsequent requests, the
// URI is correct.
func accountLink(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
	data = data[offset:]
	i := 1 // skip @ of username
	n := len(data)
	d := 0
	for i < n && (data[i] >= 'a' && data[i] <= 'z' ||
		data[i] >= 'A' && data[i] <= 'Z' ||
		data[i] >= '0' && data[i] <= '9' ||
		data[i] == '@' ||
		data[i] == '.' ||
		data[i] == '_' ||
		data[i] == '-') {
		if data[i] == '@' {
			if d != 0 {
				// more than one @ is invalid
				return 0, nil
			} else {
				d = i + 1 // skip @ of domain
			}
		}
		i++
	}
	for i > 1 && (data[i-1] == '.' ||
		data[i-1] == '-') {
		i--
	}
	if i == 0 || d == 0 {
		return 0, nil
	}
	user := data[0 : d-1] // includes @
	domain := data[d:i]   // excludes @
	account := data[1:i]  // excludes @
	accounts.RLock()
	uri, ok := accounts.uris[string(account)]
	defer accounts.RUnlock()
	if !ok {
		log.Printf("Looking up %s\n", account)
		uri = "https://" + string(domain) + "/users/" + string(user[1:])
		accounts.uris[string(account)] = uri // prevent more lookings
		go lookUpAccountUri(string(account), string(domain))
	}
	link := &ast.Link{
		AdditionalAttributes: []string{`class="account"`},
		Destination:          []byte(uri),
		Title:                data[0:i],
	}
	ast.AppendChild(link, &ast.Text{Leaf: ast.Leaf{Literal: data[0 : d-1]}})
	return i, link
}

// lookUpAccountUri is called for accounts that haven't been seen before. It calls webfinger and parses the JSON. If
// possible, it extracts the link to the profile page and replaces the entry in accounts.
func lookUpAccountUri(account, domain string) {
	uri := "https://" + domain + "/.well-known/webfinger"
	resp, err := http.Get(uri + "?resource=acct:" + account)
	if err != nil {
		log.Printf("Failed to look up %s: %s", account, err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read from %s: %s", account, err)
		return
	}
	var wf webFinger
	err = json.Unmarshal([]byte(body), &wf)
	if err != nil {
		log.Printf("Failed to parse the JSON from %s: %s", account, err)
		return
	}
	uri, err = parseWebFinger(body)
	if err != nil {
		log.Printf("Could not find profile URI for %s: %s", account, err)
	}
	log.Printf("Found profile for %s: %s", account, uri)
	accounts.Lock()
	defer accounts.Unlock()
	accounts.uris[account] = uri
}

// link a link in the WebFinger JSON.
type link struct {
	Rel  string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}

// webFinger is a structure used to unmarshall JSON.
type webFinger struct {
	Subject string   `json:"subject"`
	Aliases []string `json:"aliases"`
	Links   []link   `json:"links"`
}

// parseWebFinger parses the web finger JSON and returns the profile page URI. For unmarshalling the JSON, it uses the
// Link and WebFinger structs.
func parseWebFinger(body []byte) (string, error) {
	var wf webFinger
	err := json.Unmarshal(body, &wf)
	if err != nil {
		return "", err
	}
	for _, link := range wf.Links {
		if link.Rel == "http://webfinger.net/rel/profile-page" &&
			link.Type == "text/html" {
			return link.Href, nil
		}
	}
	return "", err
}
