# Oddµ: A minimal wiki

This program runs a wiki. It serves all the Markdown files (ending in
`.md`) into web pages and allows you to edit them.

This is a minimal wiki. There is no version history. It probably makes
sense to only use it as one person or in very small groups.

It's very minimal and only uses Markdown. No wiki extras, so double
square brackets are not a link. If you're used to that, it'll be
strange as you need to repeat the name: `[like this](like this)`.

## Building

```sh
go build
```

## Test

```sh
mkdir wiki
cd wiki
go run ..
```

The program serves the local directory as a wiki on port 8080. Point
your browser to http://localhost:8080/ to get started. This is
equivalent to http://localhost:8080/view/index – the first page
you'll create, most likely.

If you ran it in the source directory, try
http://localhost:8080/view/README – this serves the README file you're
currently reading.

## Deploying it using systemd

As root:

```sh
# on your server
adduser --system --home /home/oddmu oddmu
```

Copy all the files into `/home/oddmu` on your server: `oddmu`, `oddmu.service`, `view.html` and `edit.html`.

Set the ODDMU_PORT environment variable in the `oddmu.service` file (or accept the default, 8080).

Install the service file and enable it:

```sh
ln -s /home/oddmu/oddmu.service /etc/systemd/system/
systemctl enable --now oddmu
```

Check the log:

```sh
journalctl --unit oddmu
```

Follow the log:

```sh
journalctl --follow --unit oddmu
```

Edit the first page using `lynx`:

```sh
lynx http://localhost:8080/view/index
```

## Web server setup

HTTPS is not part of the wiki. You probably want to configure this in
your webserver. If you're using Apache, you might have set up a site
like the following. In my case, that'd be
`/etc/apache2/sites-enabled/500-transjovian.conf`:

```apache
MDomain transjovian.org
MDCertificateAgreement accepted

<VirtualHost *:80>
    ServerName transjovian.org
    RewriteEngine on
    RewriteRule ^/(.*) https://%{HTTP_HOST}/$1 [redirect]
</VirtualHost>
<VirtualHost *:443>
    ServerAdmin alex@alexschroeder.ch
    ServerName transjovian.org
    SSLEngine on

    RewriteEngine on
    RewriteRule ^/$ http://%{HTTP_HOST}:8080/view/index [redirect]
    RewriteRule ^/(view|edit|save)/(.*) http://%{HTTP_HOST}:8080/$1/$2 [proxy]
</VirtualHost>
```

First, it manages the domain, getting the necessary certificates. It
redirects regular HTTP traffic from port 80 to port 443. It turns on
the SSL engine for port 443. It redirects `/` to `/view/index` and any
path that starts with `/view/`, `/edit/` or `/save/` is proxied to
port 8080 where the Oddmu program can handle it.

Thus, this is what happens:

* The user tells the browser to visit `http://transjovian.org` (on port 80)
* Apache redirects this to `http://transjovian.org/` by default (still on port 80)
* Our first virtual host redirects this to `https://transjovian.org/` (encrypted, on port 443)
* Our second virtual host redirects this to `https://transjovian.org/wiki/view/index` (still on port 443)
* This is proxied to `http://transjovian.org:8080/view/index` (no on port 8080, without encryption)
* The wiki converts `index.md` to HTML, adds it to the template, and serves it.

## Access

Access control is not part of the wiki. By default, the wiki is
editable by all. This is most likely not what you want unless you're
running it stand-alone, unconnected to the Internet.

You probably want to configure this in your webserver. If you're using
Apache, you might have set up a site like the following.

Create a new password file called `.htpasswd` and add the user "alex":

```sh
cd /home/oddmu
htpasswd -c .htpasswd alex
```

To add more users, don't use the `-c` option or you will overwrite it!

To add another user:

```sh
htpasswd .htpasswd berta
```

To delete remove a user:

```sh
htpasswd -D .htpasswd berta
```

