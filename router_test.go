package urlrouter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
)

func TestRouter_SimplePathPatterns(t *testing.T) {
	g := NewGomegaWithT(t)

	client := &http.Client{}

	foundHandler := func(path string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(path))
		}
	}

	t.Run("It can add just a slash", func(t *testing.T) {
		path := "/"

		router := New()
		router.HandleFunc("POST", path, foundHandler(path))

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		request, err := http.NewRequest("POST", fmt.Sprintf("%s/", testServer.URL), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := client.Do(request)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})
}
