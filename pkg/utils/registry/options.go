package registry

import (
	"context"
	"fmt"

	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

type Options struct {
	Username    string
	Password    string
	Server      string
	Repositiory string
	Tag         string
	Digest      digest.Digest
	Destination string
	Ctx         context.Context
}

func (opts *Options) ParseReference(ref string) error {
	named, err := reference.ParseDockerRef(ref)
	if err != nil {
		return fmt.Errorf(`parse image reference "%s" error: %v`, ref, err)
	}
	opts.Server = reference.Domain(named)
	opts.Repositiory = reference.Path(named)
	if namedTaged, ok := named.(reference.NamedTagged); ok {
		opts.Tag = namedTaged.Tag()
	}
	if canonical, ok := named.(reference.Canonical); ok {
		opts.Digest = canonical.Digest()
	}

	return nil
}
