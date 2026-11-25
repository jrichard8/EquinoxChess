# Entering MarsGopherChess in its first engine tournament

These entry options were checked on November 24, 2026 for the public `baseline-2200` engine. Enter it as **MarsGopherChess — Julien Richard**. It reports `id name MarsGopherChess` and uses the 2200+ baseline search. The name combines Marseille and the Go gopher. As of November 24, the exact name did not appear in the CCRL or ChessCache engine lists checked. [Proposed engine source](https://github.com/jrichard8/MarsGopherChess/tree/a4a46cf) · [CCRL list](https://computerchess.org/rating_list_all.html) · [ChessCache](https://www.chesscache.com/ChessEngines.html)

## Play the Nightmare Chess blitz from a Mac

Joost Buijs organizes this monthly engine tournament. The November 14, 2026 edition used **11 Swiss rounds at 4 minutes plus 1 second per move**. Players were asked to connect 15 minutes before round one. The server accepts UCI engines through a client. **Confirm the next date with the organizer.** Request an engine account at `admin@nightmare-chess.nl`. The Mac must stay online and available throughout the session. [Joost Buijs's announcement and November results](https://talkchess.com/forum/viewtopic.php?t=86963)

Draft account and date request

> **Subject — Engine account request for MarsGopherChess**
>
> Hello Joost,
>
> I develop MarsGopherChess, an open-source UCI chess engine written in Go. I would like to enter it in your next monthly engine blitz. Could you create an engine account for me and confirm the date, check-in time and how to join the tournament after logging in? I plan to run the engine on an Apple Silicon Mac using icsdrone-rs.
>
> [MarsGopherChess source code](https://github.com/jrichard8/MarsGopherChess)
>
> Thank you,
>
> Julien Richard

The `icsdrone-rs` 2.2.17 client for macOS arm64 started locally after its archive checksum was verified. Over all 77 v1, v2 and run6 positions at UCI depth 6, the macOS MarsGopherChess binary matched the qualified executable on move, score, depth and nodes. It also played a legal move with 4+1 clocks. The account and a test connection still need to be checked with the organizer. [macOS client](https://github.com/jshriver/icsdrone-rs/releases/tag/v2.2.17) · [Client configuration](https://github.com/jshriver/icsdrone-rs/blob/v2.2.17/config.example.json)

On the Mac used for preparation, the engine, client and sample configuration are in `.cache/tournois/nightmare-macos/`. Once the account is issued, copy `config.example.json` to `config.json` and set `Username` to the assigned account name. The client prompts for the password at startup. The configuration points to the tested binary and port 5000 on `nightmare-chess.nl`. Run from the repository root.

```sh
cd .cache/tournois/nightmare-macos
./icsdrone-rs --config config.json
```

## Submit the engine to the Amateur Series

In October, Graham Banks accepted an entry submitted with a release link. To propose MarsGopherChess for a suitable division, email `gbanksnz@gmail.com` with the author's name, source, Windows x64 binary and local match conditions. The organizer selects the participants. Entry and a CCRL rating are still pending. [Discussion with Graham Banks](https://talkchess.com/forum/viewtopic.php?p=998196)

The local archive `.cache/tournois/MarsGopherChess-windows-x64.zip` contains the executable, its information sheet, the MIT license and SHA-256 checksums. The archived binary corresponds to the source at `a4a46cf`. Its Windows PE x86-64 format and checksum after compression were verified. Running it on Windows is still untested. The engine requires no opening book, endgame tablebases or neural network. To rebuild it, run the following in a clean checkout of that revision with Go 1.26.2.

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GOAMD64=v1 \
  go build -trimpath -buildvcs=true -o /tmp/MarsGopherChess.exe ./cmd/uci
```

The rebuild at `a4a46cf` has SHA-256 `ebb2805c850b465ffa0c37fd80a16a4dac221beb21ff51630d37bb1cfcebe745`. The executable archived before the history rewrite has SHA-256 `69331aaab9d6cfd435057bbf7638730bb80b12747c6394e085d381e817f93bb6`.
