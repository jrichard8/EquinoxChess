# Testing Strategy

## Ce Qui Existe

Le projet couvre deja :

- perft position initiale jusqu'a profondeur 5
- Kiwipete jusqu'a profondeur 4
- tests de FEN et indexation
- tests make/unmake
- tests Zobrist
- tests de legalite des coups

Commande de validation locale :

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./...
```

## Ce Qui Doit Etre Ajoute Rapidement

### 1. Regression Search

Ajouter une suite de positions fixes qui verifie :

- mat en 1
- mat en 2 simple
- gain de piece evident
- evasion unique
- promotions critiques
- finales gagnees qui ne doivent pas finir en pat

### 2. Regression Evaluation

Verifier des invariants d'eval :

- symetrie blanc/noir
- bonus d'avantage materiel
- penalite de structure de pions
- valorisation des pions passes
- cohherence opening vs endgame

### 3. Regression Elo

Industrialiser les matches :

- mini gauntlet contre plusieurs niveaux Stockfish limites
- meme controle de temps
- memes openings ou positions de depart
- resultats archives avec date, commit et options moteur

### 4. Detection De Regressions Silencieuses

Surveiller apres chaque iteration :

- taux de nulles
- taux de pat
- repartition des mats subis
- profondeur atteinte
- nodes par seconde

## Politique

- Un perft vert ne signifie pas un gain de force.
- Un gain de vitesse ne vaut que s'il preserve ou augmente l'Elo.
- Une nouvelle heuristique sans harness de regression devient une source de derive.
