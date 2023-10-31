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
			fmt.Println("calling?")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(path))
		}
	}

	t.Run("It panics if the path is empty", func(t *testing.T) {
		router := New()

		g.Expect(func() {
			router.HandleFunc("POST", "", foundHandler(""))
		}).To(Panic())
	})

	t.Run("It can add the capture all path '/'", func(t *testing.T) {
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

	t.Run("It can add a single path url'", func(t *testing.T) {
		path := "/v1"

		router := New()
		router.HandleFunc("POST", path, foundHandler(path))

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		fmt.Println("DSL:", testServer.URL)
		request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testServer.URL, path), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := client.Do(request)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})
}
