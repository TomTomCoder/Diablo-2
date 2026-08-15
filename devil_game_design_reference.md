# Devil — Référence Game Design

> **Basé sur** : mécanique Diablo 2 (Project Diablo 2 Wiki, *Reverse Design: Diablo 2* — Holleman, CRC Press 2018), adapté pour Devil
> **Dernière mise à jour** : mai 2026
>
> **⚠️ Mise à jour août 2026 — Devil devient multi-classes.** Suite à la réception d'assets visuels pour une deuxième classe (le Warrior, combattant au corps-à-corps physique), Devil n'est plus un jeu à classe unique. Ce document a été rédigé pour un Mage seul et n'a **pas encore été révisé intégralement** pour refléter ce changement — les sections §1, §2, §5 et §8 ci-dessous contiennent des affirmations "classe unique"/"pas de Strength"/"aucune arme physique" qui ne s'appliquent désormais qu'**au Mage**, pas au jeu dans son ensemble. Voir `devil_warrior_character_design.md` §4 pour la liste de ce qui reste à concevoir côté Warrior (attributs, arbre de compétences, itémisation, formule de dégâts) — volontairement non inventé ici.

---

## 1. Concept et boucle de jeu principale

**Devil** est un action-RPG solo dans lequel le joueur incarne un héros — **Mage ou Warrior** (choix de classe à la création, voir §2) — qui doit monter en puissance pour affronter Devil, un mage corrompu par la magie noire qui a créé des entités maléfiques recouvrant toute la planète.

La boucle principale repose sur trois piliers :

1. **Combat magique** : éliminer des vagues d'entités via des sorts actifs et passifs
2. **Loot** : collecter et identifier des objets et artefacts magiques
3. **Progression** : gagner de l'XP, monter en niveau, allouer des points d'Energy et de compétences

---

## 2. Création du personnage

> **⚠️ Section à réécrire** — ce qui suit décrivait l'ancien design à classe unique. Devil propose désormais un choix de classe (Mage / Warrior) à la création ; les détails exacts (attributs de départ du Warrior, équipement de départ, etc.) restent à définir — voir `devil_warrior_character_design.md` §4.

À la création d'une nouvelle partie, le joueur choisit :

- **La classe** : Mage ou Warrior (impact sur les stats, l'arbre de compétences et l'itémisation — voir `devil_mage_character_design.md` et `devil_warrior_character_design.md`)
- **Le nom** du personnage

