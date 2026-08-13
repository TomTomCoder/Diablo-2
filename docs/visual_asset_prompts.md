# Devil — Prompts de génération d'assets visuels (style Final Fantasy)

Ce document fournit des prompts prêts à l'emploi pour un générateur d'images externe
(Midjourney, DALL-E, Stable Diffusion...). **Aucun outil de génération d'image n'est
disponible dans cet environnement de code** — ces prompts doivent être exécutés par
un humain, dans l'outil de son choix.

## Ce que ces prompts produisent — et ce qu'ils ne produisent pas

Un générateur d'images produit une **image fixe de concept art / key art**, pas une
planche de sprites jouable. Intégrer le résultat dans le moteur (Phase 6) reste un
travail séparé :

- Le moteur affiche des sprites isométriques multi-directionnels (8 à 16 orientations)
  au format `.DC6`, chaque état d'animation (idle/marche/incantation/impact/mort)
  ayant ses propres frames.
- Une image de concept art unique ne couvre qu'une orientation/pose. Il faudra soit
  faire repiquer/décliner l'art par un·e artiste 2D vers les frames isométriques,
  soit accepter un usage limité (icônes d'objets, portraits, artwork de chargement/menu
  — ces usages-là n'ont besoin que d'une seule image statique et sont directement
  exploitables).
- Pour les objets ramassables au sol et le Mage jouable, la conversion vers des frames
  isométriques animées + encodage `.DC6` reste un chantier de production distinct.

Autrement dit : ces prompts débloquent immédiatement les **icônes d'objets, portraits,
artwork de menu/chargement, et références de direction artistique** ; ils ne
débloquent pas à eux seuls le sprite du Mage jouable en jeu.

## Direction artistique commune (à coller au début de chaque prompt)

```
Final Fantasy-style painterly fantasy concept art, dramatic cinematic lighting,
semi-realistic proportions, ornate ceremonial armor design in the vein of
Tetsuya Nomura / Yoshitaka Amano, ambient occlusion, rich saturated color grading,
highly detailed, digital painting, no text, no watermark, no signature
```

Palette à respecter (`devil_mage_character_design.md` §2) :
- Cramoisi profond `#8b1a1a` (robe, dominante)
- Or ancien `#c8922a` (ornements, bordures, chaînes)
- Métal sombre `#2e2e38` (épaulières, plastron, gantelets)

---

## 1. Le Mage — variante Femme

```
[direction artistique commune]
Full-body character portrait of a tall, statuesque female battle-mage, aristocratic
cold posture, bare face with black hair in a strict bun, no helmet.
Asymmetric black engraved metal pauldrons (left shoulder larger than right).
Dark metal breastplate with gold inlay over a deep crimson (#8b1a1a) long gown with a
structured neckline, gown slit at the front revealing dark undergarment.
Wide burgundy red belt with a gold buckle. Right arm in a metal gauntlet, left arm
covered by the gown's sleeve. A golden decorative chain hangs loosely from the hip.
She holds a long two-handed staff of dark twisted wood topped with a worked crystal.
Cold blue stained-glass window light from behind, warm rim light on the gold trim.
Standing idle pose, staff held vertically.
```

## 2. Le Mage — variante Homme

```
[direction artistique commune]
Full-body character portrait of a tall, imposing male battle-mage, monolithic
powerful posture, bare face with short black hair, no helmet.
Symmetric massive gold-and-black metal pauldrons, larger than a female counterpart's.
Imposing dark metal breastplate with raised gold relief details, wide torso coverage.
Wide burgundy red belt with a gold buckle. Both forearms in metal gauntrets, heavier
coverage than a female counterpart. Deep crimson (#8b1a1a) long robe, straighter cut.
Same golden decorative hip chain. He holds a long two-handed staff of dark twisted
wood topped with worked metal.
Cold blue stained-glass window light from behind, warm rim light on the gold trim.
Standing idle pose, staff held vertically.
```

## 3. Progression visuelle de l'armure (5 paliers, `devil_mage_character_design.md` §6)

Ajouter une ligne parmi les suivantes au prompt de base (Femme ou Homme) selon le palier :

| Palier | Ligne à ajouter |
|---|---|
| Niveau 1–10 | `Plain simple robe, dull worn metal, staff of raw untreated wood, no shoulder armor yet.` |
| Niveau 11–25 | `Pauldrons now visible, golden ornamental trim beginning to appear on the robe hem.` |
| Niveau 26–40 | `Full breastplate, runes inlaid across the armor plates, staff now topped with a glowing crystal.` |
| Niveau 41–60 | `Entire suit of armor engraved with glowing runes, faint runic aura visible swirling around the staff.` |
| Niveau 61+ (endgame) | `Deep crimson robe with luminous embroidery, staff topped with two glowing crystals, permanent soft magical aura enveloping the whole body.` |

## 4. Auras de compétences (à superposer/mentionner selon l'arbre dominant, §7)

| Arbre | Ligne à ajouter |
|---|---|
| Élémentalisme | `Orange-red aura, swirling embers circling the hands.` |
| Arcane | `Pale cyan aura, glowing geometric shapes floating in the air around the character.` |
| Ésotérisme | `Deep violet aura, a translucent veil of dark magic enveloping the robe.` |

