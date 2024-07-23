package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"slices"
	"testing"
)

func TestSortNames(t *testing.T) {
	index.Lock()
	defer index.Unlock()
	for _, s := range []string{"Alex", "Berta", "Chris", "2015-06-14", "2023-09-26"} {
		index.titles[s] = s
	}
	terms := []string{"Z"}
	fn := sortNames(terms)
	assert.Equal(t, 1, fn("Berta", "Alex"), "B is after A")
	assert.Equal(t, -1, fn("Alex", "Berta"), "A is before B")
	assert.Equal(t, 0, fn("Berta", "Berta"), "B and B are equal")
	assert.Equal(t, -1, fn("2023-09-26", "Alex"), "numbers before letters")
	assert.Equal(t, 1, fn("Alex", "2023-09-26"), "numbers after letters")
	assert.Equal(t, -1, fn("2023-09-26", "2015-06-14"), "higher numbers before lower numbers")
	assert.Equal(t, 1, fn("2015-06-14", "2023-09-26"), "lower numbers after higher numbers")

	names := []string{"Berta", "Chris", "Alex"}
	slices.SortFunc(names, sortNames(terms))
	assert.True(t, slices.IsSorted(names), fmt.Sprintf("Sorted: %v", names))
}

func TestPrependMatches(t *testing.T) {
	index.Lock()
	for _, s := range []string{"Alex", "Berta", "Chris"} {
		index.titles[s] = s
	}
	index.Unlock()
	r := []string{"Berta", "Chris"}         // does not prepend
	u := []string{"Alex", "Berta", "Chris"} // does prepend
	v, _ := prependQueryPage(r, "", "Alex")
	assert.Equal(t, u, v, "prepend q")
	v, _ = prependQueryPage(r, "", "lex")
	assert.Equal(t, r, v, "exact matches only")
	v, _ = prependQueryPage(r, "", "#Alex")
	assert.Equal(t, u, v, "prepend hashtag")
	v, _ = prependQueryPage(r, "", "#Alex #Berta")
	assert.Equal(t, r, v, "do not prepend two hashtags")
	v, _ = prependQueryPage(r, "", "#alex")
	assert.Equal(t, r, v, "do not ignore case")
	v, _ = prependQueryPage(u, "", "Alex")
	assert.Equal(t, u, v, "do not prepend q twice")
	v, _ = prependQueryPage([]string{"Berta", "Alex", "Chris"}, "", "Alex")
	assert.Equal(t, u, v, "sort q to the front")
	v, _ = prependQueryPage([]string{"Berta", "Chris", "Alex"}, "", "Alex")
	assert.Equal(t, u, v, "sort q to the front")
}

func TestSearch(t *testing.T) {
	// working in the main directory
	index.reset()
	index.load()

	data := url.Values{}
	data.Set("q", "oddµ")

	body := assert.HTTPBody(makeHandler(searchHandler, false), "GET", "/search/", data)
	assert.Contains(t, body, "Welcome")
	assert.Contains(t, body, `<span class="score">5</span>`)

	body = assert.HTTPBody(makeHandler(searchHandler, false), "GET", "/search/testdata", data)
	assert.NotContains(t, body, "Welcome")
}

func TestSearchFilter(t *testing.T) {
	names := []string{"a", "public/b", "secret/c"}

	f := filterPath(names, "", "")
	assert.Equal(t, names, f)

	f = filterPath(names, "public/", "")
	assert.Equal(t, []string{"public/b"}, f)

	f = filterPath(names, "secret/", "")
	assert.Equal(t, []string{"secret/c"}, f)

	// critically, this no longer returns c
	f = filterPath(names, "", "^secret/")
	assert.Equal(t, []string{"a", "public/b"}, f)

	// unchanged
	f = filterPath(names, "public/", "^secret/")
	assert.Equal(t, []string{"public/b"}, f)

	// unchanged
	f = filterPath(names, "secret/", "^secret/")
	assert.Equal(t, []string{"secret/c"}, f)

}

func TestSearchFilterLong(t *testing.T) {
	cleanup(t, "testdata/filter")
	p := &Page{Name: "testdata/filter/one", Body: []byte(`# One

One day, I heard you say
Just one more day and I'd know
But that was last spring`)}
	p.save()
	p = &Page{Name: "testdata/filter/public/two", Body: []byte(`# Two
Oh, the two of us
Have often seen this forest
But this bird is new`)}
	p.save()
	p = &Page{Name: "testdata/filter/secret/three", Body: []byte(`# Three
Three years have gone by
And we're good, we live, we breathe
But we don't say it`)}
	p.save()

	// normal search works
	items, _ := search("spring", "testdata/", "", 1, false)
	assert.Equal(t, len(items), 1)
	assert.Equal(t, "One", items[0].Title)

	// not found because it's in /secret and we start at /
	items, _ = search("year", "testdata/", "^testdata/filter/secret/", 1, false)
	assert.Equal(t, 0, len(items))

	// only found two because the third one is in /secret and we start at /
	items, _ = search("but", "testdata/", "^testdata/filter/secret/", 1, false)
	assert.Equal(t, 2, len(items))
	assert.Equal(t, "One", items[0].Title)
	assert.Equal(t, "Two", items[1].Title)

	// starting in the public/ directory, we find only one page
	items, _ = search("but", "testdata/filter/public/", "^testdata/filter/secret/", 1, false)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, "Two", items[0].Title)

	// starting in the secret/ directory, we find only one page
	items, _ = search("but", "testdata/filter/secret/", "^testdata/filter/secret/", 1, false)
	assert.Equal(t, 1, len(items))
	assert.Contains(t, "Three", items[0].Title)
}

