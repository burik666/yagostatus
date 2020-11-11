# Yagostatus plugins

## Overview

Yagostatus supports [go plugins](https://golang.org/pkg/plugin/).
Plugins can be used to add or replace existing widgets.

To load a plugin, you need to specify it in the config file.
`yagostatus.yml`:
```yaml
plugins:
  path: /path/to/plugins
  load:
    - plugin: example.so
      param: value
```
- `path` - Directory where the `.so` file are located (default: current working directory).
- `plugin` - Plugin file (you can specify an absolute path).
- Plugins can have parameters.

## Example

See [example](example)

## Builtin

Plugins can be embedded in the yagostatus binary file.

    go get -tags plugin_example github.com/burik666/yagostatus

See [example_builtin.go](example_builtin.go), [example/builtin.go](example/builtin.go)

