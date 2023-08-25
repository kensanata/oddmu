# Oddµ: A minimal wiki

This program runs a wiki. It serves all the Markdown files (ending in
`.md`) into web pages and allows you to edit them. If your files don't
provide their own title (`# title`), the file name (without `.md`) is
used for the title. Subdirectories are created as necessary.

This is a minimal wiki. There is no version history. It's well suited
as a *secondary* medium: collaboration and conversation happens
elsewhere, in chat, on social media. The wiki serves as the text
repository that results from these discussions.

The wiki lists no recent changes. The expectation is that the people
that care were involved in the discussions beforehand.

The wiki also produces no feed. The assumption is that announcements
are made on social media: blogs, news aggregators, discussion forums,
the fediverse, but humans. There is no need for bots.

As you'll see below, the idea is that the webserver handles as many
tasks as possible. It logs requests, does rate limiting, handles
encryption, gets the certificates, and so on. The web server acts as a
reverse proxy and the wiki ends up being a content management system
with almost no structure – or endless malleability, depending on your
point of view.

And last but not least: µ is the letter mu, so Oddµ is usually written
Oddmu. 🙃

## Markdown

This wiki uses Markdown. There is no additional wiki markup, most
importantly double square brackets are not a link. If you're used to
that, it'll be strange as you need to repeat the name: `[like
this](like this)`. The Markdown processor comes with a few extensions,
some of which are enable by default:

* emphasis markers inside words are ignored
* tables are supported
* fenced code blocks are supported
* autolinking of "naked" URLs are supported
* strikethrough using two tildes is supported (`~~like this~~`)
* it is strict about prefix heading rules
* you can specify an id for headings (`{#id}`)
* trailing backslashes turn into line breaks
* definition lists are supported
* MathJax is supported (but needs a separte setup)

See the section on
[extensions](https://github.com/gomarkdown/markdown#extensions) in the
Markdown library for information on the various extensions.

A table with footers and a columnspan:

```text
Name    | Age
--------|------
Bob     ||
Alice   | 23
========|======
Total   | 23
```

A definition list:

```text
Cat
: Fluffy animal everyone likes

Internet
: Vector of transmission for pictures of cats
```

## Templates

Feel free to change the templates `view.html` and `edit.html` and
restart the server. Modifying the styles in the templates would be a
good start to get a feel for it.

The first change you should make is to replace the email address in
`view.html`. 😄

The templates can refer to the following properties of a page:

`{{.Title}}` is the page title. If the page doesn't provide its own
title, the page name is used.

`{{.Name}}` is the page name. The page name doesn't include the `.md`
extension.

`{{.Html}}` is the rendered Markdown, as HTML.

`{{printf "%s" .Body}}` is the Markdown, as a string (the data itself
is a byte array and that's why we need to call `printf`).

When calling the `save` action, the page name is take from the URL and
the page content is taken from the `body` form parameter. To
illustrate, here's how to edit a page using `curl`:

```sh
curl --form body="Did you bring a towel?" \
  http://localhost:8080/save/welcome
```

## Non-English hyphenation

Automatic hyphenation by the browser requires two things: The style
sheet must indicate `hyphen: auto` for an HTML element such as `body`,
and that element must have a `lang` set (usually a two letter language
code such as `de` for German). This happens in the template files,
such as `view.html` and `search.html`.

If have languages in different languages, the problem is that 

## Building

```sh
go build
```

## Test

The working directory is where pages are saved and where templates are
loaded from. You need a copy of the template files in this directory.
Here's how to start it in the source directory:

```sh
go run .
```

The program serves the local directory as a wiki on port 8080. Point
your browser to http://localhost:8080/ to get started. This is
equivalent to http://localhost:8080/view/index – the first page
you'll create, most likely.

If you ran it in the source directory, try
http://localhost:8080/view/README – this serves the README file you're
currently reading.

## Deploying it using systemd

As root, on your server:

```sh
adduser --system --home /home/oddmu oddmu
```

Copy all the files into `/home/oddmu` to your server: `oddmu`,
`oddmu.service`, `view.html` and `edit.html`.

Edit the `oddmu.service` file. These are the three lines you most
likely have to take care of:

```
ExecStart=/home/oddmu/oddmu
WorkingDirectory=/home/oddmu
Environment="ODDMU_PORT=8080"
```

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
your webserver. I guess you could use stunnel, too. If you're using
Apache, you might have set up a site like I did, below. In my case,
that'd be `/etc/apache2/sites-enabled/500-transjovian.conf`:

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
    RewriteRule ^/(view|edit|save|search)/(.*) http://%{HTTP_HOST}:8080/$1/$2 [proxy]
</VirtualHost>
```

First, it manages the domain, getting the necessary certificates. It
redirects regular HTTP traffic from port 80 to port 443. It turns on
the SSL engine for port 443. It redirects `/` to `/view/index` and any
path that starts with `/view/`, `/edit/`, `/save/` or `/search/` is
proxied to port 8080 where the Oddmu program can handle it.

Thus, this is what happens:

* The user tells the browser to visit `http://transjovian.org` (on port 80)
* Apache redirects this to `http://transjovian.org/` by default (still on port 80)
* Our first virtual host redirects this to `https://transjovian.org/` (encrypted, on port 443)
* Our second virtual host redirects this to `https://transjovian.org/wiki/view/index` (still on port 443)
* This is proxied to `http://transjovian.org:8080/view/index` (no on port 8080, without encryption)
* The wiki converts `index.md` to HTML, adds it to the template, and serves it.

Restart the server, gracefully:

```
apachectl graceful
```

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

## Serve static files

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
directory). Populate it with files.

Make sure that none of the static files look like the wiki paths
`/view/`, `/edit/`, `/save/` or `/search/`.

For example, create a file called `robots.txt` containing the
following, tellin all robots that they're not welcome.

```text
User-agent: *
Disallow: /
```

You site now serves `/robots.txt` without interfering with the wiki,
and without needing a wiki page.

[Wikipedia](https://en.wikipedia.org/wiki/Robot_exclusion_standard)
has more information.

## Different logins for different access rights

What if you have a site with various subdirectories and each
subdirectory is for a different group of friends? You can set this up
using your webserver. One way to do this is to require specific
usernames (which must have a password in the password file mentioned
above.

This requires a valid login by the user "alex" or "berta":

```apache
<LocationMatch "^/(edit|save)/intetebi/">
  Require user alex berta
</LocationMatch>
```

## Private wikis

Based on the above, you can prevent people from reading the wiki, too.
The `LocationMatch` must cover the `/view/` URLs. In order to protect
*everything*, use a [Location directive](https://httpd.apache.org/docs/current/mod/core.html#location)
that matches everything:

```apache
<Location />
  AuthType Basic
  AuthName "Password Required"
  AuthUserFile /home/oddmu/.htpasswd
  Require valid-user
</Location>
```

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
func loadPage(name string) (*Page, error) {
	filename := name + ".md"
	body, err := os.ReadFile(filename)
	if err == nil {
		return &Page{Title: name, Name: name, Body: body}, nil
	}
	filename = name + ".gmi"
	body, err = os.ReadFile(filename)
	if err == nil {
		return &Page{Title: name, Name: name, Body: body}, nil
	}
	return nil, err
}
```

There is a small problem, however: By default, Markdown expects an
empty line before a list begins. The following change to `renderHtml`
uses the `NoEmptyLineBeforeBlock` extension for the parser:

```go
func (p* Page) renderHtml() {
    // Here is where a new extension is added!
	extensions := parser.CommonExtensions | parser.NoEmptyLineBeforeBlock
	markdownParser := parser.NewWithExtensions(extensions)
	maybeUnsafeHTML := markdown.ToHTML(p.Body, markdownParser, nil)
	html := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)
	p.Html = template.HTML(html);
}
```

## Limitations

Page titles are filenames with `.md` appended. If your filesystem
cannot handle it, it can't be a page title. Specifically, *no slashes*
in filenames.

The pages are indexed as the server starts and the index is kept in
memory. If you have a ton of pages, this surely wastes a lot of
memory.

## References

[Writing Web Applications](https://golang.org/doc/articles/wiki/)
provided the initial code for this wiki.

For the proxy stuff, see
[Apache: mod_proxy](https://httpd.apache.org/docs/current/mod/mod_proxy.html).

For the usernames and password stuff, see
[Apache: Authentication and Authorization](https://httpd.apache.org/docs/current/howto/auth.html).
