package accesstoken

import (
	"errors"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	oapi "github.com/openshift/origin/pkg/oauth/api"
	"github.com/openshift/origin/pkg/oauth/registry/test"
)

func TestCreateValidationError(t *testing.T) {
	registry := test.AccessTokenRegistry{}
	storage := REST{
		registry: &registry,
	}
	accessToken := &oapi.AccessToken{
		ObjectMeta: api.ObjectMeta{Name: "accessTokenName"},
		// ClientName: "clientName", // Missing required field
		UserName: "userName",
		UserUID:  "userUID",
	}

	ctx := api.NewContext()
	_, err := storage.Create(ctx, accessToken)
	if err == nil {
		t.Errorf("Expected validation error")
	}
}

func TestCreateStorageError(t *testing.T) {
	registry := test.AccessTokenRegistry{
		Err: errors.New("Sample Error"),
	}
	storage := REST{
		registry: &registry,
	}
	accessToken := &oapi.AccessToken{
		ObjectMeta: api.ObjectMeta{Name: "accessTokenName"},
		ClientName: "clientName",
		UserName:   "userName",
		UserUID:    "userUID",
	}

	ctx := api.NewContext()
	channel, err := storage.Create(ctx, accessToken)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case r := <-channel:
		switch r := r.Object.(type) {
		case *api.Status:
			if r.Message == registry.Err.Error() {
				// expected case
			} else {
				t.Errorf("Got back unexpected error: %#v", r)
			}
		default:
			t.Errorf("Got back non-status result: %v", r)
		}
	case <-time.After(time.Millisecond * 100):
		t.Error("Unexpected timeout from async channel")
	}
}

func TestCreateValid(t *testing.T) {
	registry := test.AccessTokenRegistry{}
	storage := REST{
		registry: &registry,
	}
	accessToken := &oapi.AccessToken{
		ObjectMeta: api.ObjectMeta{Name: "accessTokenName"},
		ClientName: "clientName",
		UserName:   "userName",
		UserUID:    "userUID",
	}

	ctx := api.NewContext()
	channel, err := storage.Create(ctx, accessToken)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case r := <-channel:
		switch r := r.Object.(type) {
		case *api.Status:
			t.Errorf("Got back unexpected status: %#v", r)
		case *oapi.AccessToken:
		// expected case
		default:
			t.Errorf("Got unexpected type: %#v", r)
		}
	case <-time.After(time.Millisecond * 100):
		t.Error("Unexpected timeout from async channel")
	}
}

func TestGetError(t *testing.T) {
	registry := test.AccessTokenRegistry{
		Err: errors.New("Sample Error"),
	}
	storage := REST{
		registry: &registry,
	}
	ctx := api.NewContext()
	_, err := storage.Get(ctx, "name")
	if err == nil {
		t.Errorf("expected error")
		return
	}
	if err != registry.Err {
		t.Errorf("got unexpected error: %v", err)
		return
	}
}

func TestGetValid(t *testing.T) {
	registry := test.AccessTokenRegistry{
		AccessToken: &oapi.AccessToken{
			ObjectMeta: api.ObjectMeta{Name: "accessTokenName"},
		},
	}
	storage := REST{
		registry: &registry,
	}
	ctx := api.NewContext()
	token, err := storage.Get(ctx, "name")
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
		return
	}
	if token != registry.AccessToken {
		t.Errorf("got unexpected token: %v", token)
		return
	}
}

func TestListError(t *testing.T) {
	registry := test.AccessTokenRegistry{
		Err: errors.New("Sample Error"),
	}
	storage := REST{
		registry: &registry,
	}
	ctx := api.NewContext()
	_, err := storage.List(ctx, labels.Everything(), labels.Everything())
	if err == nil {
		t.Errorf("expected error")
		return
	}
	if err != registry.Err {
		t.Errorf("got unexpected error: %v", err)
		return
	}
}

func TestListEmpty(t *testing.T) {
	registry := test.AccessTokenRegistry{
		AccessTokens: &oapi.AccessTokenList{},
	}
	storage := REST{
		registry: &registry,
	}
	ctx := api.NewContext()
	tokens, err := storage.List(ctx, labels.Everything(), labels.Everything())
	if err != registry.Err {
		t.Errorf("got unexpected error: %v", err)
		return
	}
	switch tokens := tokens.(type) {
	case *oapi.AccessTokenList:
		if len(tokens.Items) != 0 {
			t.Errorf("expected empty list, got %#v", tokens)
		}
	default:
		t.Errorf("expected AccessTokenList, got: %v", tokens)
		return
	}
}

func TestList(t *testing.T) {
	registry := test.AccessTokenRegistry{
		AccessTokens: &oapi.AccessTokenList{
			Items: []oapi.AccessToken{
				{},
				{},
			},
		},
	}
	storage := REST{
		registry: &registry,
	}
	ctx := api.NewContext()
	tokens, err := storage.List(ctx, labels.Everything(), labels.Everything())
	if err != registry.Err {
		t.Errorf("got unexpected error: %v", err)
		return
	}
	switch tokens := tokens.(type) {
	case *oapi.AccessTokenList:
		if len(tokens.Items) != 2 {
			t.Errorf("expected list with 2 items, got %#v", tokens)
		}
	default:
		t.Errorf("expected AccessTokenList, got: %v", tokens)
		return
	}
}

func TestUpdateNotSupported(t *testing.T) {
	registry := test.AccessTokenRegistry{
		Err: errors.New("Storage Error"),
	}
	storage := REST{
		registry: &registry,
	}
	accessToken := &oapi.AccessToken{
		ObjectMeta: api.ObjectMeta{Name: "accessTokenName"},
	}

	ctx := api.NewContext()
	_, err := storage.Update(ctx, accessToken)
	if err == nil {
		t.Errorf("expected unsupported error, but update succeeded")
		return
	}
	if err == registry.Err {
		t.Errorf("expected unsupported error, but registry was called")
		return
	}
}

func TestDeleteError(t *testing.T) {
	registry := test.AccessTokenRegistry{
		Err: errors.New("Sample Error"),
	}
	storage := REST{
		registry: &registry,
	}

	ctx := api.NewContext()
	channel, err := storage.Delete(ctx, "foo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case r := <-channel:
		switch r := r.Object.(type) {
		case *api.Status:
			if r.Message == registry.Err.Error() {
				// expected case
			} else {
				t.Errorf("Got back unexpected error: %#v", r)
			}
		default:
			t.Errorf("Got back non-status result: %v", r)
		}
	case <-time.After(time.Millisecond * 100):
		t.Error("Unexpected timeout from async channel")
	}
}

func TestDeleteValid(t *testing.T) {
	registry := test.AccessTokenRegistry{}
	storage := REST{
		registry: &registry,
	}

	ctx := api.NewContext()
	channel, err := storage.Delete(ctx, "foo")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	select {
	case r := <-channel:
		switch r := r.Object.(type) {
		case *api.Status:
			if r.Status != "Success" {
				t.Errorf("Got back non-success status: %#v", r)
			}
		default:
			t.Errorf("Got back non-status result: %v", r)
		}
	case <-time.After(time.Millisecond * 100):
		t.Error("Unexpected timeout from async channel")
	}

	if registry.DeletedAccessTokenName != "foo" {
		t.Error("Unexpected access token deleted: %s", registry.DeletedAccessTokenName)
	}
}