# Berlin Termin-prufen-go

<!-- TOC -->

- [Berlin Termin-prufen-go](#berlin-termin-prufen-go)
  - [Disclaimer](#disclaimer)
  - [CLI](#cli)
  - [API](#api)
  - [Config file](#config-file)

<!-- /TOC -->

## Disclaimer

Highly likely this project won't be ever supported,
I've implemented it privately half a year before this README was created.
I'd successfully got my residence permit, forgot about the project, found it again, and decided to finally push in as-is state (and get rid of hardcode and similar stuff).

## CLI

Install [Go binary file](https://go.dev/dl/) and then open a terminal window.

Install the `termin-prufen-go` binary with the following command in the terminal window:

```bash
go install github.com/zerospiel/termin-prufen-go@latest
```

Adjust a [configuration file](#config-file), refering the following example:

```yaml
# ABH config example
citizenship: "Russian Federation"
people_number: 2
live_in_berlin: "yes"
family_member_citizenship: "Russian Federation"
# reason: "apply" # "extend" is not supported

# Telegram API config example
telegram_chat_id: 12345678
telegram_bot_token: "1234567890:qwertyuiopasdfghjklzxcvbnmQWERTYUIO"

# Application config example
# screenshots_dir: "path/to/put/screenshots/to" # mostly for debug
scenario_timeout: 50s
poll_interval: 5m
# single_run_mode: false
# debug: false
```

Follow the links on how to obtain [Telegram
Chat ID](https://stackoverflow.com/a/32572159/1561149) and [Telegram Bot Token](https://core.telegram.org/bots/tutorial#obtain-your-bot-token).

(BTW, I've no idea if the above links are correct, but at least the first had helped me and the second is the link to the official website).

Start the binary with the following command in the terminal window:

```bash
./termin-prufen-go --config-file config.yaml

```

Check messages in the corresponding chat.

Please, be advised, if you don't use `single_run_mode`, ensure that
the terminal window does continue to be opened (even in background)
or either start `termin-prufen-go` on any dedicated machine.

## API

TODO: examples of how to use the module's api

## Config file

TODO: detailed descriptions of options
