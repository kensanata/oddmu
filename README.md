# Oddμ: A minimal wiki

This program helps you run a minimal wiki, blog, digital garden, memex
or Zettelkasten. There is no version history.

It's well suited as a self-hosted, single-user web application, when
there is no need for collaboration on the site itself. Links and email
connect you to the rest of the net. The wiki can be public or private.
Perhaps it just runs on your local machine, unreachable from the
Internet.

It's well suited as a secondary medium for a close-knit group:
collaboration and conversation happens elsewhere, in chat, on social
media. The wiki serves as the text repository that results from these
discussions. As there are no logins and no version histories, it is
not possible to undo vandalism and spam. Only allow people you trust
write-access to the site.

It's well suited as a simple static site generator. There are no
plugins.

When Oddμ runs as a web server, it serves all the Markdown files
(ending in `.md`) as web pages. These pages can be edited via the web.

Oddmu adds the following extensions to Markdown: local links `[[like
this]]`, hashtags `#Like_This` and fediverse account links like
`@alex@alexschroeder.ch`.

If your pages don't provide their own title (`# title`), the file name
(without `.md`) is used as the title. Subdirectories are created as
necessary.

Other files can be uploaded and images (ending in `.jpg`, `.jpeg`,
`.png`, `.heic` or `.webp`) can be resized when they are uploaded
(resulting in `.jpg`, `.png` or `.webp` files).

## Documentation

