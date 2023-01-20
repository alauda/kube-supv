package kubeclient

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	DefaultTimeout = time.Second * 60
)

var (
	EventFilter = predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			type withSpec interface {
				SpecInterface() interface{}
			}
			if !reflect.DeepEqual(e.ObjectOld.GetLabels(), e.ObjectNew.GetLabels()) {
				return true
			}

			if !reflect.DeepEqual(e.ObjectOld.GetAnnotations(), e.ObjectNew.GetAnnotations()) {
				return true
			}
			oldObj, ok1 := e.ObjectOld.(withSpec)
			newObj, ok2 := e.ObjectNew.(withSpec)
			if ok1 && ok2 {
				if !reflect.DeepEqual(oldObj, newObj) {
					return true
				}
			}
			return false
		},
	}
)

type Client struct {
	client client.Client
}

func Wrap(c client.Client) *Client {
	return &Client{client: c}
}

func NewClient(cfg *rest.Config, scheme *runtime.Scheme) (*Client, error) {
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return &Client{client: c}, nil
}

func (c *Client) Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DefaultTimeout)
}

func Ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DefaultTimeout)
}

func (c *Client) Get(key client.ObjectKey, obj client.Object) error {
	ctx, cancel := c.Context()
	defer cancel()

	return c.client.Get(ctx, key, obj)
}

func (c *Client) GetByName(namespace, name string, obj client.Object) error {
	return c.Get(types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, obj)
}

func (c *Client) List(list client.ObjectList, opts ...client.ListOption) error {
	ctx, cancel := c.Context()
	defer cancel()

	return c.client.List(ctx, list, opts...)
}

func (c *Client) ListByLabel(list client.ObjectList, key, val string) error {
	opts := &client.ListOptions{}
	client.MatchingLabels{
		key: val,
	}.ApplyToList(opts)

	return c.List(list, opts)
}

func (c *Client) ListInNamespace(list client.ObjectList, namespace string) error {
	opts := &client.ListOptions{}
	client.InNamespace(namespace).ApplyToList(opts)

	return c.List(list, opts)
}

func (c *Client) ListInNamespaceByLabel(list client.ObjectList, namespace, key, name string) error {
	opts := &client.ListOptions{}
	client.InNamespace(namespace).ApplyToList(opts)
	client.MatchingLabels{
		key: name,
	}.ApplyToList(opts)

	return c.List(list, opts)
}

func (c *Client) Create(obj client.Object, opts ...client.CreateOption) error {
	ctx, cancel := c.Context()
	defer cancel()

	return c.client.Create(ctx, obj, opts...)
}

func (c *Client) Update(obj client.Object, opts ...client.UpdateOption) error {
	ctx, cancel := c.Context()
	defer cancel()

	return c.client.Update(ctx, obj, opts...)
}

func (c *Client) UpdateStatus(obj client.Object, opts ...client.SubResourceUpdateOption) error {

	ctx, cancel := Ctx()
	defer cancel()

	return c.client.Status().Update(ctx, obj, opts...)
}

func (c *Client) Delete(obj client.Object, opts ...client.DeleteOption) error {

	ctx, cancel := c.Context()
	defer cancel()

	err := c.client.Delete(ctx, obj, opts...)
	if apierrors.IsNotFound(err) {
		err = nil
	}
	return err
}

func (c *Client) DeleteAllOf(obj client.Object, opts ...client.DeleteAllOfOption) error {
	ctx, cancel := c.Context()
	defer cancel()

	return c.client.DeleteAllOf(ctx, obj, opts...)
}

func (c *Client) Patch(obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	ctx, cancel := Ctx()
	defer cancel()

	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c *Client) PatchJSON(obj client.Object, json string, opts ...client.PatchOption) error {
	return c.Patch(obj, client.RawPatch(types.MergePatchType, []byte(json)), opts...)
}

// PatchStatus is a wrapper for client status Patch
func (c *Client) PatchStatus(obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	ctx, cancel := Ctx()
	defer cancel()

	return c.client.Status().Patch(ctx, obj, patch, opts...)
}

func (c *Client) PatchStatusJSON(obj client.Object, json string, opts ...client.SubResourcePatchOption) error {
	return c.PatchStatus(obj, client.RawPatch(types.MergePatchType, []byte(json)), opts...)
}

