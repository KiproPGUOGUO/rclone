package serve

import (
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/rclone/rclone/cmd/serve/httplib/serve/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetTemplate(t *testing.T) *template.Template {
	htmlTemplate, err := data.GetTemplate()
	require.NoError(t, err)
	return htmlTemplate
}

func TestNewDirectory(t *testing.T) {
	d := NewDirectory("z", GetTemplate(t))
	assert.Equal(t, "z", d.DirRemote)
	assert.Equal(t, "Directory listing of /z", d.Title)
}

func TestSetQuery(t *testing.T) {
	d := NewDirectory("z", GetTemplate(t))
	assert.Equal(t, "", d.Query)
	d.SetQuery(url.Values{"potato": []string{"42"}})
	assert.Equal(t, "?potato=42", d.Query)
	d.SetQuery(url.Values{})
	assert.Equal(t, "", d.Query)
}

func TestAddEntry(t *testing.T) {
	var d = NewDirectory("z", GetTemplate(t))
	d.AddEntry("", true)
	d.AddEntry("dir", true)
	d.AddEntry("a/b/c/d.txt", false)
	d.AddEntry("a/b/c/colon:colon.txt", false)
	d.AddEntry("\"quotes\".txt", false)
	assert.Equal(t, []DirEntry{
		{remote: "", URL: "/", Leaf: "/"},
		{remote: "dir", URL: "dir/", Leaf: "dir/"},
		{remote: "a/b/c/d.txt", URL: "d.txt", Leaf: "d.txt"},
		{remote: "a/b/c/colon:colon.txt", URL: "./colon:colon.txt", Leaf: "colon:colon.txt"},
		{remote: "\"quotes\".txt", URL: "%22quotes%22.txt", Leaf: "\"quotes\".txt"},
	}, d.Entries)

	// Now test with a query parameter
	d = NewDirectory("z", GetTemplate(t)).SetQuery(url.Values{"potato": []string{"42"}})
	d.AddEntry("file", false)
	d.AddEntry("dir", true)
	assert.Equal(t, []DirEntry{
		{remote: "file", URL: "file?potato=42", Leaf: "file"},
		{remote: "dir", URL: "dir/?potato=42", Leaf: "dir/"},
	}, d.Entries)
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	err := errors.New("help")
	Error("potato", w, "sausage", err)
	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, "sausage.\n", string(body))
}

func TestServe(t *testing.T) {
	d := NewDirectory("aDirectory", GetTemplate(t))
	d.AddEntry("file", false)
	d.AddEntry("dir", true)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/aDirectory/", nil)
	d.Serve(w, r)
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Directory listing of /aDirectory</title>
</head>
<body>
<h1>Directory listing of /aDirectory</h1>
<a href="file">file</a><br />
<a href="dir/">dir/</a><br />
</body>
</html>
`, string(body))
}
