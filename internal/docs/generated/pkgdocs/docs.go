// Code generated by "mdtogo"; DO NOT EDIT.
package pkgdocs

var PkgShort = `Get, update, and describe packages with resources`
var PkgLong = `
The ` + "`" + `pkg` + "`" + ` command group contains subcommands for fetching, updating and describing ` + "`" + `kpt` + "`" + ` packages
from git repositories.
`

var CatShort = `Print the resources in a file/directory`
var CatLong = `
  kpt pkg cat [FILE | DIR]

Args:

  FILE | DIR:
    Path to a directory either a directory containing files with KRM resources, or
    a file with KRM resource(s). Defaults to the current directory.
`
var CatExamples = `
  # Print resource from a file.
  $ kpt pkg cat path/to/deployment.yaml

  # Print resources from current directory.
  $ kpt pkg cat
`

var DiffShort = `Show differences between a local package and upstream.`
var DiffLong = `
  kpt pkg diff [PKG_PATH@VERSION] [flags]

Args:

  PKG_PATH:
    Local package path to compare. diff will fail if the directory doesn't exist, or does not
    contain a Kptfile. Defaults to the current working directory.
  
  VERSION:
    A git tag, branch, or commit. Specified after the local_package with @, for
    example my-package@master.
    Defaults to the local package version that was last fetched.

Flags:

  --diff-type:
    The type of changes to view (local by default). Following types are
    supported:
  
    local: Shows changes in local package relative to upstream source package
           at original version.
    remote: Shows changes in upstream source package at target version
            relative to original version.
    combined: Shows changes in local package relative to upstream source
              package at target version.
    3way: Shows changes in local package and source package at target version
          relative to original version side by side.
  
  --diff-tool:
    Command line diffing tool ('diff' by default) for showing the changes.
    Note that it overrides the KPT_EXTERNAL_DIFF environment variable.
  
    # Show changes using 'meld' commandline tool.
    kpt pkg diff @master --diff-tool meld
  
  --diff-tool-opts:
    Commandline options to use with the command line diffing tool.
    Note that it overrides the KPT_EXTERNAL_DIFF_OPTS environment variable.
  
    # Show changes using the diff command with recursive options.
    kpt pkg diff @master --diff-tool meld --diff-tool-opts "-r"

Environment Variables:

  KPT_EXTERNAL_DIFF:
    Commandline diffing tool ('diff; by default) that will be used to show
    changes.
  
    # Use meld to show changes
    KPT_EXTERNAL_DIFF=meld kpt pkg diff
  
  KPT_EXTERNAL_DIFF_OPTS:
    Commandline options to use for the diffing tool. For ex.
    # Using "-a" diff option
    KPT_EXTERNAL_DIFF_OPTS="-a" kpt pkg diff --diff-tool meld
  
  KPT_CACHE_DIR:
    Controls where to cache remote packages when fetching them.
    Defaults to <HOME>/.kpt/repos/
    On macOS and Linux <HOME> is determined by the $HOME env variable, while on
    Windows it is given by the %USERPROFILE% env variable.
`
var DiffExamples = `

  # Show changes in current package relative to upstream source package.
  $ kpt pkg diff
`

