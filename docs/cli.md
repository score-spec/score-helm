
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