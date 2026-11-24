# Première participation de MarsGopherChess à un tournoi de moteurs

Pistes vérifiées le 24 novembre 2026 pour la base publique `baseline-2200`. Présenter le moteur comme **MarsGopherChess — Julien Richard**. Le moteur annonce `id name MarsGopherChess` et joue avec la recherche de la base 2200+. Le nom associe Marseille et le gopher de Go. Au 24 novembre, le nom exact était absent des listes de moteurs CCRL et ChessCache consultées. [Source du moteur proposé](https://github.com/jrichard8/EquinoxChess/tree/af8582a) · [Liste CCRL](https://computerchess.org/rating_list_all.html) · [ChessCache](https://www.chesscache.com/ChessEngines.html)

## Jouer depuis le Mac au blitz de Nightmare Chess

Joost Buijs organise ce tournoi mensuel de moteurs. Le 14 novembre 2026, l’édition s’est jouée en **11 rondes suisses à 4 minutes + 1 seconde par coup**, avec connexion demandée 15 minutes avant la première ronde. Le serveur accepte les moteurs UCI via un client. **La prochaine date reste à confirmer auprès de l’organisateur.** Il attribue les comptes moteur sur demande à `admin@nightmare-chess.nl`. Le Mac doit rester connecté et disponible pour toute la session. [Annonce de Joost Buijs et résultats de novembre](https://talkchess.com/forum/viewtopic.php?t=86963)

Brouillon pour demander le compte et la date, en anglais :

> **Subject — Engine account request for MarsGopherChess**
>
> Hello Joost,
>
> I develop MarsGopherChess, an open-source UCI chess engine written in Go. I would like to enter it in your next monthly engine blitz. Could you create an engine account for me and confirm the date, check-in time and how to join the tournament after logging in? I plan to run the engine on an Apple Silicon Mac using icsdrone-rs.
>
> [MarsGopherChess source code](https://github.com/jrichard8/EquinoxChess)
>
> Thank you,
>
> Julien Richard

Le client `icsdrone-rs` 2.2.17 pour Mac arm64 a démarré localement après vérification de l’empreinte de son archive. Le binaire macOS MarsGopherChess a donné les mêmes coups, scores, profondeurs et nombres de nœuds que le binaire qualifié sur les 77 positions v1, v2 et run6 en UCI à D6. Il a aussi joué un coup légal avec les pendules réglées à 4+1. Le compte et une connexion d’essai restent à valider avec l’organisateur. [Client macOS](https://github.com/jshriver/icsdrone-rs/releases/tag/v2.2.17) · [Configuration du client](https://github.com/jshriver/icsdrone-rs/blob/v2.2.17/config.example.json)

Sur le Mac de préparation, le moteur, le client et un exemple de configuration sont dans `.cache/tournois/nightmare-macos/`. Après attribution du compte, copier `config.example.json` vers `config.json` et remplacer le champ `Username` par l’identifiant attribué. Le client demande le mot de passe au lancement. Le fichier pointe sur le binaire testé et sur le port 5000 de `nightmare-chess.nl`. Depuis la racine du dépôt :

```sh
cd .cache/tournois/nightmare-macos
./icsdrone-rs --config config.json
```

## Proposer le moteur à l’Amateur Series

Graham Banks a accepté en octobre une candidature présentée avec un lien de version. On peut lui proposer MarsGopherChess pour une division adaptée à `gbanksnz@gmail.com`, avec le nom de l’auteur, les sources, le binaire Windows x64 et les conditions des matchs locaux. L’organisateur choisit les participants. Une participation ou une cote CCRL restent à obtenir. [Échange avec Graham Banks](https://talkchess.com/forum/viewtopic.php?p=998196)

L’archive locale `.cache/tournois/MarsGopherChess-windows-x64.zip` contient l’exécutable, sa fiche, la licence MIT et les empreintes SHA-256. Le binaire a été compilé depuis `af8582a`, vérifié au format Windows PE x86-64 et contrôlé après compression. Son exécution sous Windows reste à tester. Le code n’exige ni bibliothèque d’ouvertures, ni table de finales, ni réseau neuronal. Pour reconstruire l’exécutable depuis un checkout propre de cette révision avec Go 1.26.2 :

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GOAMD64=v1 \
  go build -trimpath -buildvcs=true -o /tmp/MarsGopherChess.exe ./cmd/uci
```

SHA-256 du binaire Windows : `69331aaab9d6cfd435057bbf7638730bb80b12747c6394e085d381e817f93bb6`.
