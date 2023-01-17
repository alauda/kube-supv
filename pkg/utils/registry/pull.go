package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/alauda/kube-supv/pkg/utils/untar"
	registryclient "github.com/distribution/distribution/registry/client"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/ocischema"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func PullImageLayerToLocal(imageRef, destDir, username, password string) error {
	opts := &Options{
		Username: username,
		Password: password,
		Ctx:      context.Background(),
	}
	if err := opts.ParseReference(imageRef); err != nil {
		return err
	}
	cli, err := NewClient(opts)
	if err != nil {
		return err
	}
	repo, err := cli.NewRepository(opts.Repositiory, PullAction)
	if err != nil {
		return err
	}

	manifestService, err := repo.Manifests(opts.Ctx)
	if err != nil {
		return err
	}

	var man distribution.Manifest
	if opts.Tag != "" {
		man, err = manifestService.Get(opts.Ctx, "", distribution.WithTag(opts.Tag), registryclient.ReturnContentDigest(&opts.Digest))
	} else {
		man, err = manifestService.Get(opts.Ctx, opts.Digest)
	}
	if err != nil {
		return err
	}

	currentPlatform := fmt.Sprintf("linux/%s", runtime.GOARCH)

	var platform string
	var blobID digest.Digest
	switch realMan := man.(type) {
	case *manifestlist.DeserializedManifestList:
		for _, ref := range realMan.Manifests {
			man, err := manifestService.Get(opts.Ctx, ref.Digest)
			if err != nil {
				return err
			}
			platform, blobID, err = getManifestInfo(opts, repo, man)
			if err != nil {
				return err
			}
			if platform == currentPlatform {
				break
			}
		}
	default:
		platform, blobID, err = getManifestInfo(opts, repo, man)
		if err != nil {
			return err
		}
	}

	if platform != currentPlatform {
		return fmt.Errorf(`image's platform "%s" is different current platform "%s"`,
			platform, currentPlatform)
	}

	if err := downloadBlob(opts, repo, blobID, destDir); err != nil {
		return err
	}

	return nil
}

func getManifestInfo(opts *Options, repo distribution.Repository, manifest distribution.Manifest) (string, digest.Digest, error) {
	var blobID digest.Digest
	switch realMan := manifest.(type) {
	case *schema1.SignedManifest:
		if len(realMan.FSLayers) > 0 {
			blobID = realMan.FSLayers[0].BlobSum
		}
		return fmt.Sprintf("linux/%s", realMan.Architecture), blobID, nil
	case *schema2.DeserializedManifest:
		if realMan.Config.MediaType == schema2.MediaTypeImageConfig || realMan.Config.MediaType == ocispec.MediaTypeImageConfig {
			image, err := getImage(opts, repo, realMan.Config.Digest)
			if err != nil {
				return "", blobID, err
			}
			if len(realMan.Layers) > 0 {
				blobID = realMan.Layers[0].Digest
			}
			return fmt.Sprintf("%s/%s", image.OS, image.Architecture), blobID, nil
		} else {
			return "", blobID, fmt.Errorf("unknown media type: %s", realMan.Config.MediaType)
		}
	case *ocischema.DeserializedManifest:
		if realMan.Config.MediaType == schema2.MediaTypeImageConfig || realMan.Config.MediaType == ocispec.MediaTypeImageConfig {
			image, err := getImage(opts, repo, realMan.Config.Digest)
			if err != nil {
				return "", blobID, err
			}
			if len(realMan.Layers) > 0 {
				blobID = realMan.Layers[0].Digest
			}
			return fmt.Sprintf("%s/%s", image.OS, image.Architecture), blobID, nil
		} else {
			return "", blobID, fmt.Errorf("unknown media type: %s", realMan.Config.MediaType)
		}
	}
	return "", blobID, fmt.Errorf("unknown digest type: %v", reflect.TypeOf(manifest))
}

func getImage(opts *Options, repo distribution.Repository, dgst digest.Digest) (*ocispec.Image, error) {
	config, err := repo.Blobs(opts.Ctx).Get(opts.Ctx, dgst)
	if err != nil {
		return nil, err
	}
	image := ocispec.Image{}
	if err := json.Unmarshal(config, &image); err != nil {
		return nil, err
	}
	return &image, nil
}

func downloadBlob(opts *Options, repo distribution.Repository, blobID digest.Digest, destDir string) error {
	reader, err := repo.Blobs(opts.Ctx).Open(opts.Ctx, blobID)
	if err != nil {
		return err
	}
	defer reader.Close()

	destDir = filepath.FromSlash(destDir)
	destDir, err = filepath.Abs(destDir)
	if err != nil {
		return err
	}
	if err := utils.MakeDir(destDir); err != nil {
		return err
	}
	if err := untar.Untar(reader, destDir); err != nil {
		return err
	}
	return nil
}
