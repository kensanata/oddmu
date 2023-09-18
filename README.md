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
the fediverse, but humans.

This wiki uses a [Markdown
library](https://github.com/gomarkdown/markdown) to generate the web
pages from Markdown. There are two extensions Oddmu adds to the
library: local links `[[like this]]` and hashtags `#Like_This`.

This wiki uses the [lingua](github.com/pemistahl/lingua-go) library to
detect languages in order to get hyphenation right.

This wiki uses the standard
[html/template](https://pkg.go.dev/html/template) library to generate
HTML.

When saving a page, the page name is take from the URL and the page
content is taken from the `body` form parameter. To illustrate, here's
how to edit a page using `curl`:

```sh
curl --form body="Did you bring a towel?" \
  http://localhost:8080/save/welcome
```

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
