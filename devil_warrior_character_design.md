# Devil — Character Design : Le Warrior

> **Référence visuelle** : turnaround IA (9 angles fixes + 8 frames de marche + 1 portrait) fourni par le studio, 2026-08-13.
> **Dernière mise à jour** : août 2026 — nouveau document, créé suite à la décision de faire de Devil un jeu **multi-classes** (voir `devil_game_design_reference.md`).
> **Statut : identité visuelle uniquement.** Aucune mécanique de jeu (statistiques, arbre de compétences, itémisation) n'existe encore pour le Warrior — contrairement au Mage, dont `devil_game_design_reference.md` §5-7 définit intégralement les attributs et les 30 compétences. Ce travail de design reste à faire par le studio ; ce document ne l'invente pas.

---

## 1. Concept général

Le Warrior est la deuxième classe jouable de Devil, aux côtés du Mage. Contrairement au Mage (magie pure, aucune arme physique), le Warrior est un combattant au corps-à-corps équipé d'armes physiques — un changement de philosophie de design par rapport à la règle initiale de Devil ("aucune arme de corps-à-corps non magique", `devil_game_design_reference.md` §8), qui ne s'applique donc plus qu'au Mage.

---

## 2. Identité visuelle

| Attribut | Description |
|----------|--------------|
| **Silhouette** | Humaine, athlétique, posture de duelliste plutôt que de tank lourd |
| **Cheveux** | Courts, gris/argentés |
| **Tenue** | Long manteau de cuir noir (duster), boutonné sur toute sa longueur, col ouvert laissant le torse partiellement visible |
| **Épaules** | Épaulières métal argenté ornées (motifs style crânes/rivets), la seule pièce d'armure visible — le reste de la tenue est du cuir souple |
| **Armes** | Deux sabres/katanas fins à lame courbe, portés à la ceinture, dégainés en combat |
| **Mains** | Gants de cuir noir |
| **Bas du corps** | Pantalon sombre, hautes bottes de cuir noir à boucles |
| **Rendu** | Style peinture semi-réaliste sombre, faible luminosité ambiante, contraste marqué — plus terreux/gothique que la palette lumineuse du Mage |

### Personnalité visuelle

Discipliné, silencieux, dangereux — un duelliste plutôt qu'un guerrier brutal. L'armure minimale (juste les épaulières) et le manteau long suggèrent la mobilité et la précision plutôt que l'encaissement de dégâts.

---

## 3. Références d'animation disponibles

- **9 rotations statiques** : Front, Front 3/4, Side, Back 3/4, Back, Back 3/4, Side, Front 3/4, Front
- **8 frames de marche** (un seul jeu, contrairement aux deux jeux du Mage — logique, il n'y a pas d'orbes flottants dont l'orientation change la lisibilité du sprite)
- **1 portrait** rapproché, épée dégainée, éclairage bas et dramatique — utilisable tel quel pour un écran de sélection de personnage ou une vignette de menu

Même limite technique que le Mage (voir `devil_mage_character_design.md` §5 et `docs/visual_asset_prompts.md`) : rendu à hauteur d'œil, pas de caméra isométrique — utilisable directement pour du UI/portrait/menu, pas comme sprite de gameplay sans nouveau rendu ou repaint.

---

## 4. Ce qui reste à concevoir (hors de portée de ce document)

Cette section liste ce qu'un document équivalent à `devil_game_design_reference.md` devrait couvrir pour le Warrior, une fois que le studio aura défini son gameplay :

- **Attributs** : Strength redevient probablement pertinent (contrairement au Mage, où `devil_game_design_reference.md` §5 le supprime explicitly) — à confirmer, ainsi que le rôle exact d'Energy/Dexterity/Vitality pour cette classe.
- **Arbre de compétences** : un ou plusieurs arbres dédiés au Warrior (mêlée, esquive, buffs de combat ?) — sans lien avec les 30 compétences magiques du Mage.
- **Itémisation** : armes physiques (sabres, autres ?), armure de cuir/métal léger — une rupture avec la règle "orienté magie" de `devil_game_design_reference.md` §8, à formaliser plutôt qu'à laisser en contradiction implicite.
- **Formule de dégâts** : `resolveAttackDamage` (`d2networking/d2server/game_server.go`) est actuellement câblé uniquement sur la formule magique du Mage (`base_sort × (1 + Energy / 100)`) — le Warrior aura besoin de sa propre formule (dégâts d'arme × Strength, par exemple) une fois ses mécaniques définies.
- **Écran de création de personnage** : `devil_game_design_reference.md` §2 ne décrit qu'un choix cosmétique de genre pour une classe unique — à réécrire pour un vrai choix de classe (Mage / Warrior).

---

*Voir `devil_mage_character_design.md` pour l'identité visuelle du Mage.*
*Voir `devil_game_design_reference.md` pour les mécaniques de jeu (mise à jour multi-classes en cours — le Warrior n'y est pas encore intégré).*
