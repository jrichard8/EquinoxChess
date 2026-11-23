# Deux effets des promotions sans prise en quiétude

**Septembre (`819a8d5`) reste la référence.** Les deux positions ont été rejouées à profondeur 6 avec la seule autre révision de l’expérience, `81fdfe3`. Sur Game17, le candidat prolonge des suites de prises en quiétude. Sur 0JLoe, une promotion change l’ordre des coups à la racine, puis la réduction de profondeur empêche de départager les deux poussées de pions. Les quatre recherches instrumentées reproduisent le coup, le score et le nombre de nœuds du [bilan apparié](paired.json).

## Game17_W_ply17 · où vont les 4 253 226 nœuds

| Visites, itérations 1 à 6 cumulées | Septembre | Candidat | Écart |
|---|---:|---:|---:|
| Recherche principale | 53 407 | 53 407 | 0 |
| Quiétude avant toute promotion sans prise admise par le candidat | 2 734 644 | 2 640 291 | −94 353 |
| Quiétude à partir de la première de ces promotions, enfant inclus | 0 | 4 347 579 | +4 347 579 |
| **Total** | **2 788 051** | **7 041 277** | **+4 253 226** |

Chaque moteur entre **41 694 fois** en quiétude depuis la recherche principale. L’empreinte des visites principales dans l’ordre d’exécution est identique, fenêtres alpha-bêta comprises (`49d84c88663f124423d6634698567037136de60da6338deaed5895627132b3ff`). Les six itérations choisissent `d1f3` avec le même score dans les deux moteurs, jusqu’à +67 cp à profondeur 6.

**Le pion qui promeut peut partir du milieu de l’échiquier.** Dans la branche la plus coûteuse après une première promotion, la recherche principale joue `d1f3 b8d7 a2a3 e7d6 g1h1`. La quiétude enchaîne ensuite `c5d4 c3d5 d4e3 f3f6 e3d2 f6d8`. Le pion noir de c5 vient de prendre trois fois et arrive en d2. Le candidat examine alors `d2d1q`, qui ouvre à lui seul un sous-arbre de **25 786 nœuds** dans cette occurrence. À cet endroit, septembre filtre la poussée sans prise. Cette branche est un exemple de suite explorée pendant la recherche.

Le candidat visite **198 871 premières promotions** hors échec sur leurs chemins respectifs. La sonde compte chaque sous-arbre une seule fois, même s’il contient d’autres promotions. Les promotions noires en d1 et b1 réunissent 87,6 % de ces 4 347 579 nœuds. Les promotions en dame en représentent 2 563 766 (59,0 %), les trois sous-promotions réunies 1 783 813. Le coût vient surtout des suites ouvertes par ces coups. **3 785 952** nœuds y sont atteints par une prise et **788 280** par une réponse à un échec, avec recouvrement lorsqu’une prise pare l’échec. Hors échec, les promotions sont exemptées de l’élagage delta. En échec, le moteur explore les évasions légales. Le surcoût net est inférieur au volume de ces branches, car les autres visites de quiétude diminuent de 94 353.

## 0JLoe · pourquoi h5h6 perd la priorité

La référence Stockfish du lot donne `h5h6 g7h6 g5h6 e7f8 e5f6 f8g8 h6h7`. Le premier coup sacrifie le pion h pour permettre au pion g de reprendre en h6. Les scores ci-dessous sont ceux d’Equinox, en centipions du point de vue des Blancs. [Position et analyse de référence](../../v2/positions.json)

1. **Le premier écart apparaît sur un troisième coup, `g5g6`.** Dès la profondeur 1, après `g5g6 f3h5 g6f7 h5f7`, le candidat peut encore jouer `a7a8q` en quiétude. Dans la fenêtre réellement explorée, `g5g6` remonte de −90 à −29. À profondeur 2, il reste sur la borne −54 chez septembre et atteint −39 chez le candidat. Il améliore donc alpha et reçoit un bonus d’historique chez le candidat. Les deux moteurs finissent pourtant ces deux itérations sur `f5f6` à −4.
2. **Ce bonus inverse l’ordre d’examen.** À profondeur 3, après les quatre promotions immédiates en a8, septembre classe `h5h6` et `f5f6` à 1 002 devant `g5g6` à 1 000 et examine `h5h6` en premier. Chez le candidat, les trois coups valent 1 002 et `g5g6` est le premier de la liste. Pendant sa recherche, le score d’ordre de `f5f6` monte à 1 003. Il passe devant `h5h6`. Ce dernier revient en tête à profondeur 4. À profondeur 5, l’historique classe néanmoins `f5f6` devant lui, 1 013 contre 1 008, puis 1 019 contre 1 005 à profondeur 6.
3. **La recherche réduite accepte ensuite une égalité à faible profondeur.** Aux profondeurs 5 et 6, `f5f6` pose alpha à +26. Le candidat sonde ensuite `h5h6` en recherche réduite (LMR), à profondeur enfant 3 dans les deux cas. Pour les Noirs, la fenêtre vaut [−27, −26]. La table de transposition contient déjà un score exact de −26 **à profondeur 3**, avec `g7h6` comme réponse. La sonde rend donc +26 aux Blancs en **un nœud**. La condition de reprise à pleine profondeur exige un score strictement supérieur à alpha. À égalité, le moteur garde `f5f6`.

| Contrôle à pleine profondeur, avec fenêtre ouverte | Septembre, h5h6 | Candidat, h5h6 | Septembre, f5f6 | Candidat, f5f6 |
|---|---:|---:|---:|---:|
| Profondeur racine 5, enfant 4 | +46 | +46 | +26 | +26 |
| Profondeur racine 6, enfant 5 | **+38** | **+38** | +26 | +26 |

Chaque contrôle part d’une copie du plateau, de la table et des heuristiques juste avant l’essai réduit du coup concerné. Il garde les règles de recherche de la révision et ouvre la fenêtre à pleine profondeur **pour ce coup à la racine**. Les réductions internes continuent de s’appliquer. Le contrôle du candidat consacre 2 538 nœuds à `h5h6` en profondeur 6 et retrouve +38. Son résultat est écarté avant de poursuivre la recherche observée. Le +26 utilisé en jeu explique le choix de `f5f6` sans constituer une mesure de `h5h6` à la profondeur demandée.

## Rejouer les observations

Depuis la racine du dépôt, avec Go 1.26 et les deux révisions disponibles localement, ces commandes appliquent la [sonde hors ligne](diagnostic.patch) uniquement aux copies temporaires. Les tables sont neuves pour chaque position, avec 65 536 entrées pour Game17 et 1 048 576 pour 0JLoe. L’empreinte SHA-256 porte sur les clés Zobrist, compteurs de demi-coups, niveaux, profondeurs et fenêtres de chaque entrée dans la recherche principale. Les recherches de contrôle ont leur propre compteur et leur propre empreinte.

```sh
set -eu
probe="$PWD/tools/tactics/experiments/quiet-promotions/diagnostic.patch"
replay=$(mktemp -d)
for revision in 819a8d5 81fdfe3; do
  mkdir "$replay/$revision"
  git archive "$revision" | tar -x -C "$replay/$revision"
  (
    cd "$replay/$revision"
    git apply "$probe"
    go test -count=1 -run '^TestQuietPromotionDiagnosis$' -v ./internal/search
  ) > "$replay/$revision.log"
done
echo "$replay"
```
