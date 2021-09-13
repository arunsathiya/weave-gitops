package middleware_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
)

var (
	jwtClient      *authfakes.FakeJWTClient
	defaultHandler http.HandlerFunc
)

var _ = Describe("WithProviderToken", func() {
	_ = BeforeEach(func() {
		jwtClient = &authfakes.FakeJWTClient{
			VerifyJWTStub: func(s string) (*auth.Claims, error) {
				return &auth.Claims{
					ProviderToken: "provider-token",
				}, nil
			},
		}

		defaultHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	})

	It("does nothing when no token is passed", func() {
		midware := middleware.WithProviderToken(jwtClient, defaultHandler)

		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com/", nil)
		res := httptest.NewRecorder()
		midware.ServeHTTP(res, req)

		Expect(res.Result().StatusCode).To(Equal(http.StatusOK))
		Expect(jwtClient.VerifyJWTCallCount()).To(Equal(0))
	})

	It("extracts JWT token from the header", func() {
		midware := middleware.WithProviderToken(jwtClient, defaultHandler)

		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com", nil)
		req.Header.Add("Authorization", "token my-jwt-token")

		res := httptest.NewRecorder()

		midware.ServeHTTP(res, req)

		Expect(jwtClient.VerifyJWTArgsForCall(0)).To(Equal("my-jwt-token"))
	})

	It("returns forbidden when token is no valid", func() {
		jwtClient.VerifyJWTStub = func(s string) (*auth.Claims, error) {
			return nil, auth.ErrUnauthorizedToken
		}

		midware := middleware.WithProviderToken(jwtClient, defaultHandler)
		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com", nil)
		req.Header.Add("Authorization", "token my-jwt-token")

		res := httptest.NewRecorder()

		midware.ServeHTTP(res, req)

		Expect(res.Result().StatusCode).To(Equal(http.StatusUnauthorized))
	})

	It("passes the provider token into the context", func() {
		var request *http.Request

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			request = r
		})

		midware := middleware.WithProviderToken(jwtClient, next)
		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com", nil)
		req.Header.Add("Authorization", "token my-jwt-token")

		res := httptest.NewRecorder()

		midware.ServeHTTP(res, req)

		providerToken, err := middleware.ExtractProviderToken(request.Context())
		Expect(err).ToNot(HaveOccurred())

		Expect(providerToken.AccessToken).To(Equal("provider-token"))
	})
})

var _ = Describe("ExtractProviderToken", func() {
	_ = BeforeEach(func() {
		jwtClient = &authfakes.FakeJWTClient{
			VerifyJWTStub: func(s string) (*auth.Claims, error) {
				return &auth.Claims{
					ProviderToken: "",
				}, nil
			},
		}
	})

	It("errors out when no provider token in context", func() {
		var request *http.Request
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			request = r
		})

		midware := middleware.WithProviderToken(jwtClient, next)
		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com", nil)
		req.Header.Add("Authorization", "token my-jwt-token")

		res := httptest.NewRecorder()

		midware.ServeHTTP(res, req)

		_, err := middleware.ExtractProviderToken(request.Context())
		Expect(err).To(MatchError("no token specified"))
	})
})