Le Mage dispose de deux formes visuelles (Mage / Mage Divin, une progression de puissance plutôt qu'un choix cosmétique — voir `devil_mage_character_design.md` §1). Le Warrior n'a pas encore de variantes définies.

---

## 3. Structure narrative et monde

Le monde est divisé en **5 régions** progressives, chacune dominée par une entité maléfique créée par Devil. La progression est linéaire mais les zones sont générées procéduralement à chaque partie.

| Région | Environnement | Entité finale |
|--------|---------------|---------------|
| I | Terres dévastées / Ruines | Gardien des Cendres |
| II | Marécages corrompus | La Noyée |
| III | Forêt de cristal maudit | L'Entrelaceur |
| IV | Citadelle du Néant | Ombre Prime |
| V | Tour de Devil | **Devil** (boss final) |

### Niveaux de difficulté

Chaque région est rejouable en **trois niveaux de difficulté** : Éveil, Corruption, Apocalypse. Les entités y sont plus puissantes, leurs résistances augmentent modérément (sans jamais atteindre des valeurs négatives punitives), et le loot de qualité supérieure y apparaît plus fréquemment.

---

## 4. Génération procédurale

### 4.1 Cartes

Les cartes sont générées par assemblage de blocs pré-fabriqués thématiques. Chaque partie produit une structure différente tout en restant cohérente visuellement.

**Améliorations par rapport à Diablo 2 :**
- Mini-map toujours visible, mise à jour en temps réel au fur et à mesure de l'exploration
- Tiles conçus pour éviter les labyrinthes sans issue (couloirs clairs, objectifs indiqués)
- Points d'intérêt (boss, waypoints, coffres rares) signalés par une icône sur la mini-map dès leur découverte

### 4.2 Loot (Treasure Classes)

Lorsqu'une entité est tuée, le jeu sélectionne dans une Treasure Class (TC) définie par le niveau de la zone. Chaque TC contient une liste d'objets avec des probabilités associées, un nombre de "picks" et une valeur NoDrop.

**Drop rates** : les taux de base sont généreux sur les niveaux intermédiaires. Les objets très rares (hauts TC) restent difficiles à obtenir mais sont soutenus par un système d'**Events temporaires** (voir section 9).

Le **Magic Find (MF)** reste l'attribut clé pour augmenter la probabilité d'obtenir des objets Magiques, Rares, Set ou Uniques.

---

## 5. Attributs du Mage

> **⚠️ Spécifique au Mage.** Ce qui suit (suppression de Strength, mitigation 100% magique) ne s'applique plus qu'à la classe Mage. Les attributs du Warrior restent à définir — voir `devil_warrior_character_design.md` §4.

La force physique n'existant pas dans ce monde dominé par la magie **pour le Mage**, ses attributs sont recentrés :

| Attribut | Rôle |
|----------|------|
| **Energy** | Augmente les dégâts magiques ET la réserve de mana. C'est l'attribut offensif et ressource principal. |
| **Vitality** | Détermine les points de vie. Essentiel pour la survie. |
| **Dexterity** | Améliore la vitesse d'incantation et réduit le temps de récupération après un impact. |

À chaque montée de niveau : **5 points d'attributs** à répartir librement + **1 point de compétence**.

> **Suppression (Mage uniquement)** : l'attribut Strength n'existe pas pour le Mage. La Défense physique non plus — toute sa mitigation de dégâts passe par les résistances magiques et les sorts défensifs. Le Warrior a probablement besoin de Strength et/ou de Défense physique ; à concevoir séparément.

---

## 6. Système de combat

### Moteur de calcul

Le moteur fonctionne en arithmétique entière avec précision 1/256 via bit-shifting pour les valeurs continues (mana, régénération). Les animations d'incantation suivent un système de breakpoints : certains seuils de Dexterity font passer l'animation à la frame inférieure, réduisant le temps de cast.

### Mécaniques offensives clés

| Mécanique | Description |
|-----------|-------------|
| **Puissance magique** | Dépend directement du niveau d'Energy : `Dégâts = base_sort × (1 + Energy / 100)` |
| **Surcharge (Overload)** | Équivalent du Crushing Blow : consume du mana supplémentaire pour infliger un % fixe de la vie actuelle de l'entité |
| **Marque ardente** | Applique des dégâts par seconde après impact (ignore la régénération) |
| **Breakpoints d'incantation** | Seuils de Dexterity où la vitesse d'animation de cast diminue d'une frame |

### Défense et résistances

Toute la mitigation est magique :

- **Résistances élémentaires** : Feu, Froid, Foudre, Ombre (cap à 75% en Éveil, cap maintenu en Corruption et Apocalypse — jamais négatif)
- **Bouclier de mana** : sort défensif permettant d'absorber des dégâts avec la réserve de mana
- **Régénération de mana** : attribut clé, boosté par l'Energy et certains objets

---

## 7. Arbre de compétences du Mage

Le Mage dispose de **30 compétences** réparties en 3 arbres de 10 compétences. Les arbres partagent des synergies inter-branches. Un point dans une compétence inférieure est requis pour débloquer celles du niveau au-dessus.

### Arbre I — Élémentalisme (dégâts directs)

| Niveau requis | Compétence | Type | Description |
|---------------|------------|------|-------------|
| 1 | Trait de feu | Actif | Projectile de feu à courte portée, dégâts immédiats |
| 1 | Éclat de glace | Actif | Projectile froid, ralentit la cible |
| 6 | Éclair en chaîne | Actif | Foudre qui rebondit sur 3 cibles |
| 6 | Nova de givre | Actif | Explosion de froid en zone autour du Mage |
| 12 | Boule de feu | Actif | Projectile AoE, dégâts feu élevés |
| 12 | Tempête statique | Actif | Invoque une zone d'éclair persistante |
| 18 | Orbe glaciale | Actif | Projectile lent, explose en large AoE de froid |
| 18 | Maîtrise élémentaire | Passif | Augmente tous les dégâts élémentaires (+% par point) |
| 24 | Météore | Actif | Frappe retardée sur une zone, dégâts feu massifs |
| 30 | Apocalypse | Actif | Pluie d'éclairs + feu + froid sur toute la zone visible |

### Arbre II — Arcane (contrôle et amplification)

| Niveau requis | Compétence | Type | Description |
|---------------|------------|------|-------------|
| 1 | Télékinésie | Actif | Repousse les entités, interaction avec les objets à distance |
| 1 | Champ statique | Actif | Réduit la vie de toutes les entités à l'écran d'un % fixe |
| 6 | Ralentissement | Actif | Zone qui réduit la vitesse des entités de 50% |
| 6 | Amplification | Actif | Augmente les dégâts magiques reçus par la cible |
| 12 | Téléportation | Actif | Déplacement instantané vers la position visée |
| 12 | Rupture arcane | Actif | Projectile qui supprime les résistances d'une cible |
| 18 | Prison de glace | Actif | Immobilise un groupe d'entités |
| 18 | Résonance magique | Passif | Synergie : chaque sort lancé augmente les dégâts du suivant (+% temporaire) |
| 24 | Vortex | Actif | Aspire toutes les entités proches vers un point |
| 30 | Distorsion temporelle | Actif | Ralentit toutes les entités à l'écran pendant 5 secondes |

### Arbre III — Ésotérisme (défense, survie, invocations)

| Niveau requis | Compétence | Type | Description |
|---------------|------------|------|-------------|
| 1 | Bouclier de mana | Actif | Absorbe les dégâts avec la réserve de mana |
| 1 | Régénération accélérée | Passif | Augmente la vitesse de régénération du mana |
| 6 | Armure de glace | Actif | Réduit les dégâts reçus et ralentit les attaquants au contact |
| 6 | Familier | Actif | Invoque une entité magique alliée qui attaque les cibles proches |
| 12 | Double ésotérique | Actif | Crée un double illusoire du Mage pour distraire les entités |
| 12 | Absorption d'énergie | Passif | Chaque entité tuée restaure un % de mana |
| 18 | Golem arcane | Actif | Invoque un golem de mana qui absorbe les dégâts à la place du Mage |
| 18 | Transcendance | Passif | Au lieu de mourir, le Mage se régénère une fois par zone (longue recharge) |
| 24 | Tempête de lames | Actif | Invoque des lames de mana orbitant autour du Mage |
| 30 | Éveil du Nexus | Actif | Ultime défensif : immunité magique pendant 8 secondes, soigne le Mage |

### Règles des synergies

- Chaque point dans **Trait de feu** augmente les dégâts de **Boule de feu** et **Météore**
- Chaque point dans **Éclat de glace** augmente la durée de gel de **Nova de givre**, **Orbe glaciale** et **Prison de glace**
- Chaque point dans **Bouclier de mana** augmente l'absorption de **Armure de glace**
- **Résonance magique** est une synergie globale : elle amplifie tous les sorts actifs des arbres I et II

---

## 8. Objets et inventaire

### Types d'objets

| Qualité | Couleur | Caractéristique |
|---------|---------|-----------------|
| Normal | Blanc | Aucun affix |
| Magique | Bleu | 1 préfixe + 1 suffixe |
| Rare | Jaune | 2 à 3 préfixes + 2 à 3 suffixes |
| Set | Vert | Bonus supplémentaires si set complet |
| Unique | Doré | Affixes fixes prédéfinis |
| Runeglyphe | Doré/Gris | Combinaison de glyphes dans des emplacements dédiés |

**Pour le Mage** (règle historique, ne s'applique plus à la classe Warrior — voir `devil_warrior_character_design.md` §4) : les objets sont orientés magie : orbes/cristaux, robes, amulettes, anneaux, grimoires, aucune arme physique empoignée (les turnarounds reçus montrent le Mage lançant ses sorts à mains nues, sans bâton). Aucune armure physique lourde, aucune arme de corps-à-corps non magique. L'itémisation du Warrior (sabres, cuir/métal léger) reste à formaliser.

### Stockage

Identique à Diablo 2 : stockage volontairement limité pour créer des choix de gestion.
- **Inventaire** : grille portée sur le personnage (taille contrainte)
- **Coffre (stash)** : accessible en ville uniquement
- **Ceinture** : jusqu'à 16 emplacements pour potions de mana et parchemins

### Cube de Nexus (remplace le Cube Horadrique)

Le Cube de Nexus permet de transformer et améliorer les objets. Son usage est introduit par une quête obligatoire en Région II. Les recettes disponibles sont **documentées progressivement in-game** via un Codex intégré : chaque recette découverte (par quête ou utilisation) est enregistrée et consultable à tout moment.

### Identification des objets

Les objets non identifiés créent un effet de suspense (slot machine). Le sage de chaque ville peut identifier en masse tout l'inventaire gratuitement.

### Portails de ville

Le Parchemin de Portail est l'un des deux objets de départ (avec le Parchemin d'Identification). Il permet un retour instantané en ville et peut être ouvert avant un combat de boss pour faciliter le retour après une mort.

---

## 9. Events temporaires de boost de drops

Pour compenser la rareté des objets de hauts TC sans supprimer le sentiment de récompense, Devil intègre un système d'Events :

| Type d'event | Déclencheur | Effet |
|--------------|-------------|-------|
| **Tempête de loot** | Aléatoire (durée limitée) | NoDrop divisé par 2 sur toute la session |
| **Boss corrompu** | Spawn rare dans une zone | Drops garantis de haute qualité |
| **Nuit de l'Apocalypse** | Event hebdomadaire | Multiplicateur x1.5 sur le Magic Find |
| **Quête d'urgence** | Déclenchée par la progression | Drop unique garanti à la complétion |

---

## 10. Respec progressif

Le Mage peut modifier son build selon trois méthodes, rendant les erreurs moins pénalisantes :

| Méthode | Accès | Portée |
|---------|-------|--------|
| **Respec partiel** | 1 fois par difficulté (quête Den of Nexus) | Tous les points de compétences et d'attributs |
| **Glyphe d'oubli** | Objet rare (drop ou craft) | 1 compétence ou 1 point d'attribut |
| **Respec complet** | Combinaison de 4 essences de boss dans le Cube de Nexus | Remise à zéro totale |

---

## 11. Philosophie de design

Devil hérite de la **Schaefer variation** identifiée dans Diablo 2 : la génération procédurale reproduit la courbe de difficulté montée/descente des jeux d'arcade dans un contexte RPG, créant un phénomène d'**acceleration flow** — le Mage alterne entre phases de domination et phases de danger, maintenant l'engagement sur la durée.

La suppression des classes multiples concentre toute la profondeur sur un seul archétype, poussé à son maximum : chaque arbre de compétences est une façon radicalement différente de jouer le même Mage.

---

## 12. Synthèse des ajustements par rapport à Diablo 2

| Élément | Diablo 2 | Devil |
|---------|----------|-------|
| Classes | 7 | 1 (Mage H/F) |
| Attribut offensif | Strength (physique) | Energy (magique) |
| Défense physique | Armor / Defense | Supprimée (résistances magiques uniquement) |
| Résistances négatives en fin de jeu | Oui (punitif) | Non (cap maintenu à 75%) |
| Recettes de craft | Non documentées in-game | Codex progressif in-game |
| Respec | Quasi-irréversible | Partiel / par objet / complet |
| Drop rates hauts TC | Très faibles, pas de filet | Faibles + Events temporaires |
| Mini-map | Optionnelle, incomplète | Toujours visible, temps réel |
| Multijoueur | Oui (Battle.net, 8 joueurs) | Solo uniquement |
| Stockage | Limité (intentionnel) | Limité (intentionnel, conservé) |

---

*Document de référence interne — non exhaustif.*
