# XonStat
XonStat is a web application used to store and display information about games 
played in the arena-style FPS [Xonotic][xonotic]. It stores key details of games
(kills, deaths, score, etc) and aggregates them for later perusal. It does
so without requiring user registration or passwords by virtue of the wonderful
[d0_blind_id][d0_blind_id] library. XonStat is backed by [XonStatDB][xonstatdb], 
a PostgreSQL database.

This repository is a port of the [original XonStat][xonstat] application to the Go
programming language. 

## Prerequisites
xonstat-go requires the development headers of `libcairo2` to build. Install that with your system's package manager. 

## Clone
`git clone git@gitlab.com:xonotic/xonstat-go.git`

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
  -a, --addr string   listen address (default "0.0.0.0:6543")

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

### Extract Submissions from Logs

Sometimes stats submissions are interesting for one reason or another. In such scenarios it is
useful to be able to extract its POST body from the log for examination or even resubmission.
The `parseLog` subcommand is used for this purpose. Given a log file of request bodies, it can
extract every available submission out of it, or it can extract a single submission body by its 
match ID value ('I' line). In addition it can also anonymize the raw hashkeys provided to 
preserve player privacy.

```
Parse XonStat submission POST request log files into individual match files
for later inspection or even resubmission.

Usage:
  xonstat parseLog [flags]

Flags:
  -a, --anonymize      anonymize player hashkeys
  -f, --file string    log file
  -h, --help           help for parseLog
  -m, --match string   extract only this match_id

Global Flags:
  -c, --config string   config file (default is $HOME/.xonstat.toml)
```

### Submit Stats from Logs

The `submit` subcommand takes an output file written from the `parseLog` 
subcommand above and submits it to the stats server via the URL you provide. 

```
Submit stats requests from files.

Usage:
  xonstat submit [flags]

Flags:
  -f, --file string   submission request file
  -h, --help          help for submit
  -u, --url string    XonStat server submission URL (default "http://localhost:8080/stats/submit")

Global Flags:
  -c, --config string   config file (default is $HOME/.xonstat.toml)
```

### Generate Badges from Player Stats

The `badges` subcommand takes player data from the database and generates
images called "badges". Badges come in three styles and contain summary information
like K:D ratio, win ratio, time played, and the number of games played for the 
three most-played game types.

```
Generate badge images from player stats

Usage:
  xonstat badges [flags]

Flags:
  -a, --all           Generate for all players
  -d, --delta int     Generate for players having activity within this number of hours (default 6)
  -h, --help          help for badges
  -l, --limit int     Max number of badges to generate (default -1)
  -o, --out string    Output directory (default "output")
  -p, --pid int       Generate just for this player ID (default -1)
  -w, --workers int   Number of worker threads (default 5)

Global Flags:
  -c, --config string   config file (default is $HOME/.xonstat.toml)
```

### Version Information
The `version` subcommand prints out the application's build information. It's 
usage is as follows:

```
Show the build version in the form of the git commit hash from which the binary was built

Usage:
  xonstat version [flags]

Flags:
  -h, --help   help for version

Global Flags:
  -c, --config string   config file (default is $HOME/.xonstat.toml)
```

## Roadmap and Issues
Both roadmap items and issues are handled via the [GitLab issue tracker][issues].
Roadmap items will have the `task` tag applied.

## Chat
If you'd like to discuss this project or have a question about it, feel free to 
drop by the `#xonotic` channel on Freenode's IRC network. Please take care to 
follow [IRC etiquette][etiquette].

## License
XonStat is licensed under the GNU AGPLv3 license. See the `LICENSE.TXT` file
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
