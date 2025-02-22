# YaGoStatus
Yet Another i3status replacement written in Go.

[![GitHub release](https://img.shields.io/github/release/burik666/yagostatus.svg)](https://github.com/burik666/yagostatus)
[![Build Status](https://cloud.drone.io/api/badges/burik666/yagostatus/status.svg)](https://cloud.drone.io/burik666/yagostatus)
[![GitHub license](https://img.shields.io/github/license/burik666/yagostatus.svg)](https://github.com/burik666/yagostatus/blob/master/LICENSE)


![yagostatus.gif](https://raw.githubusercontent.com/wiki/burik666/yagostatus/yagostatus.gif)

## Features
- Instant and independent updating of widgets.
- Handling click events.
- Shell scripting widgets and events handlers.
- Wrapping other status programs (i3status, py3status, conky, etc.).
- Different widgets on different workspaces.
- Templates for widgets outputs.
- Update widget via http/websocket requests.
- Update widget by POSIX Real-Time Signals (SIGRTMIN-SIGRTMAX).
- [Snippets](https://github.com/burik666/ygs-snippets).
- [Plugins](plugins).

## Installation

    go install github.com/burik666/yagostatus@latest
    mkdir -p ~/.config/yagostatus
    yagostatus -dump > ~/.config/yagostatus/yagostatus.yml

Replace `status_command` to `~/go/bin/yagostatus --config ~/.config/yagostatus/yagostatus.yml` in your i3 config file.

If you using Sway add the `--sway` parameter.

### Troubleshooting
Yagostatus outputs error messages in stderr, you can log them by redirecting stderr to a file.

`status_command exec ~/go/bin/yagostatus --config /path/to/yagostatus.yml 2> /tmp/yagostatus.log`

## Configuration

If `--config` is not specified, yagostatus is looking for `yagostatus.yml` in `$HOME/.config/yagostatus` (or `$XDG_HOME_CONFIG/yagostatus` if set) or in the current working directory.

Yagostatus uses a configuration file in the yaml format.

Example:
```yml
widgets:
  - widget: static
    blocks: >
        [
            {
                "full_text": "YaGoStatus",
                "color": "#2e9ef4"
            }
        ]
    events:
      - button: 1
        command: xdg-open https://github.com/burik666/yagostatus/

  - widget: wrapper
    command: /usr/bin/i3status

  - widget: clock
    format: Jan _2 Mon 15:04:05 # https://golang.org/pkg/time/#Time.Format
    templates: >
        [{
            "color": "#ffffff",
            "separator": true,
            "separator_block_width": 21
        }]
```
## Widgets

### Common parameters

- `widget` - Widget name.
- `workspaces` - List of workspaces to display the widget.

Example:
```yml
- widget: static
  workspaces:
    - "1:www"
    - "2:IM"

  blocks: >
    [
        {
            "full_text": "Visible only on 1:www and 2:IM workspaces"
        }
    ]

- widget: static
  workspaces:
    - "!1:www"

  blocks: >
    [
        {
            "full_text": "Visible on all workspaces except 1:www"
        }
    ]
```

- `templates` - The templates that apply to widget blocks.
- `events` - List of commands to be executed on user actions.
    * `button` - X11 button ID (0 for any, 1 to 3 for left/middle/right mouse button. 4/5 for mouse wheel up/down. Default: `0`).
    * `modifiers` - List of X11 modifiers condition.
    * `command` - Command to execute (via `sh -c`).
    Сlick_event json will be written to stdin.
    Also env variables are available: `$I3_NAME`, `$I3_INSTANCE`, `$I3_BUTTON`, `$I3_MODIFIERS`, `$I3_{X,Y}`, `$I3_OUTPUT_{X,Y}`, `$I3_RELATIVE_{X,Y}`, `$I3_{WIDTH,HEIGHT}`, `$I3_MODIFIERS`.
    The clicked widget fields are available as ENV variables with the prefix `I3_` (example:` $ I3_full_text`).
    * `workdir` - Set a working directory.
    * `env` - Set environment variables.
    * `output_format` - The command output format (`none`, `text`, `json`, `auto`) (default: `none`).
    * `name` - Filter by `name` for widgets with multiple blocks (default: empty).
    * `instance` - Filter by `instance` for widgets with multiple blocks (default: empty).
    * `override` - If `true`, previously defined events with the same `button`, `modifier`, `name` and `instance` will be ignored (default: `false`)

Example:
```yml
- widget: static
  blocks: >
    [
        {
            "full_text": "Firefox",
            "name": "ff"
        },
        {
            "full_text": "Chrome",
            "name": "ch"
        }
    ]
  templates: >
    [
        {
            "color": "#ff8000"
        },
        {
            "color": "#ff3030"
        }
    ]
  events:
    - button: 1
      command: /usr/bin/firefox
      name: ff

    - button: 1
      modifiers:
        - "!Control" # "!" must be quoted
      command: /usr/bin/chrome
      name: ch

    - button: 1
        - Control
      command: /usr/bin/chrome --incognito
      name: ch
```

### Snippets

Yagostatus supports the inclusion of snippets from files.
```yml
  - widget: $ygs-snippets/snip.yaml
    msg: hello world
    color: #00ff00
```

`ygs-snippets/snip.yaml`:
```yml
variables:
  msg: "default messsage"
  color: #ffffff
widgets:
  - widget: static
    blocks: >
        [
            {
                "full_text": "message: ${msg}",
                "color": "${color}"
            }
        ]
```

`ygs-snippets/snip.yaml` - relative path from the current file.


### Widget `clock`

The clock widget returns the current time in the specified format.

- `format` - Time format (https://golang.org/pkg/time/#Time.Format).
- `interval` - Clock update interval in seconds (default: `1`).


### Widget `exec`

This widget runs the command at the specified interval.

- `command` - Command to execute (via `sh -c`).
- `workdir` - Set a working directory.
- `env` - Set environment variables.
- `interval` - Update interval in seconds (`0` to run once at start; `-1` for loop without delay; default: `0`).
- `retry` - Retry interval in seconds if command failed (default: none).
- `silent` - Don't show error widget if command failed (default: `false`).
- `events_update` - Update widget if an event occurred (default: `false`).
- `output_format` - The command output format (`none`, `text`, `json`, `auto`) (default: `auto`).
- `signal` - SIGRTMIN offset to update widget. Should be between 0 and `SIGRTMIN`-`SIGRTMAX`.

The current widget fields are available as ENV variables with the prefix `I3_` (example: `$I3_full_text`).
For widgets with multiple blocks, an suffix with an index will be added. (example: `$I3_full_text`, `$I3_full_text_1`, `$I3_full_text_2`, etc.)

Use pkill to send signals:

    pkill -SIGRTMIN+1 yagostatus


### Widget `wrapper`

The wrapper widget starts the command and proxy received blocks (and click_events).
See: https://i3wm.org/docs/i3bar-protocol.html

- `command` - Command to execute.
- `workdir` - Set a working directory.
- 'env' - Set environment variables.


### Widget `static`

The static widget renders the blocks. Useful for labels and buttons.

- `blocks` - JSON List of i3bar blocks.


### Widget `http`

The http widget starts http server and accept HTTP or Websocket requests.

- `network` - `tcp` or `unix` (default `tcp`).
- `listen` - Hostname and port or path to the socket file to bind (example: `localhost:9900`, `/tmp/yagostatus.sock`).
- `path` - Path for receiving requests (example: `/mystatus/`).
Must be unique for multiple widgets with same `listen`.

For example, you can update the widget with the following command:

    curl http://localhost:9900/mystatus/ -d '[{"full_text": "hello"}, {"full_text": "world"}]'

Send an empty array to clear:

    curl http://localhost:9900/mystatus/ -d '[]'

Unix socket:

    curl --unix-socket /tmp/yagostatus.sock localhost/mystatus/ -d '[{"full_text": "hello"}]'


## Examples

### Counter

This example shows how you can use custom fields.

- Left mouse button - increment
- Right mouse button - decrement
- Middle mouse button - reset

```yml
  - widget: static
    blocks: >
        [
            {
                "full_text":"COUNTER"
            }
        ]
    events:
      - command: |
          printf '[{"full_text":"Counter: %d", "_count":%d}]' $((I3__COUNT + 1)) $((I3__COUNT + 1))
        output_format: json
        button: 1
      - command: |
          printf '[{"full_text":"Counter: %d", "_count":%d}]' $((I3__COUNT - 1)) $((I3__COUNT - 1))
        output_format: json
        button: 3
      - command: |
          printf '[{"full_text":"Counter: 0", "_count":0}]'
        output_format: json
        button: 2
```

### Volume control
i3 config:
```
bindsym XF86AudioLowerVolume exec pactl set-sink-volume @DEFAULT_SINK@ -1%; exec pkill -SIGRTMIN+1 yagostatus
bindsym XF86AudioRaiseVolume exec pactl set-sink-volume @DEFAULT_SINK@ +1%; exec pkill -SIGRTMIN+1 yagostatus
bindsym XF86AudioMute exec pactl set-sink-mute @DEFAULT_SINK@ toggle; exec pkill -SIGRTMIN+1 yagostatus
```

```yml
  - widget: exec
    command: |
        color="#ffffff"
        if [ $(pacmd list-sinks |sed  '1,/* index/d'|grep -E '^\smuted:'|head -n1|awk '{print $2}') = "yes" ]; then
          color="#ff0000"
        fi
        volume=$(pacmd list-sinks |sed  '1,/* index/d'|grep -E '^\svolume:'|head -n1|awk '{print $5}')
        echo -e '[{"full_text":"♪ '${volume}'","color":"'$color'"}]'

    interval: 0
    signal: 1
    events_update: true
    events:
        - button: 1
          command: pactl set-sink-mute @DEFAULT_SINK@ toggle

        - button: 4
          command: pactl set-sink-volume @DEFAULT_SINK@ +1%

        - button: 5
          command: pactl set-sink-volume @DEFAULT_SINK@ -1%

    templates: >
        [{
            "markup": "pango",
            "separator": true,
            "separator_block_width": 21
        }]
```

### Weather

To get access to weather API you need an APIID.
See https://openweathermap.org/appid for details.

Requires [jq](https://stedolan.github.com/jq/) for json parsing.

```yml
  - widget: static
    blocks: >
        [
            {
                "full_text": "Weather:",
                "color": "#2e9ef4",
                "separator": false
            }
        ]
  - widget: exec
    command: curl -s 'http://api.openweathermap.org/data/2.5/weather?q=London,uk&units=metric&appid=<APPID>'|jq .main.temp
    interval: 300
    templates: >
        [{
            "separator": true,
            "separator_block_width": 21
        }]
```

You can use the [weather snippet](https://github.com/burik666/ygs-snippets/tree/master/weather) instead.

### Conky

```yml
    - widget: wrapper
      command: /usr/bin/conky -c /home/burik/.config/i3/conky.conf
```
Specify the full path to `conky.conf`.

**conky.conf** (conky 1.10 or higher):
```lua
conky.config = {
    out_to_x = false,
    own_window = false,
    out_to_console = true,
    background = false,
    max_text_width = 0,

    -- Update interval in seconds
    update_interval = 1.0,

    -- This is the number of times Conky will update before quitting.
    -- Set to zero to run forever.
    total_run_times = 0,

    -- Shortens units to a single character (kiB->k, GiB->G, etc.). Default is off.
    -- short_units yes

    -- Add spaces to keep things from moving about?  This only affects certain objects.
    -- use_spacer should have an argument of left, right, or none
    use_spacer = 'left',

    -- Force UTF8? note that UTF8 support required XFT
    override_utf8_locale = false,

    -- number of cpu samples to average
    -- set to 1 to disable averaging
    cpu_avg_samples = 3,
    net_avg_samples = 3,
    diskio_avg_samples = 3,

    format_human_readable = true,
    lua_load = '~/.config/i3/conky.lua',
};

conky.text = [[

[

{ "full_text": "CPU:", "color": "\#2e9ef4", "separator": false },
{ ${lua_parse cpu cpu1} , "min_width": "100%", "align": "right", "separator": false },
{ ${lua_parse cpu cpu2} , "min_width": "100%", "align": "right", "separator": false },
{ ${lua_parse cpu cpu3} , "min_width": "100%", "align": "right", "separator": false },
{ ${lua_parse cpu cpu4} , "min_width": "100%", "align": "right", "separator": true, "separator_block_width":21 },

{ "full_text": "RAM:", "color": "\#2e9ef4", "separator": false },
{ "full_text": "${mem} / ${memeasyfree}", "color": ${if_match ${memperc}<80}"\#ffffff"${else}"\#ff0000"${endif}, "separator": true, "separator_block_width":21 },

{ "full_text": "sda:", "color": "\#2e9ef4", "separator": false },
{ "full_text": "▼ ${diskio_read sda} ▲ ${diskio_write sda}", "color": "\#ffffff", "separator": true, "separator_block_width":21 },

{ "full_text": "eth0:", "color": "\#2e9ef4", "separator": false },
{ "full_text": "▼ ${downspeed eth0} ▲ ${upspeed eth0}", "color": "\#ffffff", "separator": true, "separator_block_width":21 }

]

]];
```


**conky.lua**:
```lua
function gradient_red(min, max, val)
  local min = tonumber(min)
  local max = tonumber(max)
  local val = tonumber(val)
  if (val > max) then val = max end
  if (val < min) then val = min end

  local v = val - min
  local d = (max - min) * 0.5
  local red, green

  if (v <= d) then
    red = math.floor((255 * v) / d + 0.5)
    green = 255
  else
    red = 255
    green = math.floor(255 - (255 * (v-d)) / (max - min - d) + 0.5)
  end

  return string.format("%02x%02x00", red, green)
end

function conky_cpu (cpun)
    local val = tonumber(conky_parse("${cpu " .. cpun .. "}"))
    if val == nil then val = 0 end
    return "\"full_text\": \"" .. val  .. "%\", \"color\": \"\\#" .. gradient_red(0, 100, val) .. "\""
end
```

## License

YaGoStatus is licensed under the GNU GPLv3 License.

