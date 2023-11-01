package urlrouter

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
)

func TestInternalFunction_splitPaths(t *testing.T) {
	g := NewGomegaWithT(t)

	t.Run("It splits an empty string into nil", func(t *testing.T) {
		paths, wildcard := splitPaths("")
		g.Expect(paths).To(BeNil())
		g.Expect(wildcard).To(BeFalse())
	})

	t.Run("It splits a single '/' to nil, but the wildcard is true", func(t *testing.T) {
		paths, wildcard := splitPaths("/")
		g.Expect(paths).To(Equal([]string{"/"}))
		g.Expect(wildcard).To(BeTrue())
	})

	t.Run("It splits '/abc' to the strings '/', 'abc' and wildcard is false", func(t *testing.T) {
		paths, wildcard := splitPaths("/abc")
		g.Expect(paths).To(Equal([]string{"/", "abc"}))
		g.Expect(wildcard).To(BeFalse())
	})

	t.Run("It splits '/abc/' to the strings '/', 'abc', '/' and wildcard is true", func(t *testing.T) {
		paths, wildcard := splitPaths("/v1/")
		g.Expect(paths).To(Equal([]string{"/", "v1", "/"}))
		g.Expect(wildcard).To(BeTrue())
	})

	t.Run("It splits a multiple paths 'abc/def/hij' into multiple strings and wildcard is false", func(t *testing.T) {
		paths, wildcard := splitPaths("abc/def/hij")
		g.Expect(paths).To(Equal([]string{"abc", "/", "def", "/", "hij"}))
		g.Expect(wildcard).To(BeFalse())
	})

	t.Run("It splits a multiple paths '/abc/def/hij/' into multiple strings and wildcard is true", func(t *testing.T) {
		paths, wildcard := splitPaths("/abc/def/hij/")
		g.Expect(paths).To(Equal([]string{"/", "abc", "/", "def", "/", "hij", "/"}))
		g.Expect(wildcard).To(BeTrue())
	})

	t.Run("It splits a two '//' into 2 strings with a true wildcard", func(t *testing.T) {
		paths, wildcard := splitPaths("//")
		g.Expect(paths).To(Equal([]string{"/", "/"}))
		g.Expect(wildcard).To(BeTrue())
	})
}

func TestRouter_defults(t *testing.T) {
	g := NewGomegaWithT(t)

	foundHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	t.Run("It panics if the path is empty", func(t *testing.T) {
		router := New()
		g.Expect(func() { router.HandleFunc("POST", "", foundHandler) }).To(Panic())
	})

	t.Run("It panics if the handler is empty", func(t *testing.T) {
		router := New()
		g.Expect(func() { router.HandleFunc("POST", "/something", nil) }).To(Panic())
	})
}

