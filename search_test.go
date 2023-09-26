package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"slices"
	"testing"
)

func TestSortNames(t *testing.T) {
	index.Lock()
	for _, s := range []string{"Alex", "Berta", "Chris", "2022", "2023"} {
		index.titles[s] = s
	}
	index.Unlock()
	terms := []string{"Z"}
	fn := sortNames(terms)
	assert.Equal(t, 1, fn("Berta", "Alex"), "B is after A")
	assert.Equal(t, -1, fn("Alex", "Berta"), "A is before B")
	assert.Equal(t, 0, fn("Berta", "Berta"), "B and B are equal")
	assert.Equal(t, -1, fn("2023", "Alex"), "numbers before letters")
	assert.Equal(t, 1, fn("Alex", "2023"), "numbers after letters")
	assert.Equal(t, -1, fn("2023", "2022"), "higher numbers before lower numbers")
	assert.Equal(t, 1, fn("2022", "2023"), "lower numbers after higher numbers")

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

// wipes testdata
func TestSearchQuestionmark(t *testing.T) {
	_ = os.RemoveAll("testdata")
	p := &Page{Name: "testdata/Odd?", Body: []byte(`# Even?

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
	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}

// wipes testdata
func TestSearchPagination(t *testing.T) {
	_ = os.RemoveAll("testdata")
	index.load()
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, r := range alphabet {
		s := fmt.Sprintf("secret%c secretX", r)
		p := &Page{Name: "testdata/" + string(r), Body: []byte(s)}
		p.save()
	}

	items, more := search("secretA", 1)
	assert.Equal(t, 1, len(items), "one page found, %v", items)
	assert.Equal(t, "testdata/A", items[0].Name)
	assert.False(t, more)

	items, more = search("secretX", 1)
	assert.Equal(t, itemsPerPage, len(items))
	assert.Equal(t, "testdata/A", items[0].Name)
	assert.Equal(t, "testdata/T", items[itemsPerPage-1].Name)
	assert.True(t, more)

	items, more = search("secretX", 2)
	assert.Equal(t, 6, len(items))
	assert.Equal(t, "testdata/U", items[0].Name)
	assert.Equal(t, "testdata/Z", items[5].Name)
	assert.False(t, more)

	t.Cleanup(func() {
		_ = os.RemoveAll("testdata")
	})
}
