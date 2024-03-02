# Oddµ: A minimal wiki

This program helps you run a minimal wiki, blog, digital garden, memex
or Zettelkasten. There is no version history.

It's well suited as a self-hosted, single-user web application: the is
no need for collaboration on the site. Links and email connect you to
the rest of the net. The wiki can be public or private. Perhaps it
just runs on your local machine, unreachable from the Internet.

It's well suited as a secondary medium for a group: collaboration and
conversation happens elsewhere, in chat, on social media. The wiki
serves as the text repository that results from these discussions.

It's well suited as a simple static site generator. There are no
plugins.

Oddµ can be used as a web server behind a reverse proxy such as Apache
or it can be used as a static site generator.

When Oddµ runs as a web server, it serves all the Markdown files
(ending in `.md`) as web pages and allows you to edit them.

Oddmu adds the following extensions to Markdown: local links `[[like
this]]`, hashtags `#Like_This` and fediverse account links like
`@alex@alexschroeder.ch`.

If your files don't provide their own title (`# title`), the file name
(without `.md`) is used for the title. Subdirectories are created as
necessary.

Other files can be uploaded and images can be resized when they are
uploaded.

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

[oddmu-releases(7)](/oddmu.git/blob/main/man/oddmu-releases.7.txt):
This man page lists all the Oddmu versions and their user-visible
changes.

[oddmu-releases(7)](/oddmu.git/blob/main/man/oddmu-releases.7.txt):
This man page lists all the Oddmu versions and their user-visible
changes.

[oddmu-version(1)](/oddmu.git/blob/main/man/oddmu-version.1.txt): This
man page documents the "version" subcommand which you can use to get
installed Oddmu version.

[oddmu-list(1)](/oddmu.git/blob/main/man/oddmu-list.1.txt): This man
page documents the "list" subcommand which you can use to get page
names and page titles.

[oddmu-search(1)](/oddmu.git/blob/main/man/oddmu-search.1.txt): This
man page documents the "search" subcommand which you can use to build
indexes – lists of page links. These are important for feeds.

[oddmu-search(7)](/oddmu.git/blob/main/man/oddmu-search.7.txt): This
man page documents how search and scoring work.

[oddmu-filter(7)](/oddmu.git/blob/main/man/oddmu-filter.7.txt): This
man page documents how to exclude subdirectories from search and
archiving.

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
man page documents how to set up the Apache web server for various
common tasks such as using logins to limit what visitors can edit.

[oddmu-nginx(5)](/oddmu.git/blob/main/man/oddmu-nginx.5.txt): This man
page documents how to set up the freenginx web server for various
common tasks such as using logins to limit what visitors can edit.

[oddmu.service(5)](/oddmu.git/blob/main/man/oddmu.service.5.txt): This
man page documents how to setup a systemd unit and have it manage
Oddmu. “Great configurability brings great burdens.”

## Building

To build the binary:

```sh
go build
```