func TestRouter_UrlPathPatterns(t *testing.T) {
	g := NewGomegaWithT(t)

	client := &http.Client{}

	foundHandler := func(path string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(path))
		}
	}

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

		request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testServer.URL, path), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := client.Do(request)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})

	t.Run("It can add multiple single path url'", func(t *testing.T) {
		router := New()

		for i := 0; i < 10; i++ {
			path := fmt.Sprintf("/v%d", i)
			router.HandleFunc("POST", path, foundHandler(path))
		}

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		for i := 0; i < 10; i++ {
			path := fmt.Sprintf("/v%d", i)

			request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testServer.URL, path), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := io.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(body)).To(Equal(path))
		}
	})

	t.Run("It can add a url path with multiple seperators'", func(t *testing.T) {
		path := "/v1/applications/local/unix"

		router := New()
		router.HandleFunc("POST", path, foundHandler(path))

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testServer.URL, path), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := client.Do(request)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})

	t.Run("It can add multiple url paths with multiple seperators'", func(t *testing.T) {
		router := New()

		for i := 0; i < 100; i++ {
			path := fmt.Sprintf("/v%d/%d/%d", i%2, i%5, i)
			router.HandleFunc("POST", path, foundHandler(path))
		}

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		for i := 0; i < 100; i++ {
			path := fmt.Sprintf("/v%d/%d/%d", i%2, i%5, i)

			request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testServer.URL, path), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := io.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(body)).To(Equal(path))
		}
	})

	t.Run("Context behaviors of paths", func(t *testing.T) {
		t.Run("It returns a 404 if no paths match appropriately", func(t *testing.T) {
			path := "/some/path"

			router := New()
			router.HandleFunc("POST", path, foundHandler(path))

			testServer := httptest.NewServer(router)
			defer testServer.Close()

			// single path
			request, err := http.NewRequest("POST", fmt.Sprintf("%s/not_found", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusNotFound))

			// match first part, but not second
			request, err = http.NewRequest("POST", fmt.Sprintf("%s/some.bad_path", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err = client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		t.Run("It allows paths ending in a '/' to wildcard match a path not captured with explicit paths", func(t *testing.T) {
			catchAll := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`catch all`))
			}

			fullmMatch := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`full match`))
			}

			router := New()
			router.HandleFunc("POST", "/", catchAll)
			router.HandleFunc("POST", "/v1/full_match", fullmMatch)

			//testServer := httptest.NewServer(mux)
			testServer := httptest.NewServer(router)
			defer testServer.Close()

			// catchAll works because it is jus the '/' and not an exact match
			request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/full_match/something", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := io.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(body)).To(Equal("catch all"))

			// full_match catches the exact url
			request, err = http.NewRequest("POST", fmt.Sprintf("%s/v1/full_match", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err = client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err = io.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(body)).To(Equal("full match"))
		})

		t.Run("It matches the the wildcard that has the most in common paths", func(t *testing.T) {
			catchAll := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`catch all`))
			}

			catchV1 := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`catch v1`))
			}

			fullmMatch := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`full match`))
			}

			router := New()
			router.HandleFunc("POST", "/", catchAll)
			router.HandleFunc("POST", "/v1/", catchV1)
			router.HandleFunc("POST", "/v1/full_match", fullmMatch)

			testServer := httptest.NewServer(router)
			defer testServer.Close()

			// v1/ catches anything after the matcher rahter than '/'
			//request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/full_match/something", testServer.URL), nil)
			//g.Expect(err).ToNot(HaveOccurred())
			//
			//resp, err := client.Do(request)
			//g.Expect(err).ToNot(HaveOccurred())
			//g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
			//
			//body, err := io.ReadAll(resp.Body)
			//g.Expect(err).ToNot(HaveOccurred())
			//g.Expect(string(body)).To(Equal("catch v1"))

			// still catches the /v1 path
			request, err := http.NewRequest("POST", fmt.Sprintf("%s/v1", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := io.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(body)).To(Equal("catch all"))

			// full_match catches the exact url properly
			request, err = http.NewRequest("POST", fmt.Sprintf("%s/v1/full_match", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err = client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err = io.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(body)).To(Equal("full match"))
		})
	})
}

