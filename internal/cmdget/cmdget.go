// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmdget contains the get command
package cmdget

import (
	"context"
	"fmt"
	"os"
	"strings"

	docs "github.com/GoogleContainerTools/kpt/internal/docs/generated/pkgdocs"
	"github.com/GoogleContainerTools/kpt/internal/errors"
	"github.com/GoogleContainerTools/kpt/internal/pkg"
	"github.com/GoogleContainerTools/kpt/internal/types"
	"github.com/GoogleContainerTools/kpt/internal/util/argutil"
	"github.com/GoogleContainerTools/kpt/internal/util/cmdutil"
	"github.com/GoogleContainerTools/kpt/internal/util/get"
	"github.com/GoogleContainerTools/kpt/internal/util/parse"
	kptfilev1 "github.com/GoogleContainerTools/kpt/pkg/api/kptfile/v1"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner
func NewRunner(ctx context.Context, parent string) *Runner {
	r := &Runner{
		ctx: ctx,
	}
	c := &cobra.Command{
		Use:        "get {REPO_URI[.git]/PKG_PATH[@VERSION]|IMAGE:TAG} [LOCAL_DEST_DIRECTORY]",
		Args:       cobra.MinimumNArgs(1),
		Short:      docs.GetShort,
		Long:       docs.GetShort + "\n" + docs.GetLong,
		Example:    docs.GetExamples,
		RunE:       r.runE,
		PreRunE:    r.preRunE,
		SuggestFor: []string{"clone", "cp", "fetch"},
	}
	cmdutil.FixDocs("kpt", parent, c)
	r.Command = c
	c.Flags().StringVar(&r.strategy, "strategy", string(kptfilev1.ResourceMerge),
		"update strategy that should be used when updating this package -- must be one of: "+
			strings.Join(kptfilev1.UpdateStrategiesAsStrings(), ","))
	return r
}

func NewCommand(ctx context.Context, parent string) *cobra.Command {
	return NewRunner(ctx, parent).Command
}

// Runner contains the run function
type Runner struct {
	ctx      context.Context
	Get      get.Command
	Command  *cobra.Command
	strategy string
}

func (r *Runner) preRunE(_ *cobra.Command, args []string) error {
	const op errors.Op = "cmdget.preRunE"
	if len(args) == 1 {
		args = append(args, pkg.CurDir)
	} else {
		_, err := os.Lstat(args[1])
		if err == nil || os.IsExist(err) {
			resolvedPath, err := argutil.ResolveSymlink(r.ctx, args[1])
			if err != nil {
				return errors.E(op, err)
			}
			args[1] = resolvedPath
		}
	}
	destination, err := r.parseArgs(args)
	if err != nil {
		return err
	}

	p, err := pkg.New(destination)
	if err != nil {
		return errors.E(op, types.UniquePath(destination), err)
	}
	r.Get.Destination = string(p.UniquePath)

	strategy, err := kptfilev1.ToUpdateStrategy(r.strategy)
	if err != nil {
		return err
	}
	r.Get.UpdateStrategy = strategy
	return nil
}

func (r *Runner) parseArgs(args []string) (string, error) {
	const op errors.Op = "cmdget.preRunE"

	t1, err1 := parse.GitParseArgs(r.ctx, args)
	if err1 == nil {
		r.Get.Git = &t1.Git
		return t1.Destination, nil
	}

	t2, err2 := parse.OciParseArgs(r.ctx, args)
	if err2 == nil {
		r.Get.Oci = &t2.Oci
		return t2.Destination, nil
	}

	return "", errors.E(op, fmt.Errorf("%v %v", err1, err2))
}

func (r *Runner) runE(c *cobra.Command, _ []string) error {
	const op errors.Op = "cmdget.runE"
	if err := r.Get.Run(r.ctx); err != nil {
		return errors.E(op, types.UniquePath(r.Get.Destination), err)
	}

	return nil
}
