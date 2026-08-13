# Devil — Character Design : Le Mage

> **Référence visuelle** : turnarounds IA (9 angles fixes + 16 frames de marche en 8 directions) fournis par le studio, 2026-08-13.
> **Dernière mise à jour** : août 2026 — **remplace intégralement la version précédente** (palette cramoisi/or/métal sombre, bâton à deux mains) suite à la réception des vrais assets visuels.
> **Statut** : Devil est désormais un jeu **multi-classes** (voir `devil_game_design_reference.md` — mise à jour en cours). Ce document couvre uniquement le Mage ; voir `devil_warrior_character_design.md` pour le Warrior.

---

## 1. Concept général

Le Mage est une des classes jouables de Devil (aux côtés du Warrior). Contrairement à la version précédente du design, le Mage n'a **pas de variantes homme/femme cosmétiques** dans les assets reçus : à la place, il existe **deux formes du même personnage**, une progression narrative/de puissance plutôt qu'un choix cosmétique à la création :

- **Mage** — forme de base
- **Mage Divin** — forme évoluée (voir §4)

> ⚠️ Hypothèse à valider : le studio a fourni 4 images sans étiquette explicite par personnage. L'attribution ci-dessous (Mage = image au look féral roux, Mage Divin = image au look immaculé argenté) est déduite de la logique visuelle (forme de base plus sauvage → forme évoluée plus divine/maîtrisée). À confirmer.

---

## 2. Identité visuelle commune

| Attribut | Description |
|----------|--------------|
| **Silhouette** | Longiligne, torse largement dénudé, posture ouverte (bras écartés tenant des orbes flottants plutôt qu'une arme empoignée) |
| **Incantation** | Aucun bâton/arme physique — le Mage lance ses sorts à mains nues, un objet magique flottant dans chaque paume (cristal, orbe élémentaire) |
| **Cheveux** | Longs et volumineux, élément visuel dominant de la silhouette (occupent une part importante du contour du personnage) |
| **Bas du corps** | Longue étoffe déchiquetée/en lambeaux tombant jusqu'aux pieds, mouvement fluide en marche |
| **Rendu** | Style peinture semi-réaliste, éclairage dramatique, cohérent avec la direction artistique "Final Fantasy" (`docs/visual_asset_prompts.md`) |

---

## 3. Mage (forme de base)

### Costume

| Zone | Description |
|------|-------------|
| **Tête** | Cheveux longs, volumineux et hérissés, rouge cramoisi vif |
| **Torse** | Torse nu et musclé, aucune armure |
| **Bas du corps** | Longue étoffe/fourrure effilochée blanc-crème, texture en lambeaux/plumes, tombant en pointes irrégulières jusqu'au sol |
| **Peau/jambes** | Teinte rougeâtre sur les jambes/pieds, aspect quasi félin/démoniaque — pieds griffus |
| **Mains** | Main gauche : cristal rouge anguleux entouré d'un anneau incandescent. Main droite : orbe blanc-argenté lumineux |

### Personnalité visuelle

Sauvage, primitif, à mi-chemin entre le mage et la créature élémentaire — une magie qui semble à peine contenue par le corps qui la porte. Aucune retenue aristocratique : c'est une force brute encore instable.

---

## 4. Mage Divin (forme évoluée)

### Costume

| Zone | Description |
|------|-------------|
| **Tête** | Cheveux longs et volumineux, blanc-argenté |
| **Torse** | Vêtement structuré violet sombre/noir avec liserés or, ventre découvert |
| **Bas du corps** | Longue tunique fendue blanc-crème à liserés dorés, superposée à un pantalon/bottes sombres avec plaques métalliques argentées |
| **Mains** | Main gauche : orbe de feu orange incandescent. Main droite : orbe de glace bleu-blanc |

### Personnalité visuelle

Maîtrisée, ascendante, "divine" — la même magie qu'à l'état de Mage, mais désormais dominée plutôt que subie. Double polarité élémentaire (feu/glace) dans une seule silhouette, contrairement à la forme de base qui n'en montre qu'une pointe (cristal) et une neutre (orbe blanc).

### Déclencheur de la transformation

Non spécifié dans les assets reçus — à définir : palier de niveau, quête, objet, ou état de jeu (ex. seuil de compétences investies). Voir `devil_game_design_reference.md` §6 pour la progression visuelle si elle est réintroduite sous une autre forme.

---

## 5. Références d'animation disponibles

Les deux formes disposent chacune de :

- **9 rotations statiques** : Front, Front 3/4, Side, Back 3/4, Back, Back 3/4, Side, Front 3/4, Front (couvre les 8 directions isométriques classiques + redondance de contrôle qualité)
- **16 frames de marche (8 directions × 2 jeux)** : un jeu orbe-avant (mains vers la caméra) et un jeu orbe-arrière (mains loin de la caméra), pour couvrir la marche dans les deux sens sans réutiliser un miroir simple

**Limite technique importante** (voir `docs/visual_asset_prompts.md` "Ce que ces prompts produisent — et ce qu'ils ne produisent pas") : ces rendus sont vus à hauteur d'œil (caméra de type écran de sélection de personnage), pas depuis la caméra isométrique top-down que le moteur utilise réellement en jeu. Les intégrer tels quels comme sprites `.DC6` produirait un rendu visuellement faux (proportions, point de contact au sol, raccourci perspectif tous calculés pour le mauvais angle de caméra). Utilisables directement pour : écran de sélection de personnage, artwork de chargement, portraits UI. Nécessitent un nouveau rendu (caméra isométrique) ou un repaint pour devenir de vrais sprites de gameplay.

---

*Voir `devil_game_design_reference.md` pour les mécaniques de stats et d'arbres de compétences (mise à jour multi-classes en cours).*
*Voir `devil_warrior_character_design.md` pour l'identité visuelle du Warrior.*
