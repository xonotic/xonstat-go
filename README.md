# XonStat
XonStat is a web application used to store and display information about games 
played in the arena-style FPS [Xonotic][xonotic]. It stores key details of games
(kills, deaths, score, etc) and aggregates them for later perusal. It does
so without requiring user registration or passwords by virtue of the wonderful
[d0_blind_id][d0_blind_id] library. XonStat is backed by [XonStatDB][xonstatdb], 
a PostgreSQL database.

This repository is a port of the [original XonStat][xonstat] application to the Go
programming language. 

## Install
`go get gitlab.com/xonotic/xonstat-go`

## Build
`make`

## Test
To run the test suite for all sub-packages in the project:

`make test`

To run the test suite and also view code coverage information in your browser: 

`make coverage`

To clean up build artifacts:

`make clean`

## Run
Before running the application or any of its components, make sure you've 
followed all of the [XonStatDB][xonstatdb] setup steps. Much of the 
functionality of the project depends on this database, so you will not get 
far without it!

The project makes use of the [Cobra][cobra] library for subcommand support and
the [Viper][viper] library for configuration. Start by copying the 
`configs/xonstat.toml` file in this repository to your home directory as 
`$HOME/.xonstat.toml`, then edit the file to match your preferences. 

### Web Application
The `web` subcommand starts up the primary web application instance. It's 
usage is as follows:

```
Run the XonStat web application server.

Usage:
  xonstat web [flags]

Flags:
  -h, --help          help for web
  -p, --port string   port number (default "8080")

Global Flags:
  -c, --config string   config file (default is $HOME/.xonstat.toml)
  ```

### Submission Inspector
If you have a Xonotic server stats request body saved to a text file (like 
those found in the `tests/submissions` directory), you can inspect the 
parsed form of its contents in JSON format with the `inspect` subcommand. 
It's usage is as follows:

```
Inspect XonStat submission files using the JSON format. This is mostly used for 
debugging purposes when you want to see the raw data that results from the parsing phase
of the stats submission process.

Usage:
  xonstat inspect [flags]

Flags:
  -f, --file string   submission file
  -h, --help          help for inspect

Global Flags:
  -c, --config string   config file (default is $HOME/.xonstat.toml)
```
## Roadmap and Issues
Both roadmap items and issues are handled via the [GitLab issue tracker][issues].
Roadmap items will have the `task` tag applied.

## Chat
If you'd like to discuss this project or have a question about it, feel free to 
drop by the `#xonotic` channel on Freenode's IRC network. Please take care to 
follow IRC etiquette.

## License
XonStat is licensed under the GNU GPLv3 license. See the `LICENSE.TXT` file
within this repository for the full legal text.

## Author
[Ant 'Antibody' Zucaro][antibody_profile] is to blame for this project.


[antibody_profile]: https://gitlab.com/antibody
[cobra]: https://github.com/spf13/cobra
[d0_blind_id]: https://gitlab.com/xonotic/d0_blind_id
[etiquette]: https://github.com/fizerkhan/irc-etiquette#irc-etiquette-by-christoph-haas
[issues]: https://gitlab.com/xonotic/xonstat-go/-/issues
[viper]: https://github.com/spf13/viper
[xonotic]: https://www.xonotic.org
[xonstat]: https://gitlab.com/xonotic/xonstat
[xonstatdb]: https://gitlab.com/xonotic/xonstatdb