The man pages are already built. If you want to rebuild them, you need
to have [scdoc](https://git.sr.ht/~sircmpwn/scdoc) installed.

```sh
make docs
```

The `Makefile` in the `man` directory has targets to create Markdown
and HTML files.

The HEIC library uses C code and prevents cross-compilation.

As the repository changed URLs a few times (from GitHub, to
self-hosted using `cgit` to self-hosted using `legit`), there is no
way to install it using `go install`. You need to `git clone` the
repository and build it locally.

## Running

The working directory is where pages are saved and where templates are
loaded from. You need a copy of the template files in this directory.

Here's how to build and run straight from the source directory:

```sh
go run .
```

The program serves the local directory as a wiki on port 8080. Point
your browser to http://localhost:8080/ to use it.

Once the `oddmu` binary is built, you can run it instead:

```sh
./oddmu
```

To read the main man page witihout installing Oddmu:

```sh
man -l man/oddmu.1
```

## Installing

This installs `oddmu` into `$HOME/.local/bin` and the manual pages
into `$HOME/.local/share/man/`.

```sh
make install
```

To install it elsewhere, here's an example using [GNU
Stow](https://www.gnu.org/software/stow/) to install it into
`/usr/local/stow` in a way that allows you to uninstall it later:

```sh
sudo mkdir /usr/local/stow/oddmu
sudo make install PREFIX=/usr/local/stow/oddmu/
cd /usr/local/stow
sudo stow oddmu
```

## Hacking

If you're interested in making changes to the code, here's a
high-level introduction to the various source files.

- `*_test.go` are the test files; a few library functions are defined
  in `wiki_test.go`.
- `*_cmd.go` are the files implementing the various subcommands with
  matching names
- `accounts.go` implements the webfinger code to fetch fediverse
  account link destinations with the URI provided by webfinger
- `add_append.go` implements the `/add` and `/append` handlers
- `archive.go` implements the `/archive` handler
- `changes.go` implements the "notifications": the automatic addition
  of links to index, changes and hashtag files when pages are edited
- `diff.go` implements the `/diff` handler
- `edit_save.go` implements the `/edit` and `/save` handlers
- `feed.go` implements the feed for a page based on the links it lists
- `highlight.go` implements the bold tags for matches when showing
  search results
- `index.go` implements the index of all the hashtags
- `languages.go` implements the language detection
- `page.go` implements the page loading and saving
- `parser.go` implements the Markdown parsing
- `score.go` implements the page scoring when showing search results
- `search.go` implements the `/search` handler
- `snippets.go` implements the page summaries for search results
- `templates.go` implements template loading and reloading
- `tokenizer.go` implements the various tokenizers used
- `upload_drop.go` implements the `/upload` and `/drop` handlers
- `view.go` implements the `/view` handler
- `watch.go` implements the filesystem notification watch
- `wiki.go` implements the main function

### Changing the markup rules

If you want to change the markup rules, your starting point should be
`parser.go`. Make sure you read the documentation of [Go
Markdown](https://github.com/gomarkdown/markdown) and note that it
offers MathJax support (needs a change to the `view.html` template so
that the MathJax Javascript gets loaded) and
[MMark](https://mmark.miek.nl/post/syntax/) support, and it shows how
extensions can be added.

### Filenames and URL path

One of the sad parts of the code is the distinction between path and
filepath. On a Linux system, this doesn't matter. I suspect that it
also doesn't matter on MacOS and Windows because the file systems
handle forward slashes just fine. The code still tries to do the right
thing. A path that is derived from a URL is a path with slashes.
Before accessing a file, it has to be turned into a filepath using
`filepath.FromSlashes` and in the rare case where the inverse happens,
use `filepath.ToSlashes`. Any path received via the URL path uses
slashes and needs to be converted to a filepath before passing it to
any `os` function. Any path received within a `path/filepath.WalkFunc`
is a filepath and needs to be converted to use slashes when used in
HTML output.

If you need to access the page name in code that is used from a
template, you have to decode the path. See the code in `diff.go` for
an example.

### HTTP handlers

The URL paths all have the form `/action/directory/pagename` (with
directory being optional and pagename sometimes being optional). If
you need to limit access in Apache or nginx or some other web server
acting as a [reverse
proxy](https://en.wikipedia.org/wiki/Reverse_proxy), you can do that.
See `man oddmu-apache` and `man oddmu-nginx` for some configuration
examples.

This is how you can prevent some actions by simply not passing them on
to Oddmu, or you can require authentication for certain actions.
Furthermore, you can do the same for directories, allowing you to use
subdirectories as separate sites, each with their own editors.

## Dependencies

This section lists the non-standard libraries Oddmu uses.

[github.com/gomarkdown/markdown](https://github.com/gomarkdown/markdown)
is used to generate the web pages from Markdown.

[github.com/microcosm-cc/bluemonday](https://github.com/microcosm-cc/bluemonday)
is used to strip rendered search results of all HTML except for the
bold tag. Regular HTML generated from pages is *not* sanitized. Don't
give people you don't trust access to your wiki.

[github.com/pemistahl/lingua-go](https://github.com/pemistahl/lingua-go)
detects languages in order to set the language tag in templates. This
in turn can be used by browsers to get hyphenation right.

[github.com/gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype)
is used to sniff the MIME type of files with unknown filename
extensions.

[github.com/bashdrew/goheif](https://github.com/bashdrew/goheif) is
used to decode HEIC files (the new default file format for photos on
iPhones).

[github.com/disintegration/imaging](https://github.com/disintegration/imaging)
is used to resize images.

[github.com/edwvee/exiffix](https://github.com/edwvee/exiffix) is used
to rotate images before resizing them if the EXIF data says the image
wasn't taken with the default orientation of the camera. This is
necessary because after resizing, the EXIF data is gone.

[github.com/google/subcommands](https://github.com/google/subcommands)
is used for the parsing and documenting of subcommands.

[github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
is used to add colours to the search subcommand output.

[github.com/hexops/gotextdiff](https://github.com/hexops/gotextdiff)
is used to show a compact unified diff on the command line before
doing any replacements.

[github.com/sergi/go-diff/diffmatchpatch](https://github.com/sergi/go-diff/diffmatchpatch)
is used to show the page diffs on the web.

[github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify)
is used to watch the filesystem for changes.

[golang.org/x/exp/constraints](https://golang.org/x/exp/constraints)
for the computation of the intersection between two sets of pages.

[github.com/stretchr/testify/assert](https://github.com/stretchr/testify/assert)
is used for testing.

## Bugs

If you spot any, [contact](https://alexschroeder.ch/wiki/Contact) me.

## References

[Writing Web Applications](https://golang.org/doc/articles/wiki/)
provided the initial code for this wiki.
