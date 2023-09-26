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
	for _, s := range []string{"Alex", "Berta", "Chris", "2015-06-14", "2023-09-26"} {
		index.titles[s] = s
	}
	index.Unlock()
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

func TestSearch(t *testing.T) {
	data := url.Values{}
	data.Set("q", "oddµ")
	assert.Contains(t,
		assert.HTTPBody(searchHandler, "GET", "/search", data), "Welcome")
}

func TestTitleSearch(t *testing.T) {
	items, more := search("title:readme", 1)
	assert.Equal(t, 0, len(items), "no page found")
	assert.False(t, more)

	items, more = search("title:to", 1)
	assert.Equal(t, 1, len(items), "one page found")
	assert.Equal(t, "index", items[0].Name, "Welcome to Oddµ")
	assert.False(t, more)
}

func TestBlogSearch(t *testing.T) {
	cleanup(t, "testdata/grep")
	p := &Page{Name: "testdata/grep/2023-09-25", Body: []byte(`# Back then

I check the git log
Was it 2015
We met in the park?`)}
	p.save()

	items, _ := search("blog:false", 1)
	for _, item := range items {
		assert.NotEqual(t, "Back then", item.Title, item.Name)
	}

	items, _ = search("blog:true", 1)
	assert.Equal(t, 1, len(items), "one blog page found")
	assert.Equal(t, "Back then", items[0].Title, items[0].Name)
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
	body := assert.HTTPBody(searchHandler, "GET", "/search", data)
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

	items, more := search("secretA", 1)
	assert.Equal(t, 1, len(items), "one page found, %v", items)
	assert.Equal(t, "testdata/pagination/A", items[0].Name)
	assert.False(t, more)

	items, more = search("secretX", 1)
	assert.Equal(t, itemsPerPage, len(items))
	assert.Equal(t, "testdata/pagination/A", items[0].Name)
	assert.Equal(t, "testdata/pagination/T", items[itemsPerPage-1].Name)
	assert.True(t, more)

	items, more = search("secretX", 2)
	assert.Equal(t, 6, len(items))
	assert.Equal(t, "testdata/pagination/U", items[0].Name)
	assert.Equal(t, "testdata/pagination/Z", items[5].Name)
	assert.False(t, more)
}
