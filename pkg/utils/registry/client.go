package registry

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/distribution/distribution/reference"
	registryclient "github.com/distribution/distribution/registry/client"
	"github.com/distribution/distribution/registry/client/auth"
	"github.com/distribution/distribution/registry/client/transport"
	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/client/auth/challenge"
)

type Client struct {
	challengeManager challenge.Manager
	credStore        auth.CredentialStore
	opts             *Options
	baseURL          string
	httpClient       *http.Client
}

func NewClient(opts *Options) (*Client, error) {
	c := &Client{
		opts:             opts,
		challengeManager: challenge.NewSimpleManager(),
		credStore:        NewCredStore(opts),
		baseURL:          fmt.Sprintf("https://%s", opts.Server),
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
	httpsOK, err := c.tryHttps()
	if err != nil {
		return nil, err
	}
	if !httpsOK {
		c.baseURL = fmt.Sprintf("http://%s", opts.Server)
	}

	if err := c.tryEstablishChallenges(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) tryHttps() (bool, error) {
	endpointURL, err := url.Parse(c.baseURL)
	if err != nil {
		return false, err
	}

	endpointURL.Path = "/v2/"
	resp, err := c.httpClient.Get(endpointURL.String())
	if err != nil {
		if strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") {
			return false, nil
		}
		return false, err
	}
	resp.Body.Close()
	return true, nil
}

func (c *Client) NewRegistry() (registryclient.Registry, error) {
	roundTripper, err := c.GetRoundTripper("", CatalogAction)
	if err != nil {
		return nil, err
	}
	return registryclient.NewRegistry(c.baseURL, roundTripper)
}

func (c *Client) NewRepository(repo string, action Action) (distribution.Repository, error) {
	repoNamed, err := reference.WithName(repo)
	if err != nil {
		return nil, err
	}

	roundTripper, err := c.GetRoundTripper(repoNamed.String(), action)
	if err != nil {
		return nil, err
	}

	repository, err := registryclient.NewRepository(repoNamed, c.baseURL, roundTripper)
	if err != nil {
		return nil, err
	}
	return repository, nil
}

func (c *Client) tryEstablishChallenges() error {
	endpointURL, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}

	endpointURL.Path = "/v2/"
	challenges, err := c.challengeManager.GetChallenges(*endpointURL)
	if err != nil {
		return err
	}

	if len(challenges) > 0 {
		return nil
	}

	resp, err := c.httpClient.Get(endpointURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.challengeManager.AddResponse(resp); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetRoundTripper(scope string, action Action) (http.RoundTripper, error) {
	return transport.NewTransport(c.httpClient.Transport,
		auth.NewAuthorizer(c.challengeManager,
			auth.NewBasicHandler(c.credStore),
			auth.NewTokenHandler(c.httpClient.Transport, c.credStore, scope, string(action)))), nil
}
