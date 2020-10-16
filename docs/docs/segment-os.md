---
id: os
title: os
sidebar_label: OS
---

## What

Display OS specific info. Defaults to Icon.

## Sample Configuration

```json
{
  "type": "os",
  "style": "plain",
  "foreground": "#26C6DA",
  "background": "#546E7A",
  "properties": {
    "postfix": " \uE0B1",
    "macos": "mac"
  }
}
```

## Properties

- macos: `string` - the string to use for macOS - defaults to macOS icon - defaults to `\uF179`
- linux: `string` - the icon to use for Linux - defaults to Linux icon - defaults to `\uF17C`
- windows: `string` - the icon to use for Windows - defaults to Windows icon - defaults to `\uE62A`