func TestRouter_NamedParameters(t *testing.T) {
	g := NewGomegaWithT(t)

	client := &http.Client{}

	t.Run("It translates a named parameter into a request's context'", func(t *testing.T) {
		path := "/:name"

		var namedParameters = map[string]string{}
		foundHandler := func(w http.ResponseWriter, r *http.Request) {
			namedParameters = GetNamedParamters(r.Context())

			w.WriteHeader(http.StatusOK)
		}

		router := New()
		router.HandleFunc("POST", path, foundHandler)

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		request, err := http.NewRequest("POST", fmt.Sprintf("%s/the_name", testServer.URL), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := client.Do(request)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
		g.Expect(namedParameters).To(Equal(map[string]string{"name": "the_name"}))
	})

	t.Run("It can add multiple named parameters together'", func(t *testing.T) {
		path := "/:1/:2/:3"

		var namedParameters = map[string]string{}
		foundHandler := func(w http.ResponseWriter, r *http.Request) {
			namedParameters = GetNamedParamters(r.Context())

			w.WriteHeader(http.StatusOK)
		}

		router := New()
		router.HandleFunc("POST", path, foundHandler)

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		request, err := http.NewRequest("POST", fmt.Sprintf("%s/one/two/three", testServer.URL), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := client.Do(request)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
		g.Expect(namedParameters).To(Equal(map[string]string{"1": "one", "2": "two", "3": "three"}))
	})

	t.Run("Context behaviors of paths", func(t *testing.T) {
		t.Run("It overwrites the key words at the same path level", func(t *testing.T) {
			var namedParameters = map[string]string{}
			foundHandler := func(w http.ResponseWriter, r *http.Request) {
				namedParameters = GetNamedParamters(r.Context())

				w.WriteHeader(http.StatusOK)
			}

			router := New()
			router.HandleFunc("POST", "/:name", foundHandler)
			router.HandleFunc("POST", "/:value2", foundHandler)

			testServer := httptest.NewServer(router)
			defer testServer.Close()

			request, err := http.NewRequest("POST", fmt.Sprintf("%s/the_name", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
			g.Expect(namedParameters).To(Equal(map[string]string{"value2": "the_name"}))
		})

		t.Run("It overwrittes the key workds at multiple path levels", func(*testing.T) {
			var namedParameters = map[string]string{}
			foundHandler := func(w http.ResponseWriter, r *http.Request) {
				namedParameters = GetNamedParamters(r.Context())

				w.WriteHeader(http.StatusOK)
			}

			router := New()
			router.HandleFunc("POST", "/:1/:2/:3", foundHandler)
			router.HandleFunc("POST", "/:new1/:new2/:new3", foundHandler)

			testServer := httptest.NewServer(router)
			defer testServer.Close()

			request, err := http.NewRequest("POST", fmt.Sprintf("%s/one/two/three", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
			g.Expect(namedParameters).To(Equal(map[string]string{"new1": "one", "new2": "two", "new3": "three"}))
		})

		t.Run("It allows paths ending in a '/' to wildcard match a path not captured with explicit paths", func(t *testing.T) {
			var namedParameters = map[string]string{}
			wildcardHandler := func(w http.ResponseWriter, r *http.Request) {
				namedParameters = GetNamedParamters(r.Context())
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("wildcardHandler"))
			}
			exactSmallHandler := func(w http.ResponseWriter, r *http.Request) {
				namedParameters = GetNamedParamters(r.Context())
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("exactSmallHandler"))
			}
			exactLargeHandler := func(w http.ResponseWriter, r *http.Request) {
				namedParameters = GetNamedParamters(r.Context())
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("exactLargeHandler"))
			}

			router := New()
			router.HandleFunc("POST", "/:1", exactSmallHandler)
			router.HandleFunc("POST", "/:1/", wildcardHandler)
			router.HandleFunc("POST", "/:1/:2", exactLargeHandler)

			testServer := httptest.NewServer(router)
			defer testServer.Close()

			// match the small handler
			request, err := http.NewRequest("POST", fmt.Sprintf("%s/test_small", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
			g.Expect(namedParameters).To(Equal(map[string]string{"1": "test_small"}))

			body, err := io.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(body)).To(Equal("exactSmallHandler"))

			// match wildcard handler
			request, err = http.NewRequest("POST", fmt.Sprintf("%s/wildcard/or/something/you/know", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err = client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
			g.Expect(namedParameters).To(Equal(map[string]string{"1": "wildcard"}))

			body, err = io.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(body)).To(Equal("wildcardHandler"))

			// match large handler
			request, err = http.NewRequest("POST", fmt.Sprintf("%s/not/wildcard", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err = client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
			g.Expect(namedParameters).To(Equal(map[string]string{"1": "not", "2": "wildcard"}))

			body, err = io.ReadAll(resp.Body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(string(body)).To(Equal("exactLargeHandler"))
		})
	})
}

func TestRouter_Mixed_URL_and_Nmed_Paths(t *testing.T) {
	g := NewGomegaWithT(t)

	client := &http.Client{}

	t.Run("Context url paths before named paramters", func(t *testing.T) {
		t.Run("It can parse a single url properly", func(t *testing.T) {
			path := "/initial/:name"

			var namedParameters = map[string]string{}
			foundHandler := func(w http.ResponseWriter, r *http.Request) {
				namedParameters = GetNamedParamters(r.Context())
				w.WriteHeader(http.StatusOK)
			}

			router := New()
			router.HandleFunc("POST", path, foundHandler)

			testServer := httptest.NewServer(router)
			defer testServer.Close()

			request, err := http.NewRequest("POST", fmt.Sprintf("%s/initial/the_name", testServer.URL), nil)
			g.Expect(err).ToNot(HaveOccurred())

			resp, err := client.Do(request)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
			g.Expect(namedParameters).To(Equal(map[string]string{"name": "the_name"}))
		})

		t.Run("It can parse a multiple url properly", func(t *testing.T) {
			router := New()

			var namedParameters = map[string]string{}
			for i := 0; i < 10; i++ {
				path := fmt.Sprintf("/%d/:name", i)

				foundHandler := func(w http.ResponseWriter, r *http.Request) {
					namedParameters = GetNamedParamters(r.Context())
					w.WriteHeader(http.StatusOK)
				}

				router.HandleFunc("POST", path, foundHandler)
			}

			testServer := httptest.NewServer(router)
			defer testServer.Close()

			for i := 0; i < 10; i++ {
				path := fmt.Sprintf("/%d/the_name", i)

				request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", testServer.URL, path), nil)
				g.Expect(err).ToNot(HaveOccurred())

				resp, err := client.Do(request)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
				g.Expect(namedParameters).To(Equal(map[string]string{"name": "the_name"}))
				namedParameters = map[string]string{}
			}
		})
	})

	t.Run("It parses a named parameter before a url path", func(t *testing.T) {
		path := "/:name/update"

		var namedParameters = map[string]string{}
		foundHandler := func(w http.ResponseWriter, r *http.Request) {
			namedParameters = GetNamedParamters(r.Context())
			w.WriteHeader(http.StatusOK)
		}

		router := New()
		router.HandleFunc("POST", path, foundHandler)

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		request, err := http.NewRequest("POST", fmt.Sprintf("%s/the_name/update", testServer.URL), nil)
		g.Expect(err).ToNot(HaveOccurred())

		resp, err := client.Do(request)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
		g.Expect(namedParameters).To(Equal(map[string]string{"name": "the_name"}))
	})
}