func TestSearchDir(t *testing.T) {
	cleanup(t, "testdata/dir")
	p := &Page{Name: "testdata/dir/dice", Body: []byte(`# Dice

A tiny drum roll
Dice rolling bouncing stopping
Where is lady luck?`)}
	p.save()

	data := url.Values{}
	data.Set("q", "luck")

	body := assert.HTTPBody(makeHandler(searchHandler, false), "GET", "/search/", data)
	assert.Contains(t, body, "luck")

	body = assert.HTTPBody(makeHandler(searchHandler, false), "GET", "/search/testdata", data)
	assert.Contains(t, body, "luck")

	body = assert.HTTPBody(makeHandler(searchHandler, false), "GET", "/search/testdata/dir", data)
	assert.Contains(t, body, "luck")

	body = assert.HTTPBody(makeHandler(searchHandler, false), "GET", "/search/testdata/other", data)
	assert.Contains(t, body, "No results")
}

func TestTitleSearch(t *testing.T) {
	// working in the main directory
	index.reset()
	index.load()

	items, more := search("title:readme", "", "", 1, false)
	assert.Equal(t, 0, len(items), "no page found")
	assert.False(t, more)

	items, more = search("title:wel", "", "", 1, false) // README also contains "wel"
	assert.Equal(t, 1, len(items), "one page found")
	assert.Equal(t, "index", items[0].Name, "Welcome to Oddµ")
	assert.Greater(t, items[0].Score, 0, "matches result in a score")
	assert.False(t, more)

	items, more = search("wel", "", "", 1, false)
	assert.Greater(t, len(items), 1, "two pages found")
	assert.False(t, more)
}

func TestBlogSearch(t *testing.T) {
	cleanup(t, "testdata/grep")
	p := &Page{Name: "testdata/grep/2023-09-25", Body: []byte(`# Back then

I check the git log
Was it 2015
We met in the park?`)}
	p.save()

	items, _ := search("blog:false", "", "", 1, false)
	for _, item := range items {
		assert.NotEqual(t, "Back then", item.Title, item.Name)
	}

	items, _ = search("blog:true", "", "", 1, false)
	assert.Equal(t, 1, len(items), "one blog page found")
	assert.Equal(t, "Back then", items[0].Title, items[0].Name)
}

func TestHashtagSearch(t *testing.T) {
	cleanup(t, "testdata/hashtag")

	p := &Page{Name: "testdata/hashtag/Haiku", Body: []byte("# Haikus\n")}
	p.save()

	p = &Page{Name: "testdata/hashtag/2023-10-28", Body: []byte(`# Tea

My tongue is on fire
It looked so calm and peaceful
A quick sip too quick

#Haiku`)}
	p.save()

	items, _ := search("#Haiku", "testdata/hashtag", "", 1, false)
	assert.Equal(t, 2, len(items), "two pages found")
	assert.Equal(t, "Haikus", items[0].Title, items[0].Name)
	assert.Equal(t, "Tea", items[1].Title, items[1].Name)
}

func TestImageSearch(t *testing.T) {
	cleanup(t, "testdata/images")

	p := &Page{Name: "testdata/images/2024-07-21", Body: []byte(`# 2024-07-21 Pictures

![phone call](2024-07-21.jpg)

Pictures in the box
Tiny windows to our past
Where are you, my love?

`)}
	p.save()

	q := &Page{Name: "testdata/images/2024-07-22", Body: []byte(`# 2024-07-22 The Moon

When the night is light
Behind clouds the moon is bright
Please call me, my love.
`)}
	q.save()

	items, _ := search("call", "testdata/images", "", 1, false)
	assert.Equal(t, 2, len(items), "two pages found")

	assert.Equal(t, "2024-07-21 Pictures", items[0].Title)
	assert.Equal(t, "2024-07-22 The Moon", items[1].Title)

	assert.NotEmpty(t, items[0].Images)
	assert.Equal(t, "phone call", items[0].Images[0].Title)
	assert.Equal(t, "phone <b>call</b>", string(items[0].Images[0].Html))
	assert.Equal(t, "testdata/images/2024-07-21.jpg", items[0].Images[0].Name)

	assert.Empty(t, items[1].Images)
}

func TestSearchQuestionmark(t *testing.T) {
	cleanup(t, "testdata/question")
	p := &Page{Name: "testdata/question/Odd?", Body: []byte(`# Even?

We look at the plants.
They need water. We need us.
The silence streches.`)}
	p.save()
	data := url.Values{}
	data.Set("q", "look")
	body := assert.HTTPBody(makeHandler(searchHandler, false), "GET", "/search/", data)
	assert.Contains(t, body, "We <b>look</b>")
	assert.NotContains(t, body, "Odd?")
	assert.Contains(t, body, "Even?")
}

func TestSearchPagination(t *testing.T) {
	cleanup(t, "testdata/pagination")
	index.load()
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, r := range alphabet {
		s := fmt.Sprintf("secret%c secretX", r)
		p := &Page{Name: "testdata/pagination/" + string(r), Body: []byte(s)}
		p.save()
	}

	items, more := search("secretA", "", "", 1, false)
	assert.Equal(t, 1, len(items), "one page found, %v", items)
	assert.Equal(t, "testdata/pagination/A", items[0].Name)
	assert.False(t, more)

	items, more = search("secretX", "", "", 1, false)
	assert.Equal(t, itemsPerPage, len(items))
	assert.Equal(t, "testdata/pagination/A", items[0].Name)
	assert.Equal(t, "testdata/pagination/T", items[itemsPerPage-1].Name)
	assert.True(t, more)

	items, more = search("secretX", "", "", 2, false)
	assert.Equal(t, 6, len(items))
	assert.Equal(t, "testdata/pagination/U", items[0].Name)
	assert.Equal(t, "testdata/pagination/Z", items[5].Name)
	assert.False(t, more)
}
