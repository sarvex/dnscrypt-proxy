// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocommand

import (
	"context"
	"fmt"
	"strings"
)

// GoVersion reports the minor version number of the highest release
// tag built into the go command on the PATH.
//
// Note that this may be higher than the version of the go tool used
// to build this application, and thus the versions of the standard
// go/{scanner,parser,ast,types} packages that are linked into it.
// In that case, callers should either downgrade to the version of
// go used to build the application, or report an error that the
// application is too old to use the go command on the PATH.
func GoVersion(ctx context.Context, inv Invocation, r *Runner) (int, error) {
	inv.Verb = "list"
	inv.Args = []string{"-e", "-f", `{{context.ReleaseTags}}`, `--`, `unsafe`}
	inv.Env = append(append([]string{}, inv.Env...), "GO111MODULE=off")
	// Unset any unneeded flags, and remove them from BuildFlags, if they're
	// present.
	inv.ModFile = ""
	inv.ModFlag = ""
	var buildFlags []string
	for _, flag := range inv.BuildFlags {
		// Flags can be prefixed by one or two dashes.
		f := strings.TrimPrefix(strings.TrimPrefix(flag, "-"), "-")
		if strings.HasPrefix(f, "mod=") || strings.HasPrefix(f, "modfile=") {
			continue
		}
		buildFlags = append(buildFlags, flag)
	}
	inv.BuildFlags = buildFlags
	stdoutBytes, err := r.Run(ctx, inv)
	if err != nil {
		return 0, err
	}
	stdout := stdoutBytes.String()
	if len(stdout) < 3 {
		return 0, fmt.Errorf("bad ReleaseTags output: %q", stdout)
	}
	// Split up "[go1.1 go1.15]" and return highest go1.X value.
	tags := strings.Fields(stdout[1 : len(stdout)-2])
	for i := len(tags) - 1; i >= 0; i-- {
		var version int
		if _, err := fmt.Sscanf(tags[i], "go1.%d", &version); err != nil {
			continue
		}
		return version, nil
	}
	return 0, fmt.Errorf("no parseable ReleaseTags in %v", tags)
}