func (c *Client) Label(obj client.Object, name, value string) error {
	metadata := struct {
		Metadata metav1.ObjectMeta `json:"metadata,omitempty"`
	}{
		Metadata: metav1.ObjectMeta{
			Labels: map[string]string{
				name: value,
			},
		},
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return c.PatchJSON(obj, string(data))
}

func (c *Client) RemoveLabel(obj client.Object, name string) error {
	op := []struct {
		OP   string `json:"op"`
		Path string `json:"path"`
	}{
		{
			OP:   "remove",
			Path: "/metadata/labels/" + escape(name),
		},
	}
	data, err := json.Marshal(op)
	if err != nil {
		return err
	}
	err = c.Patch(obj, client.RawPatch(types.JSONPatchType, []byte(data)))
	if apierrors.IsInvalid(err) {
		err = nil
	}
	return err
}

func (c *Client) Annotaion(obj client.Object, name, value string) error {
	metadata := struct {
		Metadata metav1.ObjectMeta `json:"metadata,omitempty"`
	}{
		Metadata: metav1.ObjectMeta{
			Annotations: map[string]string{
				name: value,
			},
		},
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return c.PatchJSON(obj, string(data))
}

func (c *Client) RemoveAnnotation(obj client.Object, name string) error {
	op := []struct {
		OP   string `json:"op"`
		Path string `json:"path"`
	}{
		{
			OP:   "remove",
			Path: "/metadata/annotations/" + escape(name),
		},
	}
	data, err := json.Marshal(op)
	if err != nil {
		return err
	}
	err = c.Patch(obj, client.RawPatch(types.JSONPatchType, []byte(data)))
	if apierrors.IsInvalid(err) {
		err = nil
	}
	return err
}

func (c *Client) EnsureFinalizer(obj client.Object, finalizer string) error {
	if obj.GetDeletionTimestamp() != nil {
		return nil
	}
	finalizers := obj.GetFinalizers()
	if funk.ContainsString(finalizers, finalizer) {
		return nil
	}
	finalizers = append(finalizers, finalizer)
	obj.SetFinalizers(finalizers)
	if err := c.Update(obj); err != nil {
		return errors.Wrapf(err, `add finalizer "%s" for "%s"`, finalizer, obj.GetName())
	}
	return nil
}

func (c *Client) FinishFinalizer(obj client.Object, finalizer string) error {
	if obj.GetDeletionTimestamp() == nil {
		return nil
	}
	finalizers := obj.GetFinalizers()
	changed := false
	finalizers = funk.FilterString(finalizers, func(s string) bool {
		if s == finalizer {
			changed = true
			return false
		}
		return true
	})
	if changed {
		obj.SetFinalizers(finalizers)
		if err := c.Update(obj); err != nil {
			return errors.Wrapf(err, `remove finalizer "%s" for "%s"`, finalizer, obj.GetName())
		}
	}
	return nil
}

func (c *Client) AddOwnerReference(obj, owner client.Object) error {
	changed := false
	found := false

	ownerName := owner.GetName()
	ownerUID := owner.GetUID()
	kind := owner.GetObjectKind().GroupVersionKind().Kind
	apieVersion := owner.GetObjectKind().GroupVersionKind().GroupVersion().String()

	ownerReferences := obj.GetOwnerReferences()

	for i := range ownerReferences {
		ref := &ownerReferences[i]
		if ref.APIVersion == apieVersion && ref.Kind == kind && ref.Name == ownerName {
			found = true
			if ref.UID != ownerUID {
				ref.UID = ownerUID
				changed = true
			}
		}
	}
	if !found {
		ownerReferences = append(ownerReferences, metav1.OwnerReference{
			APIVersion: apieVersion,
			Kind:       kind,
			Name:       ownerName,
			UID:        ownerUID,
		})
		changed = true
	}
	if changed {
		obj.SetOwnerReferences(ownerReferences)
		if err := c.Update(obj); err != nil {
			return errors.Wrapf(err, `add owner references fort "%s: %s/%s" to "%s: %s/%s"`,
				obj.GetObjectKind().GroupVersionKind().String(), obj.GetNamespace(), obj.GetName(),
				owner.GetObjectKind().GroupVersionKind().String(), owner.GetNamespace(), owner.GetName())
		}
	}
	return nil
}

func escape(key string) string {
	// see https://tools.ietf.org/html/rfc6901#section-3
	key = strings.Replace(key, "~", "~0", -1)
	key = strings.Replace(key, "/", "~1", -1)
	return key
}

func IsUnknownKind(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "no matches for kind")
}
