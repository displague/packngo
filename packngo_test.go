package packngo

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	packetURLEnvVar   = "PACKET_API_URL"
	packngoAccTestVar = "PACKNGO_TEST_ACTUAL_API"
	testProjectPrefix = "PACKNGO_TEST_DELME_2d768716_"
	testFacilityVar   = "PACKNGO_TEST_FACILITY"
)

func testFacility() string {
	envFac := os.Getenv(testFacilityVar)
	if envFac != "" {
		return envFac
	}
	return "ewr1"
}

func randString8() string {
	n := 8
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("acdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// MockClient makes it simpler to test the Client
type MockClient struct {
	fnNewRequest          func(method, path string, body interface{}) (*retryablehttp.Request, error)
	fnDo                  func(req *retryablehttp.Request, v interface{}) (*Response, error)
	fnDoRequest           func(method, path string, body, v interface{}) (*Response, error)
	fnDoRequestWithHeader func(method string, headers map[string]string, path string, body, v interface{}) (*Response, error)
}

var _ requestDoer = &MockClient{}

// NewRequest uses the mock NewRequest function
func (mc *MockClient) NewRequest(method, path string, body interface{}) (*retryablehttp.Request, error) {
	return mc.fnNewRequest(method, path, body)
}

// Do uses the mock Do function
func (mc *MockClient) Do(req *retryablehttp.Request, v interface{}) (*Response, error) {
	return mc.fnDo(req, v)
}

// DoRequest uses the mock DoRequest function
func (mc *MockClient) DoRequest(method, path string, body, v interface{}) (*Response, error) {
	return mc.fnDoRequest(method, path, body, v)
}

// DoRequestWithHeader uses the mock DoRequestWithHeader function
func (mc *MockClient) DoRequestWithHeader(method string, headers map[string]string, path string, body, v interface{}) (*Response, error) {
	return mc.fnDoRequestWithHeader(method, headers, path, body, v)
}

func setupWithProject(t *testing.T) (*Client, string, func()) {
	c := setup(t)
	p, _, err := c.Projects.Create(&ProjectCreateRequest{Name: testProjectPrefix + randString8()})
	if err != nil {
		t.Fatal(err)
	}

	return c, p.ID, func() {
		_, err := c.Projects.Delete(p.ID)
		if err != nil {
			panic(fmt.Errorf("while deleting %s: %s", p, err))
		}
	}

}

func skipUnlessAcceptanceTestsAllowed(t *testing.T) {
	if os.Getenv(packngoAccTestVar) == "" {
		t.Skipf("%s is not set", packngoAccTestVar)
	}
}

func setup(t *testing.T) *Client {
	apiToken := os.Getenv(packetTokenEnvVar)
	if apiToken == "" {
		t.Fatalf("If you want to run packngo test, you must export %s.", packetTokenEnvVar)
	}
	apiURL := os.Getenv(packetURLEnvVar)
	if apiURL == "" {
		apiURL = baseURL
	}
	c, err := NewClientWithBaseURL("packngo test", apiToken, nil, apiURL)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func projectTeardown(c *Client) {
	ps, _, err := c.Projects.List(nil)
	if err != nil {
		panic(fmt.Errorf("while teardown: %s", err))
	}
	for _, p := range ps {
		if strings.HasPrefix(p.Name, testProjectPrefix) {
			_, err := c.Projects.Delete(p.ID)
			if err != nil {
				panic(fmt.Errorf("while deleting %s: %s", p, err))
			}
		}
	}
}

func organizationTeardown(c *Client) {
	ps, _, err := c.Organizations.List(nil)
	if err != nil {
		panic(fmt.Errorf("while teardown: %s", err))
	}
	for _, p := range ps {
		if strings.HasPrefix(p.Name, testProjectPrefix) {
			_, err := c.Organizations.Delete(p.ID)
			if err != nil {
				panic(fmt.Errorf("while deleting %s: %s", p, err))
			}
		}
	}
}

func TestAccInvalidCredentials(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	c := NewClientWithAuth("packngo test", "wrongApiToken", nil)
	_, r, expectedErr := c.Projects.List(nil)
	matched, err := regexp.MatchString(".*Invalid.*", expectedErr.Error())
	if err != nil {
		t.Fatalf("Err while matching err string from response err %s: %s", expectedErr, err)
	}
	if r.StatusCode != 401 {
		t.Fatalf("Expected 401 as response code, got: %d", r.StatusCode)
	}

	if !matched {
		t.Fatalf("Unexpected error string: %s", expectedErr)
	}

}