## 5. Icônes d'objets de départ (`devil_mage_character_design.md` §5)

Format recommandé pour chaque icône : carré, fond transparent ou neutre sombre, objet
seul en gros plan, même direction artistique commune, cadrage type icône d'inventaire.

- **Bâton de l'Apprenti** :
  `[direction artistique commune] Inventory icon of a simple apprentice's wooden
  staff, dark twisted wood, small worked metal cap, minor warm fire-glow ember
  wisps around the tip, square icon composition, dark neutral background.`
- **Robe du Novice** :
  `[direction artistique commune] Inventory icon of a simple crimson novice robe,
  plain cut, faint gold trim, folded/displayed flat, square icon composition, dark
  neutral background.`
- **Ceinture de Cuir Runique** :
  `[direction artistique commune] Inventory icon of a dark leather belt with four
  small rune-etched pouches for potions, gold buckle, square icon composition, dark
  neutral background.`
- **Pendentif Arcane** :
  `[direction artistique commune] Inventory icon of a small arcane pendant amulet,
  dark metal chain, a single glowing pale-blue rune stone at its center, square icon
  composition, dark neutral background.`
- **Anneau du Début** :
  `[direction artistique commune] Inventory icon of a simple gold band ring with a
  tiny crimson gem, square icon composition, dark neutral background.`

## 6. Qualités d'objets — cadres/effets visuels (`devil_game_design_reference.md` §8)

Pour un cadre d'icône ou un effet de halo à décliner par rareté (utiliser en fond ou
en bordure derrière chaque icône d'objet) :

| Qualité | Couleur | Ligne à ajouter |
|---|---|---|
| Normal | Blanc | `Plain thin white border frame, no glow.` |
| Magique | Bleu | `Soft blue magical glow around the item, thin blue border frame.` |
| Rare | Jaune | `Radiant yellow-gold glow around the item, ornate yellow border frame.` |
| Set | Vert | `Soft emerald green glow around the item, vine-like green border frame.` |
| Unique | Doré | `Intense golden divine glow around the item, elaborate gold filigree border frame.` |
| Runeglyphe | Doré/Gris | `Faint grey-gold shimmer, rune-carved stone border frame.` |

## 7. Monstres et gardiens de région (`devil_lore.md` §6)

- **Région I — Terres dévastées** (village natal de Devil, aujourd'hui ravagé) :
  `[direction artistique commune] A corrupted humanoid creature roaming a scorched,
  ash-covered ruined village, cracked dry earth, dead trees, ember-grey color
  palette, twisted charred limbs, faint embers glowing from within its chest.`
  - Boss **Gardien des Cendres** (un ancien villageois transformé par Devil) :
    `[direction artistique commune] A towering ash-grey guardian, once a human
    villager, now a hulking corrupted colossus made of cracked cinder and dying
    embers, faint humanoid facial features still visible beneath the ash-stone
    shell, glowing orange cracks across its body, mournful expression, boss
    key art, dramatic low-angle composition.`
- **Région II — Marécages corrompus** (base du Culte du Savoir, inondée par la magie) :
  `[direction artistique commune] A bloated swamp-corrupted creature emerging from
  murky flooded ruins, sickly green-purple bioluminescent magic seeping from cracked
  runic stonework half-submerged in the water, arcane cult sigils glowing faintly
  underwater.`
- **Région III — Forêt de cristal** (vie organique cristallisée par la magie) :
  `[direction artistique commune] A crystalline forest creature, its body partially
  fused with translucent glowing crystal growths, frozen mid-transformation, pale
  blue-white light refracting through crystallized trees in the background.`
- **Région IV — Citadelle du Néant** (forteresse personnelle de Devil avant sa
  corruption totale) :
  `[direction artistique commune] A void-touched armored sentinel patrolling a
  black obsidian fortress corridor, matte void-black armor with faint violet
  cracks of unraveling reality, unsettling geometric distortion in the background.`
- **Région V — Tour de Devil** (accès à la Source, épicentre de la corruption) :
  `[direction artistique commune] An apocalyptic tower interior at the epicenter of
  all magical corruption, reality visibly warping and bleeding raw magical energy
  in every color at once, silhouette of Devil's throne barely visible through the
  chaos, epic final-boss-arena key art.`

## 8. Artwork d'environnement par région (fond de menu / écran de chargement)

Réutiliser les mêmes cinq descriptions de région ci-dessus (§7) sans le sujet
créature, en vue large paysage, pour des artworks de chargement/menu :

```
[direction artistique commune] Wide landscape establishing shot of [description de
la région], no characters, atmospheric, painterly environment concept art.
```

---

## Prochaine étape

Une fois des images générées et validées par le studio, elles peuvent être :
1. Utilisées telles quelles pour les icônes d'objets (§5), les cadres de rareté (§6)
   et les artworks de menu/chargement (§8) — usage direct, aucune conversion requise.
2. Transmises à un·e artiste 2D pour découpage en frames isométriques + encodage
   `.DC6`, pour le Mage jouable (§1–4) et les monstres (§7) — travail de production
   séparé, voir `ROADMAP.md` Phase 6.
