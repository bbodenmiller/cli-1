package volume

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/docker/cli/internal/test"
	"github.com/docker/docker/api/types/volume"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestVolumeCreateErrors(t *testing.T) {
	testCases := []struct {
		args             []string
		flags            map[string]string
		volumeCreateFunc func(volume.CreateOptions) (volume.Volume, error)
		expectedError    string
	}{
		{
			args: []string{"volumeName"},
			flags: map[string]string{
				"name": "volumeName",
			},
			expectedError: "conflicting options: either specify --name or provide positional arg, not both",
		},
		{
			args:          []string{"too", "many"},
			expectedError: "requires at most 1 argument",
		},
		{
			volumeCreateFunc: func(createBody volume.CreateOptions) (volume.Volume, error) {
				return volume.Volume{}, errors.Errorf("error creating volume")
			},
			expectedError: "error creating volume",
		},
	}
	for _, tc := range testCases {
		cmd := newCreateCommand(
			test.NewFakeCli(&fakeClient{
				volumeCreateFunc: tc.volumeCreateFunc,
			}),
		)
		cmd.SetArgs(tc.args)
		for key, value := range tc.flags {
			cmd.Flags().Set(key, value)
		}
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestVolumeCreateWithName(t *testing.T) {
	name := "foo"
	cli := test.NewFakeCli(&fakeClient{
		volumeCreateFunc: func(body volume.CreateOptions) (volume.Volume, error) {
			if body.Name != name {
				return volume.Volume{}, errors.Errorf("expected name %q, got %q", name, body.Name)
			}
			return volume.Volume{
				Name: body.Name,
			}, nil
		},
	})

	buf := cli.OutBuffer()

	// Test by flags
	cmd := newCreateCommand(cli)
	cmd.Flags().Set("name", name)
	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.Equal(name, strings.TrimSpace(buf.String())))

	// Then by args
	buf.Reset()
	cmd = newCreateCommand(cli)
	cmd.SetArgs([]string{name})
	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.Equal(name, strings.TrimSpace(buf.String())))
}

func TestVolumeCreateWithFlags(t *testing.T) {
	expectedDriver := "foo"
	expectedOpts := map[string]string{
		"bar": "1",
		"baz": "baz",
	}
	expectedLabels := map[string]string{
		"lbl1": "v1",
		"lbl2": "v2",
	}
	name := "banana"

	cli := test.NewFakeCli(&fakeClient{
		volumeCreateFunc: func(body volume.CreateOptions) (volume.Volume, error) {
			if body.Name != "" {
				return volume.Volume{}, errors.Errorf("expected empty name, got %q", body.Name)
			}
			if body.Driver != expectedDriver {
				return volume.Volume{}, errors.Errorf("expected driver %q, got %q", expectedDriver, body.Driver)
			}
			if !reflect.DeepEqual(body.DriverOpts, expectedOpts) {
				return volume.Volume{}, errors.Errorf("expected drivers opts %v, got %v", expectedOpts, body.DriverOpts)
			}
			if !reflect.DeepEqual(body.Labels, expectedLabels) {
				return volume.Volume{}, errors.Errorf("expected labels %v, got %v", expectedLabels, body.Labels)
			}
			return volume.Volume{
				Name: name,
			}, nil
		},
	})

	cmd := newCreateCommand(cli)
	cmd.Flags().Set("driver", "foo")
	cmd.Flags().Set("opt", "bar=1")
	cmd.Flags().Set("opt", "baz=baz")
	cmd.Flags().Set("label", "lbl1=v1")
	cmd.Flags().Set("label", "lbl2=v2")
	assert.NilError(t, cmd.Execute())
	assert.Check(t, is.Equal(name, strings.TrimSpace(cli.OutBuffer().String())))
}
