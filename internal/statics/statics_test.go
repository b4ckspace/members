package statics

import "testing"

func TestStatics(t *testing.T) {
	fs := Statics()
	_, err := fs.Open("templates/base.html")
	if err != nil {
		t.Errorf("Could not creatte http.Filesystem: %v", err)
	}
}

func TestBaseTemplateExists(t *testing.T) {
	_, err := EmbeddedStaticFiles.Open("templates/base.html")
	if err != nil {
		t.Errorf("Mandatory templates not found: %v", err)
	}
}

func TestStaticFileExists(t *testing.T) {
	_, err := EmbeddedStaticFiles.Open("static/js/bootstrap.min.js")
	if err != nil {
		t.Errorf("Mandatory statics not found: %v", err)
	}
}