var GetShort = `Fetch a package from a git repo.`
var GetLong = `
  kpt pkg get {REPO_URI[.git]/PKG_PATH[@VERSION]|IMAGE:TAG} [LOCAL_DEST_DIRECTORY] [flags]

Args:

  REPO_URI:
    URI of a git repository containing 1 or more packages as subdirectories.
    In most cases the .git suffix should be specified to delimit the REPO_URI
    from the PKG_PATH, but this is not required for widely recognized repo
    prefixes. If get cannot parse the repo for the directory and version,
    then it will print an error asking for '.git' to be specified as part of
    the argument.
  
  PKG_PATH:
    Path to remote subdirectory containing Kubernetes resource configuration
    files or directories. Defaults to the root directory.
    Uses '/' as the path separator (regardless of OS).
    e.g. staging/cockroachdb
  
  VERSION:
    A git tag, branch, ref or commit for the remote version of the package
    to fetch. Defaults to the default branch of the repository.
  
  IMAGE:
    Reference to an OCI image containing a package in the root directory.
  
  TAG:
    An image tag or @sha256 digest for the remote version of the image
    to fetch. Defaults to the 'latest' tag on the image.
    
  LOCAL_DEST_DIRECTORY:
    The local directory to write the package to. Defaults to a subdirectory of the
    current working directory named after the upstream package.

Flags:

  --strategy:
    Defines which strategy should be used to update the package. It defaults to
    'resource-merge'.
  
      * resource-merge: Perform a structural comparison of the original /
        updated resources, and merge the changes into the local package.
      * fast-forward: Fail without updating if the local package was modified
        since it was fetched.
      * force-delete-replace: Wipe all the local changes to the package and replace
        it with the remote version.

Env Vars:

  KPT_CACHE_DIR:
    Controls where to cache remote packages when fetching them.
    Defaults to <HOME>/.kpt/repos/
    On macOS and Linux <HOME> is determined by the $HOME env variable, while on
    Windows it is given by the %USERPROFILE% env variable.
`
var GetExamples = `

  # Fetch package cockroachdb from github.com/kubernetes/examples/staging/cockroachdb
  # This creates a new subdirectory 'cockroachdb' for the downloaded package.
  $ kpt pkg get https://github.com/kubernetes/examples.git/staging/cockroachdb@master


  # Fetch package cockroachdb from github.com/kubernetes/examples/staging/cockroachdb
  # This will create a new directory 'my-package' for the downloaded package if it
  # doesn't already exist.
  $ kpt pkg get https://github.com/kubernetes/examples.git/staging/cockroachdb@master ./my-package/


  # Fetch package examples from github.com/kubernetes/examples at the specified
  # git hash.
  # This will create a new directory 'examples' for the package.
  $ kpt pkg get https://github.com/kubernetes/examples.git/@6fe2792
`

var InitShort = `Initialize an empty package.`
var InitLong = `
  kpt pkg init [DIR] [flags]

Args:

  DIR:
    init fails if DIR does not already exist. Defaults to the current working directory.

Flags:

  --description
    Short description of the package. (default "sample description")
  
  --keywords
    A list of keywords describing the package.
  
  --site
    Link to page with information about the package.
`
var InitExamples = `

  # Creates a new Kptfile with metadata in the cockroachdb directory.
  $ mkdir cockroachdb; kpt pkg init cockroachdb --keywords "cockroachdb,nosql,db"  \
      --description "my cockroachdb implementation"

  # Creates a new Kptfile without metadata in the current directory.
  $ kpt pkg init
`

var TreeShort = `Display resources, files and packages in a tree structure.`
var TreeLong = `
  kpt pkg tree [DIR]
`
var TreeExamples = `
  # Show resources in the current directory.
  $ kpt pkg tree
`

var UpdateShort = `Apply upstream package updates.`
var UpdateLong = `
  kpt pkg update [PKG_PATH][@VERSION] [flags]

Args:

  PKG_PATH:
    Local package path to update. Directory must exist and contain a Kptfile
    to be updated. Defaults to the current working directory.
  
  VERSION:
    A git tag, branch, ref or commit. Specified after the local_package
    with @ -- pkg@version.
    Defaults the ref specified in the Upstream section of the package Kptfile.
  
    Version types:
      * branch: update the local contents to the tip of the remote branch
      * tag: update the local contents to the remote tag
      * commit: update the local contents to the remote commit

Flags:

  --strategy:
    Defines which strategy should be used to update the package. This will change
    the update strategy for the current kpt package for the current and future
    updates. If a strategy is not provided, the strategy specified in the package
    Kptfile will be used.
  
      * resource-merge: Perform a structural comparison of the original /
        updated resources, and merge the changes into the local package.
      * fast-forward: Fail without updating if the local package was modified
        since it was fetched.
      * force-delete-replace: Wipe all the local changes to the package and replace
        it with the remote version.

Env Vars:

  KPT_CACHE_DIR:
    Controls where to cache remote packages when fetching them.
    Defaults to <HOME>/.kpt/repos/
    On macOS and Linux <HOME> is determined by the $HOME env variable, while on
    Windows it is given by the %USERPROFILE% env variable.
`
var UpdateExamples = `
  # Update package in the current directory.
  # git add . && git commit -m 'some message'
  $ kpt pkg update

  # Update my-package-dir/ to match the v1.3 branch or tag.
  # git add . && git commit -m 'some message'
  $ kpt pkg update my-package-dir/@v1.3

  # Update with the fast-forward strategy.
  # git add . && git commit -m "some message"
  $ kpt pkg update my-package-dir/@master --strategy fast-forward
`
