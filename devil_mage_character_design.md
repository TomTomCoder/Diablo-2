# Devil — Character Design : Le Mage

> **Référence visuelle** : image officielle des deux variantes du Mage (femme / homme)
> **Dernière mise à jour** : mai 2026

---

## 1. Concept général

Le Mage est l'unique classe jouable de Devil. Le joueur choisit sa variante (femme ou homme) en début de partie — choix purement cosmétique, sans impact sur les statistiques ni les compétences. Les deux variantes partagent la même silhouette de base, le même équipement de départ et les mêmes animations d'incantation.

---

## 2. Identité visuelle commune

| Attribut | Description |
|----------|-------------|
| **Silhouette** | Grand, élancé, posture altière. Corps partiellement recouvert d'armure lourde sur robe longue. |
| **Couleur dominante** | Cramoisi profond (`#8b1a1a`) — robe et jupe |
| **Couleur secondaire** | Or ancien (`#c8922a`) — ornements, bordures, chaînes |
| **Couleur tertiaire** | Métal sombre (`#2e2e38`) — épaulières, plastron, gantelets |
| **Arme** | Bâton long à deux mains, bois sombre torsadé, sommet avec cristal ou métal ouvragé |
| **Lumière ambiante** | Fenêtres à vitraux bleutés en arrière-plan — teinte froide sur les épaules et la nuque |

---

## 3. Variante Femme

### Costume

| Zone | Description |
|------|-------------|
| **Tête** | Visage découvert, cheveux noirs en chignon strict. Aucun casque. |
| **Épaules** | Épaulières en métal noir gravé, format asymétrique — gauche plus large que droite |
| **Buste** | Plastron métal sombre avec détails en or. Décolleté structuré sur robe cramoisi. |
| **Ceinture** | Large ceinture rouge bordeaux, boucle centrale en métal doré |
| **Jupe / bas** | Robe longue cramoisi, fendue sur le devant, laissant apparaître un sous-vêtement sombre |
| **Bras** | Gantelet métal droit, bras gauche recouvert par la manche de robe |
| **Chaînes** | Chaîne décorative dorée sur la hanche, tombant librement |

### Personnalité visuelle

Froide, calculée, aristocratique. La symétrie imparfaite (épaulières asymétriques) renforce un sentiment de puissance naturelle plutôt qu'affichée. La robe longue au sol évoque la maîtrise, pas le combat physique.

---

## 4. Variante Homme

### Costume

| Zone | Description |
|------|-------------|
| **Tête** | Visage découvert, cheveux noirs courts. Aucun casque. |
| **Épaules** | Épaulières en métal doré-noir, plus massives que la variante femme, symétriques |
| **Buste** | Plastron métal plus imposant, détails dorés en relief. Couverture plus large du torse. |
| **Ceinture** | Ceinture rouge bordeaux identique, boucle centrale dorée |
| **Jupe / bas** | Robe longue cramoisi identique en couleur, coupe légèrement plus droite |
| **Bras** | Gantelets métal aux deux bras, avant-bras plus couverts |
| **Chaînes** | Même chaîne décorative dorée sur la hanche |

### Personnalité visuelle

Plus monolithique, imposant. L'armor coverage plus élevée renforce une lecture de puissance physique — mais reste fondamentalement un mage : la robe longue et le bâton dominent la silhouette.

---

## 5. Équipement de départ (loot niveau 1)

| Slot | Objet | Statistiques de base |
|------|-------|----------------------|
| Bâton | Bâton de l'Apprenti | +5 Energy, +10% dégâts Feu |
| Robe | Robe du Novice | +15 Vie, +5 Mana |
| Ceinture | Ceinture de Cuir Runique | 4 emplacements potions |
| Amulette | Pendentif Arcane | +3 Energy |
| Anneau | Anneau du Début | +2 à tous les attributs |

---

## 6. Progression visuelle du costume

L'apparence du Mage évolue avec les objets équipés. Les grandes étapes visuelles :

| Palier de niveau | Changement visuel attendu |
|-----------------|---------------------------|
| Niveau 1–10 | Robe simple, métal terne, bâton en bois brut |
| Niveau 11–25 | Épaulières apparaissent, ornements dorés sur la robe |
| Niveau 26–40 | Plastron complet, runes incrustées sur l'armure, bâton avec cristal |
| Niveau 41–60 | Armure entière gravée, aura runique visible autour du bâton |
| Niveau 61+ (endgame) | Robe cramoisi profond avec broderies lumineuses, bâton à deux cristaux, aura permanente |

---

## 7. Aura du Mage (rendu en jeu)

L'aura visible autour du personnage varie selon l'arbre de compétences dominant :

| Arbre dominant | Couleur de l'aura | Description |
|----------------|-------------------|-------------|
| Élémentalisme | Orange-rouge | Braises tournoyant autour des mains |
| Arcane | Cyan pâle | Géométries lumineuses flottant dans l'air |
| Ésotérisme | Violet profond | Voile translucide enveloppant la robe |

---

## 8. Animation — états principaux

| État | Description |
|------|-------------|
| **Idle** | Respiration lente, bâton tenu verticalement. Léger flottement de la robe. |
| **Déplacement** | Marche fluide, robe traînant au sol. Pas de course — le Mage se déplace avec lenteur calculée (sauf Téléportation). |
| **Incantation** | Main libre levée, particules de lumière convergent vers la paume. Durée liée au breakpoint de Dexterity. |
| **Impact (hit)** | Recul d'une demi-dalle, animation de douleur brève. Bâton ne lâche pas. |
| **Mort** | S'effondre lentement, robe se déployant au sol. Le bâton tombe en dernier. |

---

*Voir `devil_game_design_reference.md` pour les mécaniques de stats et d'arbres de compétences.*
*Voir `devil_dungeon_environment_gdd.md` pour l'aura du Mage et les règles de visibilité.*
