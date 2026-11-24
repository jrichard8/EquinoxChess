# Calibration contre Stockfish 18 limité à 1850

Le seuil de travail est un score supérieur à 50 % contre Stockfish 18 réglé à 1850, à 1 seconde + 0,1 seconde par coup, avec une borne basse de l’intervalle bilatéral à 95 % supérieure à 50 %. Le calcul utilise une observation par paire d’ouverture, selon le [protocole des matchs](../README.md). Cette échelle dépend de l’adversaire, de sa configuration, du matériel et de la cadence.

## Sélection et confirmation

- Comparer à septembre deux configurations successives sur les 64 ouvertures de `openings-v2.epd`, couleurs inversées. La première accélère le calcul de l’évaluation. La seconde ajoute le tri des prises en quiétude.
- Vérifier v1, v2 et les 13 canaris du run6 à profondeur 6. Pour confirmer la configuration complète, exiger un gain direct contre septembre selon l’intervalle du projet, la conservation des coups de référence déjà trouvés et la réussite des tests.
- Tester une seule configuration retenue sur les **192 nouvelles ouvertures** de ce dossier contre Stockfish 1850. Jouer les 384 parties et calculer l’intervalle sur les 192 paires, sans arrêt anticipé. Garder ce résultat distinct des matchs de sélection.

## Origine des ouvertures

La suite a été figée dans `b971f8c` avant sa première partie avec Equinox. Le générateur part de la position initiale, joue huit demi-coups et choisit parmi cinq variantes de Stockfish à 3 000 nœuds restant à 60 cp du meilleur coup. Une vérification à 16 000 nœuds garde les positions entre −60 et +60 cp. La graine est 18502200. Le générateur rejette les positions finales déjà présentes dans la suite précédente ou dans ce lot, en ignorant les deux compteurs FEN. Plusieurs lignes peuvent appartenir à la même famille d’ouvertures.

Stockfish utilise un thread et 16 Mio de table. L’empreinte SHA-256 du fichier EPD est `590061cbd1ff2801ffac0e97e872835d13afc2a02119e9b9398e81284f6bcfac`. [Générateur](build_holdout.py) · [Ouvertures](holdout.epd)
