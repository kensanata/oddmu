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

This wiki uses a [Markdown
library](https://github.com/gomarkdown/markdown) to generate the web
pages from Markdown. There are two extensions Oddmu adds to the
library: local links and hashtags.

Local links use double square brackets `[[like this]]`. If you need to
change the link text, you need to use regular Markdown. Don't forget
to [percent-encode](https://en.wikipedia.org/wiki/Percent-encoding)
the link target. Example: `[here](like%20this)`.

Hashtags link to searches for the hashtag. Hashtags are separate from
titles because there is no space after the hash. Use the underscore to
use hashtags consisting of multiple words.

```
# Title

Text

#Tag #Another_Tag
```

The Markdown processor comes with a few extensions, some of which are
enable by default:

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

The template files are the HTML files in the working directory:
`add.html`, `edit.html`, `search.html`, `upload.html` and `view.html`.
Feel free to change the templates and restart the server. The first
change you should make is to replace the email address in `view.html`.
😄

See [Structuring the web
with HTML](https://developer.mozilla.org/en-US/docs/Learn/HTML) to
learn more about HTML.

Modifying the styles in the templates would be another good start to
get a feel for it. See [Learn to style HTML using
CSS](https://developer.mozilla.org/en-US/docs/Learn/CSS) to learn more
about style sheets.

The templates can refer to the following properties of a page:

`{{.Title}}` is the page title. If the page doesn't provide its own
title, the page name is used.

`{{.Name}}` is the page name, escaped for use in URLs. More
specifically, it is URI escaped except for the slashes. The page name
doesn't include the `.md` extension.

`{{.Html}}` is the rendered Markdown, as HTML.

`{{printf "%s" .Body}}` is the Markdown, as a string (the data itself
is a byte array and that's why we need to call `printf`).

For the `search.html` template only:

`{{.Results}}` indicates if there were any search results.

`{{.Items}}` is an array of pages, each containing a search result. A
search result is a page (with the properties seen above). Thus, to
refer to them, you need to use a `{{range .Items}}` … `{{end}}`
construct.

For search results, `{{.Html}}` is the rendered Markdown of a page
summary, as HTML.

`{{.Score}}` is a numerical score for search results.

The `upload.html` template cannot refer to anything.

When calling the `save` action, the page name is take from the URL and
the page content is taken from the `body` form parameter. To
illustrate, here's how to edit a page using `curl`:

```sh
curl --form body="Did you bring a towel?" \
  http://localhost:8080/save/welcome
```

The wiki uses the standard
[html/template](https://pkg.go.dev/html/template) library to do this.
There's more information on writing templates in the documentation for
the [text/template](https://pkg.go.dev/text/template) library.

## Non-English hyphenation

Automatic hyphenation by the browser requires two things: The style
sheet must indicate `hyphen: auto` for an HTML element such as `body`,
and that element must have a `lang` set (usually a two letter language
code such as `de` for German). This happens in the template files,
such as `view.html` and `search.html`.

Oddmu uses the [lingua](github.com/pemistahl/lingua-go) library to
detect languages. If you know that you're only going to use a small
number of languages – or just a single language! – you can set the
environment variable ODDMU_LANGUAGES to a comma-separated list of ISO
639-1 codes, e.g. "en" or "en,de,fr,pt".

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

You can change the port by setting the ODDMU_PORT environment
variable.

## Deploying it using systemd

As root, on your server:

```sh
adduser --system --home /home/oddmu oddmu
```

Copy all the files into `/home/oddmu` to your server: `oddmu`,
`oddmu.service`, and all the template files ending in `.html`.

Edit the `oddmu.service` file. These are the three lines you most
likely have to take care of:

```
ExecStart=/home/oddmu/oddmu
WorkingDirectory=/home/oddmu
Environment="ODDMU_PORT=8080"
Environment="ODDMU_LANGUAGES=en,de"
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
    ProxyPassMatch ^/(search|(view|edit|save|add|append|upload|drop)/(.*))?$ http://localhost:8080/$1
</VirtualHost>
```

First, it manages the domain, getting the necessary certificates. It
redirects regular HTTP traffic from port 80 to port 443. It turns on
the SSL engine for port 443. It proxies the requests for the wiki to
port 8080.

Thus, this is what happens:

* The user tells the browser to visit `transjovian.org`
* The browser sends a request for `http://transjovian.org` (on port 80)
* Apache redirects this to `https://transjovian.org/` by default (now on port 443)
* This is proxied to `http://transjovian.org:8080/` (no encryption, on port 8080)

Restart the server, gracefully:

```
apachectl graceful
```

To serve both HTTP and HTTPS, don't redirect from the first virtual
host to the second – instead just proxy to the wiki like you did for
the second virtual host: use a copy of the `ProxyPassMatch` directive
instead of `RewriteEngine on` and `RewriteRule`.

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

Modify your site configuration and protect the `/edit/`, `/save/`,
`/add/`, `/append/`, `/upload/` and `/drop/` URLs with a password by
adding the following to your `<VirtualHost *:443>` section:

```apache
<LocationMatch "^/(edit|save|add|append|upload|drop)/">
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
`/view/`, `/edit/`, `/save/`, `/add/`, `/append/`, `/upload/`, `/drop/`
or `/search`. For example, create a file called `robots.txt`
containing the following, tellin all robots that they're not welcome.

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
<LocationMatch "^/(edit|save|add|append|upload|drop)/intetebi/">
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

## Virtual hosting

Virtual hosting in this context means that the program serves two
different sites for two different domains from the same machine. Oddmu
doesn't support that, but your webserver does. Therefore, start an
Oddmu instance for every domain name, each listening on a different
port. Then set up your web server such that ever domain acts as a
reverse proxy to a different Oddmu instance.

## Understanding search

The index indexes trigrams. Each group of three characters is a
trigram. A document with content "This is a test" is turned to lower
case and indexed under the trigrams "thi", "his", "is ", "s i", " is",
"is ", "s a", " a ", "a t", " te", "tes", "est".

Each query is split into words and then processed the same way. A
query with the words "this test" is turned to lower case and produces
the trigrams "thi", "his", "tes", "est". This means that the word
order is not considered when searching for documents.

This also means that there is no stemming. Searching for "testing"
won't find "This is a test" because there are no matches for the
trigrams "sti", "tin", "ing".

These trigrams are looked up in the index, resulting in the list of
documents. Each document found is then scored. Each of the following
increases the score by one point:

- the entire phrase matches
- a word matches
- a word matches at the beginning of a word
- a word matches at the end of a word
- a word matches as a whole word

A document with content "This is a test" when searched with the phrase
"this test" therefore gets a score of 8: the entire phrase does not
match but each word gets four points.

Trigrams are sometimes strange: In a text containing the words "main"
and "rail", a search for "mail" returns a match because the trigrams
"mai" and "ail" are found. In this situation, the result has a score
of 0.

## Limitations

Page titles are filenames with `.md` appended. If your filesystem
cannot handle it, it can't be a page name.

The pages are indexed as the server starts and the index is kept in
memory. If you have a ton of pages, this surely wastes a lot of
memory.

Files may not end with a tilde (`~`) – these are backup files.

You cannot edit uploaded files. If you upload a file called
`hello.txt` and attempt to edit it by using `/edit/hello.txt` you will
create a page with the name `hello.txt.md` instead.

You cannot delete uploaded files via the web.

## Bugs

If you spot any, [contact](https://alexschroeder.ch/wiki/Contact) me.

## References

[Writing Web Applications](https://golang.org/doc/articles/wiki/)
provided the initial code for this wiki.

For the proxy stuff, see
[Apache: mod_proxy](https://httpd.apache.org/docs/current/mod/mod_proxy.html).

For the usernames and password stuff, see
[Apache: Authentication and Authorization](https://httpd.apache.org/docs/current/howto/auth.html).
