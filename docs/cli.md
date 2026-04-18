
## `score-helm`

- `--version`|`-v`: version for `score-helm`

## `score-helm init`

Initialize the local state directory and sample Score file.

- `--file`|`-f` - The score file to initialize (default `score.yaml`).
- `--no-sample` - Disable generation of the sample score file.

## `score-helm generate`

Run the conversion from Score file to output manifests.

- `--image`|`-i` - An optional container image to use for any container with image == '.'.
- `--output`|`-o` - The output manifests file to write the manifests to (default `value.yaml`).
- `--override-property` - An optional set of path=key overrides to set or remove.
- `--overrides-file` - `An optional file of Score overrides to merge in.

## `score-helm version`

Show the version for score-k8s and new version to update if available.

- `--no-logo` - Do not show the Score logo.
- `--no-updates-check` - Do not check for a new version.

## `score-helm check-version`

Assert that the version of score-k8s matches the required constraint.