Modify your site configuration and protect the `/edit/` and `/save/`
URLs with a password by adding the following to your `<VirtualHost
*:443>` section:

```apache
<LocationMatch "^/(edit|save)/">
  AuthType Basic
  AuthName "Password Required"
  AuthUserFile /home/oddmu/.htpasswd
  Require valid-user
</LocationMatch>
```

## Configuration

Feel free to change the templates `view.html` and `edit.html` and
restart the server. Modifying the styles in the templates would be a
good start.

### No automatic titles

You can remove the auto-generated titles from the files, for example.
If your Markdown files start with a level 1 title, then edit
`view.html` and remove the line that says `<h1>{{.Title}}</h1>` (this
is what people see when reading the page). Optionally also remove the
line that says `<title>{{.Title}}</title>` (this is what gets used for
tabs and bookmarks).

### Serve static files

If you want to serve static files as well, add a document root to your
webserver configuration. Using Apache, for example:

```apache
DocumentRoot /home/oddmu/static
<Directory /home/oddmu/static>
    Require all granted
</Directory>
```

Create this directory, making sure to give it a permission that your
webserver can read (world readable file, world readable and executable
directory). Populate it with files. For example, create a file called
`robots.txt` containing the following, tellin all robots that they're
not welcome.

```text
User-agent: *
Disallow: /
```

You site now serves `/robots.txt` without interfering with the wiki,
and without needing a wiki page.

[Wikipedia](https://en.wikipedia.org/wiki/Robot_exclusion_standard)
has more information.

All you have make sure is that none of the static files look like the
wiki paths `/view/`, `/edit/` or `/save/`.

## Customization (with recompilation)

The Markdown parser can be customized and
[extensions](https://pkg.go.dev/github.com/gomarkdown/markdown/parser#Extensions)
can be added. There's an example in the
[usage](https://github.com/gomarkdown/markdown#usage) section. You'll
need to make changes to the `viewHandler` yourself.

### Render Gemtext

In a first approximation, Gemtext is valid Markdown except for the
rocket links (`=>`). Here's how to modify the `loadPage` so that a
`.gmi` file is loaded if no `.md` is found, and the rocket links are
translated into Markdown:

```go
func loadPage(title string) (*Page, error) {
	filename := title + ".md"
	body, err := os.ReadFile(filename)
	if err == nil {
		return &Page{Title: title, Body: body}, nil
	}
	filename = title + ".gmi"
	body, err = os.ReadFile(filename)
	if err == nil {
		re := regexp.MustCompile(`(?m)^=>\s*(\S+)\s+(.+)`)
		body = []byte(re.ReplaceAllString(string(body), `* [$2]($1)`))
		return &Page{Title: title, Body: body}, nil
	}
	return nil, err
}
```

There is a small problem, however: By default, Markdown expects an
empty line before a list begins. The following change to `viewHandler`
uses the `NoEmptyLineBeforeBlock` extension for the parser:

```go
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	extensions := parser.CommonExtensions | parser.NoEmptyLineBeforeBlock
	markdownParser := parser.NewWithExtensions(extensions)
	flags := html.CommonFlags
	opts := html.RendererOptions{
		Flags: flags,
	}
	htmlRenderer := html.NewRenderer(opts)
	maybeUnsafeHTML := markdown.ToHTML(p.Body, markdownParser, htmlRenderer)
	html := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)
	p.Html = template.HTML(html);
	renderTemplate(w, "view", p)
}
```

## Limitations

Page titles are filenames with `.md` appended. If your filesystem
cannot handle it, it can't be a page title. Specifically, *no slashes*
in filenames.

## References

[Writing Web Applications](https://golang.org/doc/articles/wiki/)
provided the initial code for this wiki.

For the proxy stuff, see
[Apache: mod_proxy](https://httpd.apache.org/docs/current/mod/mod_proxy.html).

For the usernames and password stuff, see
[Apache: Authentication and Authorization](https://httpd.apache.org/docs/current/howto/auth.html).
