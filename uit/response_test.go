package uit

import (
	"net/http/httptest"
	"testing"

	"github.com/openbyte-os/sdk-go/transport"
)

func TestNewResponse(t *testing.T) {
	r := Response()
	if r == nil {
		t.Fatal("expected non-nil Response")
	}
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := Response()
	r.NoContent()
	r.Write(w)

	if w.Code != 204 {
		t.Errorf("status code = %d, want 204", w.Code)
	}
}

func TestRefresh(t *testing.T) {
	w := httptest.NewRecorder()
	r := Response()
	r.Refresh()
	r.Write(w)
	if got := w.Header().Get(transport.ResponseRefresh); got != "self" {
		t.Errorf("ResponseRefresh = %q, want %q", got, "self")
	}
}

func TestRefreshFragment(t *testing.T) {
	t.Run("single fragment", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := Response()
		r.RefreshFragment("frag1")
		r.Write(w)
		if got := w.Header().Get(transport.ResponseRefresh); got != "frag1" {
			t.Errorf("ResponseRefresh = %q, want %q", got, "frag1")
		}
	})

	t.Run("multiple fragments", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := Response()
		r.RefreshFragment("frag1", "frag2", "frag3")
		r.Write(w)
		if got := w.Header().Get(transport.ResponseRefresh); got != "frag1,frag2,frag3" {
			t.Errorf("ResponseRefresh = %q, want %q", got, "frag1,frag2,frag3")
		}
	})

	t.Run("no fragments", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := Response()
		r.RefreshFragment()
		r.Write(w)
		if got := w.Header().Get(transport.ResponseRefresh); got != "" {
			t.Errorf("ResponseRefresh = %q, want empty", got)
		}
	})
}

func TestRefreshSpace(t *testing.T) {
	t.Run("single uri", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := Response()
		r.RefreshSpace("/dashboard")
		r.Write(w)
		if got := w.Header().Get(transport.ResponseRefresh); got != "/dashboard" {
			t.Errorf("ResponseRefresh = %q, want %q", got, "/dashboard")
		}
	})

	t.Run("multiple uris", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := Response()
		r.RefreshSpace("/space1", "/space2")
		r.Write(w)
		if got := w.Header().Get(transport.ResponseRefresh); got != "/space1,/space2" {
			t.Errorf("ResponseRefresh = %q, want %q", got, "/space1,/space2")
		}
	})
}

func TestChainableResponse(t *testing.T) {
	w := httptest.NewRecorder()
	Response().Refresh().NoContent().Write(w)

	if w.Code != 204 {
		t.Errorf("status code = %d, want 204", w.Code)
	}

	if got := w.Header().Get(transport.ResponseRefresh); got != "self" {
		t.Errorf("ResponseRefresh = %q, want %q", got, "self")
	}
}

func TestRefreshReferer(t *testing.T) {
	w := httptest.NewRecorder()
	r := Response()
	r.RefreshReferer()
	r.Write(w)
	if got := w.Header().Get(transport.ResponseRefresh); got != "referer" {
		t.Errorf("ResponseRefresh = %q, want %q", got, "referer")
	}
}

func TestCloseModal(t *testing.T) {
	w := httptest.NewRecorder()
	r := Response()
	r.CloseModal()
	r.Write(w)
	if got := w.Header().Get(transport.ResponseCloseModal); got != "1" {
		t.Errorf("ResponseCloseModal = %q, want %q", got, "1")
	}
}
