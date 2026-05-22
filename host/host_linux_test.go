//go:build linux
// +build linux

package host

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/shirou/gopsutil/v3/common"
)

func TestGetRedhatishVersion(t *testing.T) {
	var ret string
	c := []string{"Rawhide"}
	ret = getRedhatishVersion(c)
	if ret != "rawhide" {
		t.Errorf("Could not get version rawhide: %v", ret)
	}

	c = []string{"Fedora release 15 (Lovelock)"}
	ret = getRedhatishVersion(c)
	if ret != "15" {
		t.Errorf("Could not get version fedora: %v", ret)
	}

	c = []string{"Enterprise Linux Server release 5.5 (Carthage)"}
	ret = getRedhatishVersion(c)
	if ret != "5.5" {
		t.Errorf("Could not get version redhat enterprise: %v", ret)
	}

	c = []string{""}
	ret = getRedhatishVersion(c)
	if ret != "" {
		t.Errorf("Could not get version with no value: %v", ret)
	}
}

func TestGetRedhatishPlatform(t *testing.T) {
	var ret string
	c := []string{"red hat"}
	ret = getRedhatishPlatform(c)
	if ret != "redhat" {
		t.Errorf("Could not get platform redhat: %v", ret)
	}

	c = []string{"Fedora release 15 (Lovelock)"}
	ret = getRedhatishPlatform(c)
	if ret != "fedora" {
		t.Errorf("Could not get platform fedora: %v", ret)
	}

	c = []string{"Enterprise Linux Server release 5.5 (Carthage)"}
	ret = getRedhatishPlatform(c)
	if ret != "enterprise" {
		t.Errorf("Could not get platform redhat enterprise: %v", ret)
	}

	c = []string{""}
	ret = getRedhatishPlatform(c)
	if ret != "" {
		t.Errorf("Could not get platform with no value: %v", ret)
	}
}

func TestGetKylinVersion(t *testing.T) {
	cases := []struct {
		name     string
		contents []string
		want     string
	}{
		{
			name:     "skip host in release",
			contents: []string{"Kylin Linux Advanced Server Host release Host V10 (Kivity)"},
			want:     "v10 (kivity)",
		},
		{
			name:     "skip host without release",
			contents: []string{"Kylin Linux Advanced Server Host V10 (Kivity)"},
			want:     "v10 (kivity)",
		},
		{
			name:     "numeric version",
			contents: []string{"Kylin Linux release 10 (Kivity)"},
			want:     "10 (kivity)",
		},
		{
			name:     "skip arbitrary description words",
			contents: []string{"Kylin Linux Advanced Server Host Edition release Server Host V10 (Kivity)"},
			want:     "v10 (kivity)",
		},
		{
			name:     "numeric dotted version",
			contents: []string{"Kylin Linux Advanced Server release 10.1 SP1"},
			want:     "10.1",
		},
		{
			name:     "no version",
			contents: []string{"Kylin Linux Advanced Server Host"},
			want:     "",
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := getKylinVersion(tt.contents); got != tt.want {
				t.Errorf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func TestPlatformInformationKylinPreferOSReleaseVersionID(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "os-release"), []byte(`NAME="Kylin Linux Advanced Server Host"
VERSION="V10 (Kivity)"
ID="kylin"
VERSION_ID="V10"
PRETTY_NAME="Kylin Linux Advanced Server Host V10 (Kivity)"
`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "system-release"), []byte("Kylin Linux Advanced Server Host release Host V10 (Kivity)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.WithValue(context.Background(),
		common.EnvKey,
		common.EnvMap{common.HostEtcEnvKey: root},
	)

	platform, _, version, err := PlatformInformationWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if platform != "kylin" {
		t.Errorf("platform: want %q, got %q", "kylin", platform)
	}
	if version != "V10" {
		t.Errorf("version: want %q, got %q", "V10", version)
	}
}

func TestPlatformInformationKylinFallbackReleaseFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "os-release"), []byte(`ID="kylin"`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "system-release"), []byte("Kylin Linux Advanced Server Host release Host V10 (Kivity)\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.WithValue(context.Background(),
		common.EnvKey,
		common.EnvMap{common.HostEtcEnvKey: root},
	)

	_, _, version, err := PlatformInformationWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != "v10 (kivity)" {
		t.Errorf("version: want %q, got %q", "v10 (kivity)", version)
	}
}

func Test_getlsbStruct(t *testing.T) {
	cases := []struct {
		root        string
		id          string
		release     string
		codename    string
		description string
	}{
		{"arch", "Arch", "rolling", "", "Arch Linux"},
		{"ubuntu_22_04", "Ubuntu", "22.04", "jammy", "Ubuntu 22.04.2 LTS"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.root, func(t *testing.T) {
			ctx := context.WithValue(context.Background(),
				common.EnvKey,
				common.EnvMap{common.HostEtcEnvKey: "./testdata/linux/lsbStruct/" + tt.root},
			)

			v, err := getlsbStruct(ctx)
			if err != nil {
				t.Errorf("error %v", err)
			}
			if v.ID != tt.id {
				t.Errorf("ID: want %v, got %v", tt.id, v.ID)
			}
			if v.Release != tt.release {
				t.Errorf("Release: want %v, got %v", tt.release, v.Release)
			}
			if v.Codename != tt.codename {
				t.Errorf("Codename: want %v, got %v", tt.codename, v.Codename)
			}
			if v.Description != tt.description {
				t.Errorf("Description: want %v, got %v", tt.description, v.Description)
			}

			t.Log(v)
		})
	}
}
