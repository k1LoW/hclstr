# hclstr

`hclstr` is a utility tool for string literals in HCL files.

## Usage

### `hclstr fmt [FILE ...]`

Format HCL files and string literals in HCL files.

For each string literal field, a different formatter can be specified.

```console
find . -type f -name '*.tf' | xargs -I{} hclstr fmt {} --field 'policy:prettier ? --write --parser json'
```

Any formatter can be specified for each field with the `--field` option ( `field:format command` ).

By formatting the file of placeholder `?` or the `FILE` environment variable, it can format string literals.


```console
--field 'policy:prettier ? --write --parser json'
```

or

```console
--field 'policy:prettier ${FILE} --write --parser json'
```

## Install

**homebrew tap:**

```console
$ brew install k1LoW/tap/hclstr
```

**go install:**

```console
$ go install github.com/k1LoW/hclstr@latest
```

**manually:**

Download binary from [releases page](https://github.com/k1LoW/hclstr/releases)
