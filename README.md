# Oddµ: A minimal wiki

This program helps you run a minimal wiki. There is no version
history. It's well suited as a *secondary* medium: collaboration and
conversation happens elsewhere, in chat, on social media. The wiki
serves as the text repository that results from these discussions.

If you're the only user and it just runs on your laptop, then you can
think of it as a [memex](https://en.wikipedia.org/wiki/Memex), a
memory extender.

Oddµ can be used as a web server behind a reverse proxy such as Apache
or it can be used as a static site generator.

When Oddµ runs as a web server, it serves all the Markdown files
(ending in `.md`) as web pages and allows you to edit them.

If your files don't provide their own title (`# title`), the file name
(without `.md`) is used for the title. Subdirectories are created as
necessary.

Oddµ uses a [Markdown library](https://github.com/gomarkdown/markdown)
to generate the web pages from Markdown. Oddmu adds the following
extensions: local links `[[like this]]`, hashtags `#Like_This` and
fediverse account links like `@alex@alexschroeder.ch`.

The [lingua](https://github.com/pemistahl/lingua-go) library detects
languages in order to get hyphenation right.

The standard [html/template](https://pkg.go.dev/html/template) library
is used to generate HTML.

## Documentation

This project uses man(1) pages. They are generated from text files
using [scdoc](https://git.sr.ht/~sircmpwn/scdoc). These are the files
available:

[oddmu(1)](/oddmu.git/blob/main/man/oddmu.1.txt): This man page has a
short introduction to Oddmu, its configuration via templates and
environment variables, plus points to the other man pages.

[oddmu(5)](/oddmu.git/blob/main/man/oddmu.5.txt): This man page talks
about the Markdown and includes some examples for the non-standard
features such as table markup. It also talks about the Oddmu
extensions to Markdown: wiki links, hashtags and fediverse account
links. Local links must use percent encoding for page names so there
is a section about percent encoding. The man page also explains how
feeds are generated.

[oddmu-list(1)](/oddmu.git/blob/main/man/oddmu-list.1.txt): This man
page documents the "list" subcommand which you can use to get page
names and page titles.

[oddmu-search(1)](/oddmu.git/blob/main/man/oddmu-search.1.txt): This
man page documents the "search" subcommand which you can use to build
indexes – lists of page links. These are important for feeds.

[oddmu-search(7)](/oddmu.git/blob/main/man/oddmu-search.7.txt): This
man page documents how search and scoring work.

[oddmu-replace(1)](/oddmu.git/blob/main/man/oddmu-replace.1.txt): This
man page documents the "replace" subcommand to make mass changes to
the files much like find(1), grep(1) and sed(1) or perl(1).

[oddmu-missing(1)](/oddmu.git/blob/main/man/oddmu-missing.1.txt): This
man page documents the "missing" subcommand to list local links that
don't point to any existing pages or files.

[oddmu-html(1)](/oddmu.git/blob/main/man/oddmu-html.1.txt): This man
page documents the "html" subcommand to generate HTML from Markdown
pages from the command line.

[oddmu-static(1)](/oddmu.git/blob/main/man/oddmu-static.1.txt): This
man page documents the "static" subcommand to generate an entire
static website from the command line, avoiding the need to run Oddmu
as a server. Also great for archiving.

[oddmu-notify(1)](/oddmu.git/blob/main/man/oddmu-notify.1.txt): This
man page documents the "notify" subcommand to add links to hashtag
pages, index and changes for a given page. This is useful when you
edit the Markdown files locally.

[oddmu-templates(5)](/oddmu.git/blob/main/man/oddmu-templates.5.txt):
This man page documents how the templates can be changed (how they
*must* be changed) and lists the attributes available for the various
templates.

[oddmu-apache(5)](/oddmu.git/blob/main/man/oddmu-apache.5.txt): This
man page documents how to set up the web server for various common
tasks such as using logins to limit what visitors can edit.

[oddmu.service(5)](/oddmu.git/blob/main/man/oddmu.service.5.txt): This
man page documents how to setup a systemd unit and have it manage
Oddmu. “Great configurability brings great burdens.”

## Building

```sh
go build
```

## Running

The working directory is where pages are saved and where templates are
loaded from. You need a copy of the template files in this directory.
Here's how to start it in the source directory:

```sh
go run .
```

The program serves the local directory as a wiki on port 8080. Point
your browser to http://localhost:8080/ to use it.

## Bugs

If you spot any, [contact](https://alexschroeder.ch/wiki/Contact) me.

## Source

If you're interested in making changes to the code, here's a
high-level introduction to the various source files.

- *_test.go are the test files; a few library functions are defined in
  wiki_test.go.
- *_cmd.go are the files implementing the various subcommands with
  matching names
- accounts.go implements the webfinger code to fetch fediverse account
  link destinations with the URI provided by webfinger
- add_append.go implements the /add and /append handlers
- diff.go implements the /diff handler
- edit_save.go implements the /edit and /save handlers
- feed.go implements the feed for a page based on the links it lists
- highlight.go implements the bold tags for matches when showing
  search results
- index.go implements the index of all the hashtags
- languages.go implements the language detection
- page.go implements the page loading and saving
- parser.go implements the Markdown parsing
- score.go implements the page scoring when showing search results
- search.go implements the /search handler
- snippets.go implements the page summaries for search results
- tokenizer.go implements the various tokenizers used
- upload_drop.go implements the /upload and /drop handlers
- view.go implements the /view handler
- wiki.go implements the main function

## References

[Writing Web Applications](https://golang.org/doc/articles/wiki/)
provided the initial code for this wiki.
