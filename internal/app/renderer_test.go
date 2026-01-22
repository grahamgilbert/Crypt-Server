package app

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRendererRender(t *testing.T) {
	tmp := t.TempDir()
	layout := filepath.Join(tmp, "base.html")
	page := filepath.Join(tmp, "page.html")

	require.NoError(t, os.WriteFile(layout, []byte("{{define \"base\"}}Hello {{block \"content\" .}}{{end}}{{end}}"), 0o600))
	require.NoError(t, os.WriteFile(page, []byte("{{define \"content\"}}World{{end}}"), 0o600))

	renderer := NewRenderer(layout, tmp)
	recorder := httptest.NewRecorder()
	require.NoError(t, renderer.Render(recorder, "page", TemplateData{}))
	require.Equal(t, "Hello World", recorder.Body.String())
}