This project uses man(1) pages. They are generated from text files
using [scdoc](https://git.sr.ht/~sircmpwn/scdoc). These are the files
available:

[oddmu(1)](https://alexschroeder.ch/view/oddmu/oddmu.1): This man page
has a short introduction to Oddmu, its configuration via templates and
environment variables, plus points to the other man pages.

[oddmu(5)](https://alexschroeder.ch/view/oddmu/oddmu.5): This man page
talks about the Markdown and includes some examples for the
non-standard features such as table markup. It also talks about the
Oddmu extensions to Markdown: wiki links, hashtags and fediverse
account links. Local links must use percent encoding for page names so
there is a section about percent encoding. The man page also explains
how feeds are generated.

[oddmu-releases(7)](https://alexschroeder.ch/view/oddmu/oddmu-releases.7):
This man page lists all the Oddmu versions and their user-visible
changes.

[oddmu-version(1)](https://alexschroeder.ch/view/oddmu/oddmu-version.1):
This man page documents the "version" subcommand which you can use to
get the installed Oddmu version.

Working locally:

[oddmu-links(1)](https://alexschroeder.ch/view/oddmu/oddmu-links.1):
This man page documents the "links" subcommand which you can use to
get the outgoing links for a page.

[oddmu-list(1)](https://alexschroeder.ch/view/oddmu/oddmu-list.1):
This man page documents the "list" subcommand which you can use to get
page names and page titles.

[oddmu-replace(1)](https://alexschroeder.ch/view/oddmu/oddmu-replace.1):
This man page documents the "replace" subcommand to make mass changes
to the files much like find(1), grep(1) and sed(1) or perl(1).

[oddmu-search(1)](https://alexschroeder.ch/view/oddmu/oddmu-search.1):
This man page documents the "search" subcommand which you can use to
build indexes – lists of page links. These are important for feeds.

[oddmu-search(7)](https://alexschroeder.ch/view/oddmu/oddmu-search.7):
This man page documents how search and scoring work.

[oddmu-toc(1)](https://alexschroeder.ch/view/oddmu/oddmu-toc.1): This
man page documents the "toc" subcommand which you can use to generate
a table of contents linking to all the headings on the page.

Reporting:

[oddmu-missing(1)](https://alexschroeder.ch/view/oddmu/oddmu-missing.1):
This man page documents the "missing" subcommand to list local links
that don't point to any existing pages or files.

[oddmu-hashtags(1)](https://alexschroeder.ch/view/oddmu/oddmu-hashtags.1):
This man page documents the "hashtags" subcommand to count the
hashtags used from the command line.

Static site generator:

[oddmu-html(1)](https://alexschroeder.ch/view/oddmu/oddmu-html.1):
This man page documents the "html" subcommand to generate HTML from
Markdown pages from the command line.

[oddmu-static(1)](https://alexschroeder.ch/view/oddmu/oddmu-static.1):
This man page documents the "static" subcommand to generate an entire
static website from the command line, avoiding the need to run Oddmu
as a server. Also great for archiving.

[oddmu-notify(1)](https://alexschroeder.ch/view/oddmu/oddmu-notify.1):
This man page documents the "notify" subcommand to add links to
hashtag pages, index and changes for a given page. This is useful when
you edit the Markdown files locally.

Configuration:

[oddmu-templates(5)](https://alexschroeder.ch/view/oddmu/oddmu-templates.5):
This man page documents how the templates can be changed (how they
*must* be changed) and lists the attributes available for the various
templates.

System administration:

[oddmu-apache(5)](https://alexschroeder.ch/view/oddmu/oddmu-apache.5):
This man page documents how to set up the Apache web server for
various common tasks such as using logins to limit what visitors can
edit.

[oddmu-filter(7)](https://alexschroeder.ch/view/oddmu/oddmu-filter.7):
This man page documents how to exclude subdirectories from search and
archiving.

[oddmu-nginx(5)](https://alexschroeder.ch/view/oddmu/oddmu-nginx.5):
This man page documents how to set up the freenginx web server for
various common tasks such as using logins to limit what visitors can
edit.

[oddmu.service(5)](https://alexschroeder.ch/view/oddmu/oddmu.service.5):
This man page documents how to setup a systemd unit and have it manage
Oddmu. “Great configurability brings great burdens.”

[oddmu-webdav(5)](https://alexschroeder.ch/view/oddmu/oddmu-webdav.5):
This man page documents how to set up the Apache web server so that
the wiki can be accessed via Web-DAV.

Leaving:

[oddmu-export(1)](https://alexschroeder.ch/view/oddmu/oddmu-export.1):
This man page documents how to export all the pages as one RSS feed so
that you can import them all into a new platform that doesn't use
Markdown files.

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

Here's an example using [GNU Stow](https://www.gnu.org/software/stow/)
to install it into `/usr/local/stow` in a way that allows you to
uninstall it later:

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
- `list.go` implements the file list page
- `page.go` implements the page loading and saving
- `parser.go` implements the Markdown parsing
- `preview.go` implements the `/preview` handler
- `score.go` implements the page scoring when showing search results
- `search.go` implements the `/search` handler
- `snippets.go` implements the page summaries for search results
- `templates.go` implements template loading and reloading
- `tokenizer.go` implements the various tokenizers used
- `upload_drop.go` implements the `/upload` and `/drop` handlers
- `view.go` implements the `/view` handler
- `watch.go` implements the filesystem notification watch
- `wiki.go` implements the main function

The code of this package is licensed to you under the
AGPL-3.0-or-later license. If you do make changes and your site is
public, be aware of section 13:

> … if you modify the Program, your modified version must prominently
> offer all users interacting with it remotely through a computer
> network (if your version supports such interaction) an opportunity
> to receive the Corresponding Source of your version by providing
> access to the Corresponding Source from a network server at no
> charge, through some standard or customary means of facilitating
> copying of software.

### Changing the markup rules

If you want to change the markup rules, your starting point should be
`parser.go`. Make sure you read the documentation of [Go
Markdown](https://github.com/gomarkdown/markdown) and note that it
offers MathJax support (needs a change to the `view.html` template so
that the MathJax Javascript gets loaded) and
[MMark](https://mmark.miek.nl/post/syntax/) support, and it shows how
extensions can be added.

### Filenames and URL path

There are some simplifications made. The code doesn't consider the
various encodings (UTF-8 NFC on the web vs UTF-8 NFD for HFS+, for
example; it also doesn't check for characters in page names that are
illegal filenames on the filesystem used).

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

### Templates

The `themes` folder has some ideas of how to tweak the HTML templates.

### Permissions

An unexplored idea would be to parse a config file that has usernames
and passwords, groups usernames into roles, and assigns access to the
various actions based on these roles. This would obviate the need for
a web server acting as a reverse proxy.

Then again, not having to care about roles and permissions has been a
relief.

## Dependencies

This section lists the non-standard libraries Oddmu uses and their
respective licenses.

[github.com/gomarkdown/markdown](https://github.com/gomarkdown/markdown)
is used to generate the web pages from Markdown. BSD-2-Clause.

[github.com/microcosm-cc/bluemonday](https://github.com/microcosm-cc/bluemonday)
is used to strip rendered search results of all HTML except for the
bold tag. Regular HTML generated from pages is *not* sanitized. Don't
give people you don't trust access to your wiki. BSD-3-Clause.

[github.com/pemistahl/lingua-go](https://github.com/pemistahl/lingua-go)
detects languages in order to set the language tag in templates. This
in turn can be used by browsers to get hyphenation right. Apache-2.0.

[github.com/gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype)
is used to sniff the MIME type of files with unknown filename
extensions. MIT.

[github.com/gen2brain/heic](https://github.com/gen2brain/heic) is used
to decode HEIC files (the new default file format for photos on
iPhones). MIT.

[github.com/gen2brain/webp](https://github.com/gen2brain/webp) is used
to encode and decode WebP files. MIT.

[github.com/disintegration/imaging](https://github.com/disintegration/imaging)
is used to resize images. MIT.

[github.com/edwvee/exiffix](https://github.com/edwvee/exiffix) is used
to rotate images before resizing them if the EXIF data says the image
wasn't taken with the default orientation of the camera. This is
necessary because after resizing, the EXIF data is gone. MIT.

[github.com/google/subcommands](https://github.com/google/subcommands)
is used for the parsing and documenting of subcommands. Apache-2.0.

[github.com/muesli/reflow/wordwrap](https://github.com/muesli/reflow/wordwrap)
is used to wrap the search subcommand output. MIT.

[github.com/hexops/gotextdiff](https://github.com/hexops/gotextdiff)
is used to show a compact unified diff on the command line before
doing any replacements. BSD-3-Clause.

[github.com/sergi/go-diff/diffmatchpatch](https://github.com/sergi/go-diff/diffmatchpatch)
is used to show the page diffs on the web. MIT.

[github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify)
is used to watch the filesystem for changes. BSD-3-Clause.

[golang.org/x/exp/constraints](https://golang.org/x/exp/constraints)
for the computation of the intersection between two sets of pages.
BSD-3-Clause.

[github.com/stretchr/testify/assert](https://github.com/stretchr/testify/assert)
is used for testing. MIT.

## Bugs

If you spot any, [contact](https://alexschroeder.ch/wiki/Contact) me.

## References

[Writing Web Applications](https://golang.org/doc/articles/wiki/)
provided the initial code for this wiki.
