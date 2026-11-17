package bridge

import (
	"encoding/json"
	"testing"
)

func TestSearchReturnsFavoriteFirstAndHighlightIndices(t *testing.T) {
	commands := []struct {
		Name     string   `json:"name"`
		Tags     []string `json:"tags"`
		Favorite bool     `json:"is_favorite"`
	}{
		{Name: "git status"},
		{Name: "git stash", Favorite: true},
	}
	result, err := Search(commands, "git")
	if err != nil {
		t.Fatal(err)
	}
	var hits []struct {
		Index   int   `json:"index"`
		Indices []int `json:"indices"`
	}
	if err := json.Unmarshal([]byte(result), &hits); err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 || hits[0].Index != 1 || len(hits[0].Indices) == 0 {
		t.Fatalf("unexpected fuzzy search result: %s", result)
	}
}

func TestTemplateParseAndRenderAcrossCABI(t *testing.T) {
	parsed, err := Parse("rm -rf {{path:/tmp}}")
	if err != nil {
		t.Fatal(err)
	}
	var info struct {
		Dangerous bool `json:"dangerous"`
		Variables []struct {
			Name    string `json:"name"`
			Default string `json:"default"`
		} `json:"variables"`
	}
	if err := json.Unmarshal([]byte(parsed), &info); err != nil {
		t.Fatal(err)
	}
	if !info.Dangerous || len(info.Variables) != 1 || info.Variables[0].Default != "/tmp" {
		t.Fatalf("unexpected template details: %s", parsed)
	}
	rendered, err := Render("echo {{name:world}}", map[string]string{"name": "snipbox"})
	if err != nil {
		t.Fatal(err)
	}
	var command string
	if err := json.Unmarshal([]byte(rendered), &command); err != nil {
		t.Fatal(err)
	}
	if command != "echo snipbox" {
		t.Fatalf("rendered command = %q", command)
	}
}

func TestEngineErrorsDoNotBecomeSuccessfulResults(t *testing.T) {
	if _, err := checked(`{"error":"invalid template input"}`); err == nil {
		t.Fatal("engine error response was accepted as a successful result")
	}
}
