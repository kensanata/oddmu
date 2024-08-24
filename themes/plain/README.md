Plain theme
===========

This makes it look as if the site consists mostly of editable plain
text. Accordingly, the user interface has been simplified and there
are no links to the preview, add, diff and upload actions and the
corresponding templates have been deleted. There is no special static
or feed template (mostly because the feed would depend on the list of
links that isn't rendered).

Now, the text is still saved in Markdown files and the Markdown is
still rendered to HTML – but the "view" template just prints the page
body inside a "pre" block and ignores the rendered HTML.

This being text files, there are also no links to follow. That is why
there's no link here back to themes. Sorry!

This also means that you can only edit new files by editing the URL in
the address bar of your browser since you can't link to them.
