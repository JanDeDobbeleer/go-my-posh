---
id: exit
title: Exit code
sidebar_label: Exit code
---

## What

Displays the last exit code or that the last command failed based on the configuration.

## Sample Configuration

```json
{
  "type": "exit",
  "style": "diamond",
  "foreground": "#ffffff",
  "background": "#00897b",
  "leading_diamond": "",
  "trailing_diamond": "\uE0B4",
  "properties": {
    "display_exit_code": false,
    "always_enabled": true,
    "error_color": "#e91e63",
    "color_background": true,
    "prefix": "<#193549>\uE0B0</> \uE23A"
  }
}
```

## Properties

- display_exit_code: `boolean` - show or hide the exit code - defaults to `true`
- always_enabled: `boolean` - always show the status - defaults to `false`
- color_background: `boolean` - color the background or foreground when an error occurs - defaults to `false`
- error_color: `string` [hex color code][colors] - color to use when an error occured

[colors]: https://htmlcolorcodes.com/color-chart/material-design-color-chart/
