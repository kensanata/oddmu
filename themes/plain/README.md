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

In addition to that, a piece of JavaScript goes through the text and
replaces rocket links with actual links. Rocket links are how Gemini
links to other pages: On a line by itself (no inline links!) write
"=>", a space, the URL, and optionally another space and the text to
use.

Unfortunately, this magic happens in JavaScript. 😭

=> ../index Themes
