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

This wiki uses a [Markdown
library](https://github.com/gomarkdown/markdown) to generate the web
pages from Markdown. There are two extensions Oddmu adds to the
library: local links `[[like this]]` and hashtags `#Like_This`.

This wiki uses the [lingua](https://github.com/pemistahl/lingua-go)
library to detect languages in order to get hyphenation right.

This wiki uses the standard
[html/template](https://pkg.go.dev/html/template) library to generate
HTML.

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

[oddmu-search(1)](/oddmu.git/blob/main/man/oddmu-search.1.txt): This
man page documents the "search" subcommand which you can use to build
indexes – lists of page links. These are important for feeds.

[oddmu-replace(1)](/oddmu.git/blob/main/man/oddmu-replace.1.txt): This
man page documents the "replace" subcommand to make mass changes to
the files much like find(1), grep(1) and sed(1) or perl (1).

[oddmu-html(1)](/oddmu.git/blob/main/man/oddmu-html.1.txt): This man
page documents the "html" subcommand to generate HTML from Markdown
pages from the command line.

[oddmu-templates(5)](/oddmu.git/blob/main/man/oddmu-templates.5.txt):
This man page documents how the templates can be changed (how they
*must* be changed) and lists the attributes available for the various
templates.

[oddmu-apache(5)](/oddmu.git/blob/main/man/oddmu-apache.5.txt): This
man page documents how to set up the web server for various common
tasks such as using logins to limit what visitors can edit.

[oddmu.service(5)](/oddmu.git/blob/main/man/oddmu.service.5.txt): This
man page documents how to setup a systemd unit and have it manage
Oddmu. “Great configurability brings brings great burdens.”

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

## References

[Writing Web Applications](https://golang.org/doc/articles/wiki/)
provided the initial code for this wiki.
