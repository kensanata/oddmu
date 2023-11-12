package main

import (
	"bytes"
	"github.com/microcosm-cc/bluemonday"
	"html/template"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Page is a struct containing information about a single page. Title
// is the title extracted from the page content using titleRegexp.
// Name is the filename without extension (so a filename of "foo.md"
// results in the Name "foo"). Body is the Markdown content of the
// page and Html is the rendered HTML for that Markdown. Score is a
// number indicating how well the page matched for a search query.
type Page struct {
	Title    string
	Name     string
	Language string
	Body     []byte
	Html     template.HTML
	Score    int
	Hashtags []string
}

var blogRe = regexp.MustCompile(`^\d\d\d\d-\d\d-\d\d`)

// santizeStrict uses bluemonday to sanitize the HTML away. No elements are allowed except for the b tag because this is
// used for snippets.
func sanitizeStrict(s string) template.HTML {
	policy := bluemonday.StrictPolicy()
	policy.AllowElements("b")
	return template.HTML(policy.Sanitize(s))
}

// santizeBytes uses bluemonday to sanitize the HTML used for pages. This is where you make changes if you want to be
// more lenient.
func sanitizeBytes(bytes []byte) template.HTML {
	policy := bluemonday.UGCPolicy()
	policy.AllowURLSchemes("gemini", "gopher")
	policy.AllowAttrs("title", "class", "style").Globally()
	policy.AllowAttrs("loading").OnElements("img") // for lazy loading
	// SVG, based on https://svgwg.org/svg2-draft/attindex.html transformed using
	// (while (zerop (forward-line 1))
	//   (when (looking-at "\\([^\t]+\\)\t\\([^\t]+\\).*")
	//     (let ((attribute (match-string 1))
	//           (elements (split-string (match-string 2) ", ")))
	//       (delete-region (point) (line-end-position))
	//       (insert "policy.AllowAttrs(\"" attribute "\").OnElements("
	//               (mapconcat (lambda (elem) (concat "\"" elem "\"")) elements ", ")
	//               ")"))))
	// Manually delete "script", "crossorigin", all attributes starting with "on", "ping"
	// and add elements without attributes allowed
	// (while (re-search-forward "\tpolicy.AllowAttrs(\\(.*\\)).OnElements(\\(.*\\))\n\tpolicy.AllowAttrs(\\1).OnElements(\\(.*\\))" nil t)
	//   (replace-match "\tpolicy.AllowAttrs(\\1).OnElements(\\2, \\3)"))
	// (while (re-search-forward "\tpolicy.AllowAttrs(\\(.*\\)).OnElements(\\(.*\\))\n\tpolicy.AllowAttrs(\\(.*\\)).OnElements(\\2)" nil t)
	//   (replace-match "\tpolicy.AllowAttrs(\\1, \\2).OnElements(\\3)"))
	policy.AllowNoAttrs().OnElements("defs")
	policy.AllowAttrs("alignment-baseline", "baseline-shift", "clip-path", "clip-rule", "color", "color-interpolation", "color-interpolation-filters", "cursor", "direction", "display", "dominant-baseline", "fill-opacity", "fill-rule", "filter", "flood-color", "flood-opacity", "font-family", "font-size", "font-size-adjust", "font-stretch", "font-style", "font-variant", "font-weight", "glyph-orientation-horizontal", "glyph-orientation-vertical", "image-rendering", "letter-spacing", "lighting-color", "marker-end", "marker-mid", "marker-start", "mask", "mask-type", "opacity", "overflow", "paint-order", "pointer-events", "shape-rendering", "stop-color", "stop-opacity", "stroke", "stroke-dasharray", "stroke-dashoffset", "stroke-linecap", "stroke-linejoin", "stroke-miterlimit", "stroke-opacity", "stroke-width", "text-anchor", "text-decoration", "text-overflow", "text-rendering", "transform-origin", "unicode-bidi", "vector-effect", "visibility", "white-space", "word-spacing", "writing-mode").Globally() // SVG elements
	policy.AllowAttrs("accumulate", "additive", "by", "calcMode", "from", "keySplines", "keyTimes", "values").OnElements("animate", "animateMotion", "animateTransform")
	policy.AllowAttrs("amplitude").OnElements("feFuncA", "feFuncB", "feFuncG", "feFuncR")
	policy.AllowAttrs("aria-activedescendant", "aria-atomic", "aria-autocomplete", "aria-busy", "aria-checked", "aria-colcount", "aria-colindex", "aria-colspan", "aria-controls", "aria-current", "aria-describedby", "aria-details", "aria-disabled", "aria-dropeffect", "aria-errormessage", "aria-expanded", "aria-flowto", "aria-grabbed", "aria-haspopup", "aria-hidden", "aria-invalid", "aria-keyshortcuts", "aria-label", "aria-labelledby", "aria-level", "aria-live", "aria-modal", "aria-multiline", "aria-multiselectable", "aria-orientation", "aria-owns", "aria-placeholder", "aria-posinset", "aria-pressed", "aria-readonly", "aria-relevant", "aria-required", "aria-roledescription", "aria-rowcount", "aria-rowindex", "aria-rowspan", "aria-selected", "aria-setsize", "aria-sort", "aria-valuemax", "aria-valuemin", "aria-valuenow", "aria-valuetext", "role").OnElements("a", "circle", "discard", "ellipse", "foreignObject", "g", "image", "line", "path", "polygon", "polyline", "rect", "svg", "switch", "symbol", "text", "textPath", "tspan", "use", "view")
	policy.AllowAttrs("attributeName").OnElements("animate", "animateTransform", "set")
	policy.AllowAttrs("autofocus").OnElements("a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "defs", "desc", "discard", "ellipse", "feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDistantLight", "feDropShadow", "feFlood", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur", "feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset", "fePointLight", "feSpecularLighting", "feSpotLight", "feTile", "feTurbulence", "filter", "foreignObject", "g", "image", "line", "linearGradient", "marker", "mask", "metadata", "mpath", "path", "pattern", "polygon", "polyline", "radialGradient", "rect", "set", "stop", "style", "svg", "switch", "symbol", "text", "textPath", "title", "tspan", "use", "view")
	policy.AllowAttrs("azimuth", "elevation").OnElements("feDistantLight")
	policy.AllowAttrs("baseFrequency", "numOctaves", "seed", "stitchTiles").OnElements("feTurbulence")
	policy.AllowAttrs("begin").OnElements("animate", "animateMotion", "animateTransform", "set", "discard")
	policy.AllowAttrs("bias", "divisor", "kernelMatrix", "order", "preserveAlpha", "targetX", "targetY").OnElements("feConvolveMatrix")
	policy.AllowAttrs("class").OnElements("a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "defs", "desc", "discard", "ellipse", "feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDistantLight", "feDropShadow", "feFlood", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur", "feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset", "fePointLight", "feSpecularLighting", "feSpotLight", "feTile", "feTurbulence", "filter", "foreignObject", "g", "image", "line", "linearGradient", "marker", "mask", "metadata", "mpath", "path", "pattern", "polygon", "polyline", "radialGradient", "rect", "set", "stop", "style", "svg", "switch", "symbol", "text", "textPath", "title", "tspan", "use", "view")
	policy.AllowAttrs("clipPathUnits").OnElements("clipPath")
	policy.AllowAttrs("cx", "cy").OnElements("circle", "ellipse", "radialGradient")
	policy.AllowAttrs("d").OnElements("path")
	policy.AllowAttrs("diffuseConstant").OnElements("feDiffuseLighting")
	policy.AllowAttrs("download").OnElements("a")
	policy.AllowAttrs("dur").OnElements("animate", "animateMotion", "animateTransform", "set")
	policy.AllowAttrs("dx", "dy").OnElements("feDropShadow", "feOffset", "text", "tspan")
	policy.AllowAttrs("edgeMode").OnElements("feConvolveMatrix", "feGaussianBlur")
	policy.AllowAttrs("end").OnElements("animate", "animateMotion", "animateTransform", "set")
	policy.AllowAttrs("exponent").OnElements("feFuncA", "feFuncB", "feFuncG", "feFuncR")
	policy.AllowAttrs("fill").Globally() // at least for all SVG elements
	policy.AllowAttrs("filterUnits").OnElements("filter")
	policy.AllowAttrs("fr", "fx", "fy").OnElements("radialGradient")
	policy.AllowAttrs("gradientTransform", "gradientUnits").OnElements("linearGradient", "radialGradient")
	policy.AllowAttrs("height").OnElements("feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDropShadow", "feFlood", "feGaussianBlur", "feImage", "feMerge", "feMorphology", "feOffset", "feSpecularLighting", "feTile", "feTurbulence", "filter", "mask", "pattern", "foreignObject", "image", "rect", "svg", "symbol", "use")
	policy.AllowAttrs("href").OnElements("a", "animate", "animateMotion", "animateTransform", "set", "discard", "feImage", "image", "linearGradient", "mpath", "pattern", "radialGradient", "textPath", "use")
	policy.AllowAttrs("hreflang").OnElements("a")
	policy.AllowAttrs("id").OnElements("a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "defs", "desc", "discard", "ellipse", "feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDistantLight", "feDropShadow", "feFlood", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur", "feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset", "fePointLight", "feSpecularLighting", "feSpotLight", "feTile", "feTurbulence", "filter", "foreignObject", "g", "image", "line", "linearGradient", "marker", "mask", "metadata", "mpath", "path", "pattern", "polygon", "polyline", "radialGradient", "rect", "set", "stop", "style", "svg", "switch", "symbol", "text", "textPath", "title", "tspan", "use", "view")
	policy.AllowAttrs("in").OnElements("feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDropShadow", "feGaussianBlur", "feMergeNode", "feMorphology", "feOffset", "feSpecularLighting", "feTile")
	policy.AllowAttrs("in2").OnElements("feBlend", "feComposite", "feDisplacementMap")
	policy.AllowAttrs("intercept").OnElements("feFuncA", "feFuncB", "feFuncG", "feFuncR")
	policy.AllowAttrs("k1", "k2", "k3", "k4").OnElements("feComposite")
	policy.AllowAttrs("kernelUnitLength").OnElements("feConvolveMatrix", "feDiffuseLighting", "feSpecularLighting")
	policy.AllowAttrs("keyPoints").OnElements("animateMotion")
	policy.AllowAttrs("lang").OnElements("a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "defs", "desc", "discard", "ellipse", "feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDistantLight", "feDropShadow", "feFlood", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur", "feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset", "fePointLight", "feSpecularLighting", "feSpotLight", "feTile", "feTurbulence", "filter", "foreignObject", "g", "image", "line", "linearGradient", "marker", "mask", "metadata", "mpath", "path", "pattern", "polygon", "polyline", "radialGradient", "rect", "set", "stop", "style", "svg", "switch", "symbol", "text", "textPath", "title", "tspan", "use", "view")
	policy.AllowAttrs("lengthAdjust").OnElements("text", "textPath", "tspan")
	policy.AllowAttrs("limitingConeAngle").OnElements("feSpotLight")
	policy.AllowAttrs("markerHeight", "markerUnits", "markerWidth").OnElements("marker")
	policy.AllowAttrs("maskContentUnits", "mask").OnElements("maskUnits")
	policy.AllowAttrs("max").OnElements("animate", "animateMotion", "animateTransform", "set")
	policy.AllowAttrs("media").OnElements("style")
	policy.AllowAttrs("method").OnElements("textPath")
	policy.AllowAttrs("min").OnElements("animate", "animateMotion", "animateTransform", "set")
	policy.AllowAttrs("mode").OnElements("feBlend")
	policy.AllowAttrs("offset").OnElements("feFuncA", "feFuncB", "feFuncG", "feFuncR", "stop")
	policy.AllowAttrs("operator").OnElements("feComposite", "feMorphology")
	policy.AllowAttrs("orient").OnElements("marker")
	policy.AllowAttrs("origin").OnElements("animateMotion")
	policy.AllowAttrs("path").OnElements("animateMotion", "textPath")
	policy.AllowAttrs("pathLength").OnElements("circle", "ellipse", "line", "path", "polygon", "polyline", "rect")
	policy.AllowAttrs("patternContentUnits", "pattern").OnElements("patternTransform")
	policy.AllowAttrs("patternUnits").OnElements("pattern")
	policy.AllowAttrs("playbackorder", "timelinebegin", "transform").OnElements("svg")
	policy.AllowAttrs("points").OnElements("polygon", "polyline")
	policy.AllowAttrs("pointsAtX", "feSpotLight").OnElements("pointsAtY")
	policy.AllowAttrs("pointsAtZ").OnElements("feSpotLight")
	policy.AllowAttrs("preserveAspectRatio").OnElements("feImage", "image", "marker", "pattern", "svg", "symbol", "view")
	policy.AllowAttrs("primitiveUnits").OnElements("filter")
	policy.AllowAttrs("r").OnElements("circle", "radialGradient")
	policy.AllowAttrs("rx", "ry").OnElements("ellipse", "rect")
	policy.AllowAttrs("radius").OnElements("feMorphology")
	policy.AllowAttrs("refX", "marker", "symbol").OnElements("refY")
	policy.AllowAttrs("referrerpolicy", "a").OnElements("rel")
	policy.AllowAttrs("repeatCount", "animate", "animateMotion", "animateTransform", "set").OnElements("repeatDur")
	policy.AllowAttrs("requiredExtensions").OnElements("a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "discard", "ellipse", "foreignObject", "g", "image", "line", "mask", "path", "polygon", "polyline", "rect", "set", "svg", "switch", "text", "textPath", "tspan", "use")
	policy.AllowAttrs("restart").OnElements("animate", "animateMotion", "animateTransform", "set")
	policy.AllowAttrs("result").OnElements("feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDropShadow", "feFlood", "feGaussianBlur", "feImage", "feMerge", "feMorphology", "feOffset", "feSpecularLighting", "feTile", "feTurbulence")
	policy.AllowAttrs("rotate").OnElements("animateMotion", "text", "tspan")
	policy.AllowAttrs("scale").OnElements("feDisplacementMap")
	policy.AllowAttrs("side").OnElements("textPath")
	policy.AllowAttrs("slope").OnElements("feFuncA", "feFuncB", "feFuncG", "feFuncR")
	policy.AllowAttrs("spacing").OnElements("textPath")
	policy.AllowAttrs("specularConstant").OnElements("feSpecularLighting")
	policy.AllowAttrs("specularExponent").OnElements("feSpecularLighting", "feSpotLight")
	policy.AllowAttrs("spreadMethod").OnElements("linearGradient", "radialGradient")
	policy.AllowAttrs("startOffset").OnElements("textPath")
	policy.AllowAttrs("stdDeviation").OnElements("feDropShadow", "feGaussianBlur")
	policy.AllowAttrs("style").OnElements("a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "defs", "desc", "discard", "ellipse", "feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDistantLight", "feDropShadow", "feFlood", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur", "feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset", "fePointLight", "feSpecularLighting", "feSpotLight", "feTile", "feTurbulence", "filter", "foreignObject", "g", "image", "line", "linearGradient", "marker", "mask", "metadata", "mpath", "path", "pattern", "polygon", "polyline", "radialGradient", "rect", "set", "stop", "style", "svg", "switch", "symbol", "text", "textPath", "title", "tspan", "use", "view")
	policy.AllowAttrs("surfaceScale").OnElements("feDiffuseLighting", "feSpecularLighting")
	policy.AllowAttrs("systemLanguage").OnElements("a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "discard", "ellipse", "foreignObject", "g", "image", "line", "mask", "path", "polygon", "polyline", "rect", "set", "svg", "switch", "text", "textPath", "tspan", "use")
	policy.AllowAttrs("tabindex").OnElements("a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "defs", "desc", "discard", "ellipse", "feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDistantLight", "feDropShadow", "feFlood", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur", "feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset", "fePointLight", "feSpecularLighting", "feSpotLight", "feTile", "feTurbulence", "filter", "foreignObject", "g", "image", "line", "linearGradient", "marker", "mask", "metadata", "mpath", "path", "pattern", "polygon", "polyline", "radialGradient", "rect", "set", "stop", "style", "svg", "switch", "symbol", "text", "textPath", "title", "tspan", "use", "view")
	policy.AllowAttrs("tableValues").OnElements("feFuncA", "feFuncB", "feFuncG", "feFuncR")
	policy.AllowAttrs("target").OnElements("a")
	policy.AllowAttrs("textLength").OnElements("text", "textPath", "tspan")
	policy.AllowAttrs("title").OnElements("style")
	policy.AllowAttrs("to").OnElements("animate", "animateMotion", "animateTransform", "set")
	policy.AllowAttrs("transform").Globally() // for almost all SVG elements (with the exception of the pattern, linearGradient and radialGradient elements)
	policy.AllowAttrs("type").OnElements("a", "animateTransform", "feColorMatrix", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feTurbulence", "style")
	policy.AllowAttrs("values").OnElements("feColorMatrix")
	policy.AllowAttrs("viewBox").OnElements("marker", "pattern", "svg", "symbol", "view")
	policy.AllowAttrs("width").OnElements("feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDropShadow", "feFlood", "feGaussianBlur", "feImage", "feMerge", "feMorphology", "feOffset", "feSpecularLighting", "feTile", "feTurbulence", "filter", "mask", "pattern", "foreignObject", "image", "rect", "svg", "symbol", "use")
	policy.AllowAttrs("x").OnElements("feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDropShadow", "feFlood", "feGaussianBlur", "feImage", "feMerge", "feMorphology", "feOffset", "feSpecularLighting", "feTile", "feTurbulence", "fePointLight", "feSpotLight", "filter", "mask", "pattern", "text", "tspan", "foreignObject", "image", "rect", "svg", "symbol", "use")
	policy.AllowAttrs("x1", "x2", "y1", "y2").OnElements("line", "linearGradient")
	policy.AllowAttrs("xChannelSelector").OnElements("feDisplacementMap")
	policy.AllowAttrs("xlink:href").OnElements("a", "image", "linearGradient", "pattern", "radialGradient", "textPath", "use", "feImage")
	policy.AllowAttrs("xlink:title").OnElements("a", "image", "linearGradient", "pattern", "radialGradient", "textPath", "use")
	policy.AllowAttrs("xml:space").OnElements("a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "defs", "desc", "discard", "ellipse", "feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDistantLight", "feDropShadow", "feFlood", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur", "feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset", "fePointLight", "feSpecularLighting", "feSpotLight", "feTile", "feTurbulence", "filter", "foreignObject", "g", "image", "line", "linearGradient", "marker", "mask", "metadata", "mpath", "path", "pattern", "polygon", "polyline", "radialGradient", "rect", "set", "stop", "style", "svg", "switch", "symbol", "text", "textPath", "title", "tspan", "use", "view")
	policy.AllowAttrs("y").OnElements("feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDropShadow", "feFlood", "feGaussianBlur", "feImage", "feMerge", "feMorphology", "feOffset", "feSpecularLighting", "feTile", "feTurbulence", "fePointLight", "feSpotLight", "filter", "mask", "pattern", "text", "tspan", "foreignObject", "image", "rect", "svg", "symbol", "use")
	policy.AllowAttrs("yChannelSelector").OnElements("feDisplacementMap")
	policy.AllowAttrs("z").OnElements("fePointLight", "feSpotLight")
	return template.HTML(policy.SanitizeBytes(bytes))
}

// nameEscape returns the page name safe for use in URLs. That is,
// percent escaping is used except for the slashes.
func nameEscape(s string) string {
	parts := strings.Split(s, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

// save saves a Page. The filename is based on the Page.Name and gets
// the ".md" extension. Page.Body is saved, without any carriage
// return characters ("\r"). Page.Title and Page.Html are not saved.
// There is no caching. Before removing or writing a file, the old
// copy is renamed to a backup, appending "~". There is no error
// checking for this.
func (p *Page) save() error {
	filename := p.Name + ".md"
	s := bytes.ReplaceAll(p.Body, []byte{'\r'}, []byte{})
	if len(s) == 0 {
		p.removeFromIndex()
		return os.Rename(filename, filename+"~")
	}
	p.Body = s
	p.updateIndex()
	d := filepath.Dir(filename)
	if d != "." {
		err := os.MkdirAll(d, 0755)
		if err != nil {
			log.Printf("Creating directory %s failed: %s", d, err)
			return err
		}
	}
	_ = os.Rename(filename, filename+"~")
	return os.WriteFile(filename, s, 0644)
}

// loadPage loads a Page given a name. The filename loaded is that
// Page.Name with the ".md" extension. The Page.Title is set to the
// Page.Name (and possibly changed, later). The Page.Body is set to
// the file content. The Page.Html remains undefined (there is no
// caching).
func loadPage(name string) (*Page, error) {
	filename := name + ".md"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: name, Name: name, Body: body, Language: ""}, nil
}

// handleTitle extracts the title from a Page and sets Page.Title, if any. If replace is true, the page title is also
// removed from Page.Body. Make sure not to save this! This is only for rendering. In a template, the title is a
// separate attribute and is not repeated in the HTML.
func (p *Page) handleTitle(replace bool) {
	s := string(p.Body)
	m := titleRegexp.FindStringSubmatch(s)
	if m != nil {
		p.Title = m[1]
		if replace {
			p.Body = []byte(strings.Replace(s, m[0], "", 1))
		}
	}
}

// score sets Page.Title and computes Page.Score.
func (p *Page) score(q string) {
	p.handleTitle(true)
	p.Score = score(q, string(p.Body)) + score(q, p.Title)
}

// summarize sets Page.Html to an extract and sets Page.Language.
func (p *Page) summarize(q string) {
	t := p.plainText()
	p.Name = nameEscape(p.Name)
	p.Html = sanitizeStrict(snippets(q, t))
	p.Language = language(t)
}

// isBlog returns true if the page name starts with an ISO date
func (p *Page) isBlog() bool {
	name := path.Base(p.Name)
	return blogRe.MatchString(name)
}

// Dir returns the directory the page is in. It's either the empty string if the page is in the Oddmu working directory,
// or it ends in a slash. This is used to create the upload link in "view.html", for example.
func (p *Page) Dir() string {
	d := filepath.Dir(p.Name)
	if d == "." {
		return ""
	}
	return d + "/"
}

// Today returns the date, as a string, for use in templates.
func (p *Page) Today() string {
	return time.Now().Format(time.DateOnly)
}
