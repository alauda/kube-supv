package registry

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func createServer() *httptest.Server {
	const (
		manifestPrefix        = "/v2/test/test/manifests/"
		blobPrefix            = "/v2/test/test/blobs/"
		blobMediaType         = "application/octet-stream"
		manifestMediaType     = "application/vnd.docker.distribution.manifest.v2+json"
		manifestListMediaType = "application/vnd.docker.distribution.manifest.list.v2+json"
		configMediaType       = "application/vnd.docker.container.image.v1+json"
		layerMediaType        = "application/vnd.docker.image.rootfs.diff.tar.gzip"
	)

	layer, _ := base64.StdEncoding.DecodeString(`H4sIAAAAAAAA/+yXzXKbMBDHOesptum5sELYzPAGvfTQQ+9yLI8ZC+SRNm6ZNO/ekZQUm+mEE6SZ6HdhtF/I4L92kQdS9mvvSGqdu2O2BIiI26oKV0ScXhFLnvFKbHFT8RKrDHmFAjPARXYz4cGRtBmiNYZei5vzT3/cO+Hzp2LX9sVOuiNj55979tYbSqzKTh2MVcseAPP6x1H/vPb653WV9L8GN/rXLsn/g9HJtm8PylE+yE4vc485/W9E+Vf/fMMz5KLmddL/GvSyUw2QcsQuyrrW9A1ceI45skOrlWsYwBeg4awa2LeWAQDslaMGCp9U7FuLVyE+J8Q4ex/L5vSL/plVXHlf0kl1Zy1pWsL/NV+pEdwXqR/idk9q4A2ENY/LsgnZJzWIZ4dgR2NOIfymA8ZAd2/bMzUwbY4M4Hpevg2eTNLv5CR9eQlL3mO2//Ox/1dbr/9ys036X4U13n/i/4WWbf2BOf2X4/wv6lJ4/ddl+v5fhdj/Hx8h/yY7BU9P4xjgjT/iwtv9A4jG78aQt8RGG8JC88294dkhpg7hHXGkgN8MvNcdldZwp93dWPOtn0cikUh8FP4EAAD//z71SRkAGAAA`)
	arm64Config := []byte(`{"architecture":"arm64","config":{"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"WorkingDir":"/","OnBuild":null},"created":"2023-01-14T16:02:06.427380463Z","history":[{"created":"2023-01-14T16:02:06.427380463Z","created_by":"COPY package /package/ / # buildkit","comment":"buildkit.dockerfile.v0"}],"moby.buildkit.buildinfo.v1":"eyJmcm9udGVuZCI6ImRvY2tlcmZpbGUudjAifQ==","os":"linux","rootfs":{"type":"layers","diff_ids":["sha256:bbe544226ffb4c61fab33fb496e0415699fa0d609b31ee8bfb1f67de40838d49"]}}`)
	amd64Config := []byte(`{"architecture":"amd64","config":{"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"WorkingDir":"/","OnBuild":null},"created":"2023-01-14T16:02:06.427380463Z","history":[{"created":"2023-01-14T16:02:06.427380463Z","created_by":"COPY package /package/ / # buildkit","comment":"buildkit.dockerfile.v0"}],"moby.buildkit.buildinfo.v1":"eyJmcm9udGVuZCI6ImRvY2tlcmZpbGUudjAifQ==","os":"linux","rootfs":{"type":"layers","diff_ids":["sha256:bbe544226ffb4c61fab33fb496e0415699fa0d609b31ee8bfb1f67de40838d49"]}}`)

	arm64Manifest := []byte(fmt.Sprintf(`{
  "mediaType": "%s",
  "schemaVersion": 2,
  "config": {
    "mediaType": "%s",
    "digest": "%s",
    "size": %d
  },
  "layers": [
    {
      "mediaType": "%s",
      "digest": "%s",
      "size": %d
    }
  ]
}`, manifestMediaType,
		configMediaType, sha256Hex(arm64Config), len(arm64Config),
		layerMediaType, sha256Hex(layer), len(layer)))

	amd64Manifest := []byte(fmt.Sprintf(`{
  "mediaType": "%s",
  "schemaVersion": 2,
  "config": {
    "mediaType": "%s",
    "digest": "%s",
    "size": %d
  },
  "layers": [
    {
      "mediaType": "%s",
      "digest": "%s",
      "size": %d
    }
  ]
}`, manifestMediaType,
		configMediaType, sha256Hex(amd64Config), len(amd64Config),
		layerMediaType, sha256Hex(layer), len(layer)))

	manifestList := []byte(fmt.Sprintf(`{
  "mediaType": "%s",
  "schemaVersion": 2,
  "manifests": [
    {
      "mediaType": "%s",
      "digest": "%s",
      "size": %d,
      "platform": {
        "architecture": "arm64",
        "os": "linux"
      }
    },
    {
      "mediaType": "%s",
      "digest": "%s",
        "size": %d,
        "platform": {
          "architecture": "amd64",
          "os": "linux"
        }
    }
  ]
}`, manifestListMediaType,
		manifestMediaType, sha256Hex(arm64Manifest), len(arm64Manifest),
		manifestMediaType, sha256Hex(amd64Manifest), len(amd64Manifest)))

	contents := map[string]struct {
		ContentType string
		Body        []byte
	}{
		"/v2/": {
			ContentType: "application/json; charset=utf-8",
			Body:        []byte(`{}`),
		},
		manifestPrefix + "v1": {
			ContentType: manifestListMediaType,
			Body:        manifestList,
		},
		manifestPrefix + sha256Hex(manifestList): {
			ContentType: manifestListMediaType,
			Body:        manifestList,
		},
		manifestPrefix + sha256Hex(arm64Manifest): {
			ContentType: manifestMediaType,
			Body:        arm64Manifest,
		},
		manifestPrefix + sha256Hex(amd64Manifest): {
			ContentType: manifestMediaType,
			Body:        amd64Manifest,
		},
		blobPrefix + sha256Hex(arm64Config): {
			ContentType: blobMediaType,
			Body:        arm64Config,
		},
		blobPrefix + sha256Hex(arm64Config): {
			ContentType: blobMediaType,
			Body:        arm64Config,
		},
		blobPrefix + sha256Hex(layer): {
			ContentType: blobMediaType,
			Body:        layer,
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		c, ok := contents[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", c.ContentType)
		w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
		w.Header().Set("Content-Length", strconv.Itoa(len(c.Body)))
		if r.URL.Path != "/v2/" {
			digest := sha256Hex(c.Body)
			w.Header().Set("Docker-Content-Digest", digest)
			w.Header().Set("Etag", digest)
		}

		if r.Method == http.MethodHead {
			w.WriteHeader(200)
		} else {
			w.Write(c.Body)
		}
	}

	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestPullImageLayerToLocal(t *testing.T) {
	server := createServer()
	defer server.Close()

	tmpDir, err := os.MkdirTemp("", "kubesupv-image-")
	if err != nil {
		t.Errorf(`MkdirTemp error: %v`, err)
		t.FailNow()
	}
	defer os.RemoveAll(tmpDir)
	os.RemoveAll(tmpDir)

	imageRef := fmt.Sprintf(`%s/test/test:v1`, strings.TrimPrefix(server.URL, "http://"))
	if err := PullImageLayerToLocal(imageRef, tmpDir, "", ""); err != nil {
		t.Errorf(`PullImageLayerToLocal %s %s "" ""`, imageRef, tmpDir)
		t.FailNow()
	}
}

func sha256Hex(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return "sha256:" + strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}
