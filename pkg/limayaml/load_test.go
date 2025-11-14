// SPDX-FileCopyrightText: Copyright The Lima Authors
// SPDX-License-Identifier: Apache-2.0

package limayaml

import (
	"fmt"
	"runtime"
	"testing"

	"gotest.tools/v3/assert"
)

func TestLoadEmpty(t *testing.T) {
	_, err := Load(t.Context(), []byte{}, "empty.yaml")
	assert.NilError(t, err)
}

func TestLoadError(t *testing.T) {
	// missing a "script:" line
	s := `
provision:
- mode: system
  script: |
    #!/bin/sh
    echo one
- mode: system
    #!/bin/sh
    echo two
- mode: system
  script: |
    #!/bin/sh
    echo three
`
	_, err := Load(t.Context(), []byte(s), "error.yaml")
	assert.ErrorContains(t, err, "map key-value is pre-defined")
}

func TestLoadDiskString(t *testing.T) {
	s := `
additionalDisks:
- name
`
	y, err := Load(t.Context(), []byte(s), "disk.yaml")
	assert.NilError(t, err)
	assert.Equal(t, len(y.AdditionalDisks), 1)
	assert.Equal(t, y.AdditionalDisks[0].Name, "name")
	assert.Assert(t, y.AdditionalDisks[0].Format == nil)
	assert.Assert(t, y.AdditionalDisks[0].FSType == nil)
	assert.Assert(t, y.AdditionalDisks[0].FSArgs == nil)
}

func TestLoadDiskStruct(t *testing.T) {
	s := `
additionalDisks:
- name: "name"
  format: false
  fsType: "xfs"
  fsArgs: ["-i","size=512"]
`
	y, err := Load(t.Context(), []byte(s), "disk.yaml")
	assert.NilError(t, err)
	assert.Assert(t, len(y.AdditionalDisks) == 1)
	assert.Equal(t, y.AdditionalDisks[0].Name, "name")
	assert.Assert(t, y.AdditionalDisks[0].Format != nil)
	assert.Equal(t, *y.AdditionalDisks[0].Format, false)
	assert.Assert(t, y.AdditionalDisks[0].FSType != nil)
	assert.Equal(t, *y.AdditionalDisks[0].FSType, "xfs")
	assert.Assert(t, len(y.AdditionalDisks[0].FSArgs) == 2)
	assert.Equal(t, y.AdditionalDisks[0].FSArgs[0], "-i")
	assert.Equal(t, y.AdditionalDisks[0].FSArgs[1], "size=512")
}

func TestLoadCPUsInteger(t *testing.T) {
	s := `
cpus: 4
`
	y, err := Load(t.Context(), []byte(s), "cpus.yaml")
	assert.NilError(t, err)
	assert.Assert(t, y.CPUs != nil)
	assert.Equal(t, *y.CPUs, fmt.Sprintf("%d", 4))
}

func TestLoadCPUsString(t *testing.T) {
	s := `
cpus: host
`
	y, err := Load(t.Context(), []byte(s), "cpus.yaml")
	assert.NilError(t, err)
	assert.Assert(t, y.CPUs != nil)
	assert.Equal(t, *y.CPUs, fmt.Sprintf("%d", runtime.NumCPU()))
}
