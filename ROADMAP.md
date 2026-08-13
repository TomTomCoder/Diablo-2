# Devil — Feuille de route

De « ça compile » à « on y joue » : plan pour construire **Devil** sur le moteur OpenDiablo2 — architecture, priorités, et ce qu'on ne fait volontairement pas tout de suite.

> Ce document a été réactualisé après lecture de `devil_lore.md`, `devil_game_design_reference.md` et `devil_mage_character_design.md`. Devil n'est pas "Diablo 2 avec un autre skin" : classe unique, stats et formule de dégâts entièrement magiques, arbres de compétences propres, lore propre. OpenDiablo2 reste le bon choix de **moteur** (rendu, chargement d'assets, pipeline de données, réseau local client/serveur) mais son **contenu** (skills.txt, monstats.txt, items.txt tels que fournis par Blizzard) ne correspond plus à ce qu'on construit — il faudra le remplacer par du contenu propre à Devil, pas le réutiliser tel quel.

## Décision de départ

On construit sur **OpenDiablo2** (Go/ebiten) comme moteur technique, pas sur AbyssEngine (trop tôt, aucune logique de jeu) et pas depuis zéro (le pipeline DC6/DCC/COF, le chargeur MPQ, le générateur de carte et le système de tables de données `d2records` sont directement réutilisables). Le contenu — classe, compétences, monstres, objets, lore — est spécifique à Devil et se construit par-dessus.

**Cible concrète de fin de parcours** : la **Région I** (Terres dévastées) jouable de bout en bout avec la classe Mage — créer un personnage, lancer des sorts, looter et équiper des objets magiques, monter au moins l'arbre Élémentalisme, vaincre le Gardien des Cendres, entendre les dialogues de Sage Wyn, sauvegarder et recharger. Pas les 5 régions tout de suite.

## Ce qui change par rapport au plan précédent (générique D2)

| Élément | Ancien plan (D2 générique) | Plan actualisé (Devil) |
|---------|----------------------------|-------------------------|
| Classes | Plusieurs, réutiliser celles de D2 | Une seule : le Mage (H/F cosmétique) |
| Attributs | Strength/Dexterity/Vitality/Energy (D2) | Energy/Vitality/Dexterity — **pas de Strength** |
| Dégâts | Arme physique équipée (déjà câblé, Phase 1) | Formule magique liée à l'Energy — **à corriger** |
| Mitigation | Armor/Defense physique | Résistances élémentaires uniquement, cap 75%, jamais négatif |
| Compétences | "Un arbre par classe" (vague) | 30 compétences précises / 3 arbres / synergies documentées |
| Objets | Réutiliser items.txt tel quel | Objets magiques uniquement (bâtons, orbes, robes...) — items.txt à retravailler |
| Monstres/boss | Génériques D2 | Noms et lore propres (Gardien des Cendres, La Noyée...) |
| Difficultés | Normal/Nightmare/Hell | Éveil/Corruption/Apocalypse (mêmes mécaniques, autre nommage) |

## Phase 0 — Remise en état des fondations ✅
*Fait — commit local `597702f`*

Avant d'ajouter du gameplay, s'assurer que ce qui existe ne casse pas silencieusement. Ce travail est indépendant du contenu du jeu — rien à revoir ici.

- ✅ `github.com/hajimehoshi/ebiten/v2` monté de v2.0.2 (2021) à v2.9.9 ; go.mod passé à Go 1.24. Build/vet/test complets vérifiés propres avec Go 1.26.5.
- ✅ Corrigé plusieurs vrais bugs révélés par `go vet` (format strings non constantes, canal `os.Signal` non bufferisé) — bloquants pour `go test`.
- ✅ Corrigé `TestRandFunction` (`rand.Seed()` devenu no-op sous Go 1.24+).
- ✅ CI GitHub Actions : build + vet + test + `golangci-lint` (v1.64.2, `only-new-issues` pour ne pas bloquer sur les centaines de violations `wsl` préexistantes).
- ✅ Test de fumée (`d2core/d2asset/d2asset_smoke_test.go`).

## Phase 1 — Combat et attributs du Mage 🚧
*Cinq items sur six faits et testés. Ne reste que la sauvegarde (limite connue ci-dessous).*

**Fait :**
- Une tranche verticale de combat de bout en bout — un monstre tuable a des PV, un cast de compétence proche de lui déclenche une résolution de dégâts côté serveur, le client reçoit et applique la mise à jour (PV ou suppression à la mort). Commits `0c9f823`, `1ebdbd3`.
- ✅ **Corrigé** : `resolveAttackDamage` (`d2networking/d2server/game_server.go`) utilisait les dégâts d'une arme physique équipée — ne correspondait pas au design (pas de Strength, pas d'arme de corps-à-corps non magique). Réécrit avec la vraie formule : `Dégâts = base_sort × (1 + Energy / 100)`, où `Energy` vient de `HeroState.Stats` et `base_sort` est un placeholder à plat (`baseSortDamage`) en attendant de vraies données de compétence (Phase 2). Commit `e5cc5d82`.

- ✅ **Trait de feu** a maintenant son propre `base_sort` (6, contre le placeholder à plat de 4) via une petite table `skillBaseSortDamage` indexée par un ID de compétence propre à Devil — délibérément numéroté loin de la plage des vrais ID de skills.txt de D2 (chargés à l'exécution depuis les MPQ du joueur, jamais présents dans ce dépôt) pour ne jamais entrer en collision. Commit `498d1fc0`.
- ✅ **Cooldown de cast lié à la Dexterity** : `canCastNow`/`castCooldownFor` — 800ms à 0 Dexterity, réduit progressivement, plancher à 200ms. Ferme un vrai trou (rien n'empêchait de spammer des `CastSkill` avant ça) et donne un premier effet réel à la Dexterity. Échelle continue en attendant les vrais paliers ("breakpoints") du design, qui demandent des données d'animation qui n'existent pas encore. Commit `d2474efc`.
- ✅ **Coût en mana + régénération** : `manaCostFor`/`manaRegenPerSecond` — chaque cast coûte du mana (3 pour Trait de feu, 2 par défaut), la régénération dépend de l'Energy (dernière stat dérivée manquante du §6, avec les dégâts et la vitesse d'incantation). `canCastNow` bloque désormais un cast si le mana est insuffisant, comme il bloque déjà sur le cooldown. Commit `6409c61f`.
- ✅ **Trait de feu castable en jeu** : chaque personnage créé reçoit maintenant `d2hero.NewTraitDeFeuSkill()` comme `LeftSkill` par défaut (`select_hero_class.go`). En creusant l'assignation, plusieurs points de crash latents ont été trouvés — le rendu HUD et le déclenchement de cast déréférençaient `LeftSkill`/`RightSkill` sans jamais vérifier le nil, pas seulement pour Devil, pour n'importe quelle compétence non résolue. Corrigé (`hud.go`, `game_controls.go`), avec un test qui garantit que `NewTraitDeFeuSkill()` ne peut jamais reproduire ce crash. Commit `425d3d13`.

**⚠️ Limite connue :** la sauvegarde ne survit pas encore à un rechargement — `HeroSkill.UnmarshalJSON` ne restaure que l'ID (`Shallow`), et la résolution qui reconstruit ensuite la compétence complète ne connaît que les skills.txt de D2. Un personnage fraîchement créé fonctionne pour la session en cours ; ça se recoupe avec l'item "Sauvegarde" ci-dessous.

**Reste à faire :**
- ✅ **Modificateurs d'équipement** : `d2core/d2hero/devil_items.go` — le Bâton de l'Apprenti (`+10% dégâts Feu`) a maintenant un vrai code d'objet Devil, assigné comme arme de départ à la création de personnage, et son modificateur s'applique dans `resolveAttackDamage` par-dessus la mise à l'échelle par Energy. Commit `3578d906`.
- ✅ **Résistances** : `FireResist`/`ColdResist`/`LightningResist`/`ShadowResist` sur `HeroStatsState`, `CapResistance` (clampe à [0, 75], jamais négatif contrairement à D2) et `MitigateDamage` (réduction en %). `tryMonsterAttack` mitige déjà ses dégâts via `FireResist` (tous les monstres traités comme dégâts de Feu pour l'instant, faute de données élémentaires par monstre). Commit `d0c46288`.
- ✅ **Bouclier de mana** : `HeroStatsState.ManaShieldActive` + `ApplyDamageWithManaShield` — draine le mana 1:1 avant de toucher la Health, le surplus retombe sur la Health normalement. Pas encore relié à un vrai sort/toggle (viendra avec la vraie compétence Ésotérisme en Phase 2) — activé directement pour l'instant. Commit `5b5f3946`.
- **Sauvegarde** : format JSON maison pour commencer (pas la compatibilité `.d2s`, hors sujet puisque les stats/formats de personnage ne sont plus ceux de D2) — et voir la limite connue ci-dessus pour la résolution de compétence au chargement.

**Phase 1 quasi close** : sur les six items listés au départ (dégâts, cooldown, mana, résistances, bouclier de mana, sauvegarde), quatre sont faits et testés. Restent les modificateurs d'équipement (bloqués par la même limite que *Trait de feu* : pas d'assignation d'objets Devil à la création de personnage) et la sauvegarde (bloquée par la même limite de résolution de compétence au chargement). Les deux vraies limites restantes convergent vers un seul sujet : **la persistance des données Devil (compétences et objets) au chargement d'une sauvegarde**, pas deux chantiers séparés.

**Simplifications volontaires qui restent acceptables pour l'instant** (documentées en commentaire `ponytail:` dans le code) : ciblage par proximité au lieu de clic-sur-cible, pas de jet de précision, une seule cible par sort (pas de zone d'effet même pour les sorts qui devraient en avoir un).

## Phase 2 — Arbres de compétences du Mage 🚧
*Démarrée — 7/10 sorts d'Élémentalisme, 6/10 Arcane.*

Le design définit **30 compétences en 3 arbres de 10** (`devil_game_design_reference.md` §7) : Élémentalisme (dégâts directs), Arcane (contrôle/amplification), Ésotérisme (défense/survie/invocations), avec synergies inter-arbres et un système de palier (`niveau requis`) qui débloque les compétences supérieures à condition d'avoir investi dans les inférieures.

- ✅ **Modèle de données** : `d2core/d2hero/devil_skill_tree.go` — `DevilSkillDef` (arbre, palier, `base_sort`, coût en mana, cible/pourcentage de synergie) et le registre `DevilSkills`, plus `CanLearnSkill` pour le palier. Remplace les deux maps ad-hoc de `game_server.go` par une seule source de vérité. Commit `1f1275c0`.
- ✅ **Éclat de glace** (2e sort d'Élémentalisme) : `base_sort` plus faible que Trait de feu (4 contre 6, coût en mana plus élevé) mais un vrai effet de ralentissement — `NPC.ApplySlow`/`IsSlowed`, `ChasePlayer` réduit sa vitesse de moitié pendant l'effet. Rendu possible par l'IA de monstre de la Phase 4 (avant ça, rien n'aurait consommé un ralentissement de façon significative). Commit `43c26f4a`.
- ✅ **Éclair en chaîne** (3e sort, palier 6 — premier sort qui exploite réellement `CanLearnSkill`) : dégâts sur jusqu'à 3 cibles. A demandé de sortir `resolveMeleeHit` de son modèle mono-cible — `nearestKillableNPC` (recherche avec exclusion) et `applyHit` (dégâts/mort/or/ralentissement/diffusion) sont maintenant partagés entre le coup simple et `resolveChainHit`. **Limite de test connue** : les champs de `NPC` sont privés hors de `d2mapentity`, donc l'itération multi-cible n'est pas testée unitairement au niveau du serveur (seules les données de compétence le sont) — documenté dans le commit plutôt que masqué. Commit `9c53f951`.
- ✅ **`HeroState.LearnSkill`** : relie enfin le modèle de données à la vraie logique de jeu — dépense un point de compétence pour apprendre n'importe quel sort du registre `DevilSkills` (pas seulement Trait de feu), avec toutes les vérifications (palier via `CanLearnSkill`, points disponibles, pas déjà connu). `NewTraitDeFeuSkill` généralisé en `NewDevilHeroSkill(skillID)`. Commit `4a85e2f7`.
- ✅ **Système d'XP/niveau** (`devil_game_design_reference.md` §5 : 5 points d'attribut + 1 point de compétence par niveau) : `HeroStatsState.GrantExperience` fait monter de niveau (potentiellement plusieurs fois d'un coup) tant que `Experience` dépasse `NextLevelExp`, courbe placeholder à plat (`experienceForLevel = niveau × 100`, aucune vraie table Devil n'existe). `d2hero.RollExperienceDrop` (10-30 XP, même forme que `RollGoldDrop`) est distribué à la mort d'un monstre via `GameServer.awardExperience`, diffusé au client par le nouveau paquet `ExperienceAwardedPacket`. Donne enfin à `LearnSkill` une vraie source de points de compétence — reste un trou documenté : aucune UI/requête réseau ne permet encore à un joueur d'invoquer `LearnSkill` lui-même. Commit `c6e1d818`.
- ✅ **Nova de givre** (4e sort, palier 6 — même palier qu'Éclair en chaîne) : première compétence centrée sur le lanceur plutôt que sur une cible/position visée — touche tous les monstres tuables dans un rayon autour du Mage. `killableNPCsWithin` généralise `nearestKillableNPC` (collecte tous les monstres en portée au lieu du seul plus proche). Même limite de test que Éclair en chaîne : le trajet multi-cible réel n'est testé qu'au niveau des données/no-op, pas de bout en bout côté serveur. Commit `8c3855ff`.
- ✅ **Boule de feu** (5e sort, palier 3/niveau 12) : AoE centrée sur la position visée plutôt que sur le lanceur — `resolveNovaHit` a été factorisé en `resolveAoeHit(centre, source, sortID, rayon)`, partagé par Nova de givre (centre = lanceur) et Boule de feu (centre = cible visée). Plus haut `base_sort` de tous les sorts d'Élémentalisme à ce stade, conforme à "dégâts feu élevés". Commit `05ad1b47`.
- ✅ **Tempête statique** (6e sort, palier 3, aux côtés de Boule de feu) : "zone d'éclair persistante" — pour l'instant résolue comme un simple coup AoE instantané sur la position visée (même chemin `resolveAoeHit` que Boule de feu), la partie "persistante" n'est pas modélisée (aucune entité zone/durée n'existe) — commentaire `ponytail` avec la voie de montée en gamme (une expiration par lancer suivie comme `lastCastAt`, re-déclenchant `resolveAoeHit` à chaque `aiTickInterval`). Le if unique de Boule de feu dans `resolveMeleeHit` est devenu `aoeAtTargetRadiusSubtiles`, une map sortID→rayon, pour ne pas empiler les if à mesure que d'autres sorts AoE-sur-cible arrivent (Orbe glaciale, Météore, Apocalypse sont du même type). Commit `c2dc45f2`.
- ✅ **Champ statique** (1er sort d'Arcane, palier 1) : "Réduit la vie de toutes les entités à l'écran d'un % fixe" — premier sort dont les dégâts ne suivent pas la formule Energy de `resolveAttackDamage` (dégâts = % de la vie actuelle de chaque monstre). "à l'écran" est modélisé comme tous les monstres tuables de la carte, faute de notion de champ de vision par client — documenté en `ponytail`. `applyHit` a été scindé en `applyHit` (dégâts Energy, inchangé) + `applyResolvedDamage` (mort/or/XP/ralentissement/diffusion, prend un montant déjà calculé) pour que Champ statique partage cette plomberie sans la dupliquer. Commit `47b74b00`.
- ✅ **Télékinésie** (2e sort d'Arcane, palier 1) : "Repousse les entités, interaction avec les objets à distance" — seule la moitié "repousse" est modélisée (la partie interaction à distance avec les objets demande le système d'objets de la Phase 5). Nouvelle méthode `d2mapentity.NPC.Knockback(source, distance)` : téléportation instantanée le long de la droite source→PNJ, sans physique/animation/collision réelles. **Première compétence multi-cible réellement testée de bout en bout** (pas seulement au niveau des données) : les tests du package `d2mapentity` peuvent construire un vrai `NPC` via `newMapEntity`, contrairement aux tests côté serveur qui ont besoin de la factory liée aux MPQ. **Limite documentée** : aucun paquet `NPCMoved`/diffusion de position n'existe encore, donc le recul est autoritatif côté serveur mais invisible pour les clients tant que ce paquet n'existe pas. Commit `cbcd5f02`.
- ✅ **Ralentissement** (3e sort d'Arcane, palier 2) : "Zone qui réduit la vitesse des entités de 50%" — sans dégâts, réutilise le mécanisme `ApplySlow`/`npcSlowedSpeedMultiplier` déjà en place pour Éclat de glace, mais sur toute une zone (`killableNPCsWithin`) au lieu d'une cible unique. Même limite documentée que le ralentissement d'Éclat de glace : aucun paquet ne diffuse encore `SlowedUntil` aux clients. Commit `2b38a11a`.
- ✅ **Amplification** (4e sort d'Arcane, palier 2, aux côtés de Ralentissement) : "Augmente les dégâts magiques reçus par la cible" — debuff mono-cible sans dégâts propres. `d2mapentity.NPC` gagne `AmplifiedUntil`/`ApplyAmplification`/`IsAmplified`, copie conforme de `SlowedUntil`/`ApplySlow`/`IsSlowed`. `applyResolvedDamage` fait désormais passer les dégâts par `amplifiedDamage(dégâts, amplifié)` (+50%, `amplificationDamagePercent = 150`, placeholder en attendant de vrais chiffres d'équilibrage) avant de les appliquer — fonction pure extraite exprès pour être testable sans PNJ réel. Commit `891b179c`.
- ✅ **Téléportation** (5e sort d'Arcane, palier 3/niveau 12) : "Déplacement instantané vers la position visée" — déplace le lanceur lui-même, pas un PNJ. Nouveau paquet `PlayerTeleportedPacket`, volontairement distinct de `MovePlayerPacket` (qui fait cheminer une marche fluide côté client) : `handlePlayerTeleportedPacket` fixe directement `Position` via le champ promu `d2vector.Position.Set`, sans cheminement. Première compétence de ce lot testée avec un vrai test comportemental côté serveur (pas seulement un no-op) : elle ne touche que `HeroState`, aucun PNJ nécessaire. Commit `e0791e2f`.
- ✅ **Orbe glaciale** (7e sort d'Élémentalisme, palier 4/niveau 18) : "Projectile lent, explose en large AoE de froid" — même forme AoE-sur-cible que Boule de feu/Tempête statique, simplement enregistrée dans `aoeAtTargetRadiusSubtiles` avec le plus grand rayon des trois (6, contre 3/4), au prix de dégâts plus faibles (même compromis que Nova de givre face à Trait de feu). Commit `f034a9e9`.
- ✅ **Prison de glace** (6e sort d'Arcane, palier 4/niveau 18) : "Immobilise un groupe d'entités" — immobilisation totale plutôt que le ralentissement partiel de Ralentissement. `d2mapentity.NPC` gagne `ImmobilizedUntil`/`ApplyImmobilize`/`IsImmobilized` (même forme que `SlowedUntil`) ; `ChasePlayer` vérifie `IsImmobilized` en premier et force la vitesse à 0, l'immobilisation l'emportant totalement sur un ralentissement simultané plutôt que de s'y cumuler. Même limite documentée que Ralentissement : aucun paquet ne diffuse encore `ImmobilizedUntil` aux clients. Commit `7665b051`.
- Reste 3 sorts d'Élémentalisme (Maîtrise élémentaire, Météore, Apocalypse), 4 sorts d'Arcane, puis Ésotérisme ensuite — Ésotérisme demande des invocations (Familier, Golem arcane) qui réutilisent le système `d2mapentity.NPC` mais côté allié, pas ennemi.
- Respec : implémenter les 3 méthodes du design (respec partiel par quête, glyphe d'oubli ciblé, respec complet via essences de boss) une fois qu'il y a des builds à corriger.

**Pourquoi Élémentalisme en premier :** c'est l'arbre qui valide le plus directement la formule de dégâts déjà posée en Phase 1, sans dépendre de mécaniques qui n'existent pas encore (contrôle de groupe, invocations alliées).

## Phase 3 — Scripting moderne ✅
*Fait — commit `5c544c9c`.*

- ✅ `d2script` utilisait `otto` (interpréteur JS pur Go sans activité depuis des années). Remplacé par `wazero` (runtime WASM pur Go, zéro cgo, activement maintenu) : `ScriptEngine` charge et exécute des modules `.wasm` compilés à l'avance plutôt que d'interpréter du texte JS.
- ✅ `Eval`/`AllowEval`/`DisallowEval` supprimés entièrement — pas d'équivalent WASM ("évaluer une chaîne de code arbitraire" n'existe pas dans ce modèle), et c'est une vraie amélioration de sécurité : un module WASM ne peut appeler que les fonctions hôte qui lui sont explicitement données, jamais un accès arbitraire au process comme avec `Eval()`.
- ✅ Test de bout en bout réel (pas juste la plomberie à vide) : le toolchain Go standard compile directement en WASM (`GOOS=wasip1 GOARCH=wasm go build`, sans tinygo/clang) — un vrai module invité compilé (`d2script/testdata/hello.wasm`) importe et appelle une fonction hôte, le test vérifie que la valeur passée arrive bien côté Go.
- La compétence `getMapEngines` (exposait tout le slice `[]*MapEngine` via le passage de valeur arbitraire d'otto) ne se traduit pas en WASM — un module invité ne peut pas garder une référence Go vivante. Remplacée par un placeholder minimal (`mapEngineCount`, retourne un compte) qui prouve juste que le mécanisme fonctionne. Rien n'exploite ce point d'accroche aujourd'hui de toute façon.
- Reste à faire : c'est le mécanisme qui portera la **narration** (Phase 4) — déclencher les dialogues de Sage Wyn, la révélation finale, les quêtes d'urgence des Events temporaires (§9) — mais l'API de script réelle (quelles fonctions hôte exposer, quel format de module invité pour une "quête") n'est pas encore conçue, seulement le mécanisme de base.

## Phase 4 — Contenu narratif & Région I
*Après qu'un sort fonctionne (Phase 1) et que le scripting moderne soit en place (Phase 3).*

Construit la première tranche de contenu jouable racontée, sur la Région I (Terres dévastées — "village natal de Devil, aujourd'hui méconnaissable", `devil_lore.md` §6).

- ✅ **IA de monstre minimale** : premier tick périodique côté serveur (`runMonsterAILoop`, 200ms — le serveur était jusqu'ici purement réactif aux paquets, n'avançait jamais rien de lui-même). Un monstre tuable détecte le joueur connecté le plus proche, le poursuit en ligne droite (`NPC.ChasePlayer`, pas de vrai pathfinding autour des obstacles), et l'attaque à portée avec son propre cooldown. Nouveau paquet `PlayerDamaged` pour synchroniser les PV du joueur au client. Chiffres à plat pour tous les monstres (portée, dégâts, cooldown) — pas encore de données par monstre. Commit `9e8fac4c`.
- **Gardien des Cendres** comme boss de fin de région : un monstre unique (pas une simple instance de monstat.txt générique), avec ses propres PV/résistances/comportement — construit par-dessus l'IA minimale ci-dessus.
- **Dialogues de Sage Wyn** (`devil_lore.md` §7) au démarrage et avant la salle du boss — premiers contenus scriptés via le pipeline WASM de la Phase 3.
- **Narration à la mort du boss** — valide qu'un événement de gameplay (mort d'un monstre) peut déclencher un texte narratif, brique nécessaire pour la révélation finale du jeu plus tard.
- **Une quête complète de bout en bout** sur cette région, pour valider tout le pipeline narratif avant d'en écrire pour les 4 autres régions.
- **Génération procédurale améliorée** (`devil_game_design_reference.md` §4.1) : mini-map toujours visible et mise à jour en temps réel (actuellement la mini-map D2 est optionnelle/incomplète), tiles sans labyrinthes sans issue, icônes de points d'intérêt (boss, waypoints, coffres rares) dès leur découverte. Améliorations du moteur `d2mapgen`/`d2maprenderer`, pas juste du contenu.

## Phase 5 — Objets, loot & progression 🚧
*Démarrée.*

- ✅ **Modificateur d'arme** : voir Phase 1 (Bâton de l'Apprenti, `+10% dégâts Feu`). Commit `3578d906`.
- ✅ **Or à la mort d'un monstre** : `d2hero.RollGoldDrop` (5-20, plage à plat faute de données de Treasure Class par niveau de monstre), nouveau paquet `GoldAwarded`. Contourne le mur des objets qui tombent au sol (voir ci-dessous) : l'or n'a pas besoin de sprite pour exister, donc c'est un mécanisme réel de bout en bout plutôt qu'une plomberie qui ne produirait rien de visible. Commit `175070f8`.
- **⚠️ Limite structurelle des objets au sol** : faire apparaître un objet ramassable (`d2mapentity.Item`) exige un vrai sprite — `MapEntityFactory.NewItem` charge l'animation depuis `item.CommonRecord().FlippyFile`, un fichier DC6 réel. Aucun sprite d'objet Devil n'existe (même mur que l'identité visuelle du Mage, Phase 6). Tant que ça reste vrai, tout objet Devil ramassable au sol échouera à s'afficher (`NewItem` retourne une erreur, gérée sans crash côté client, mais rien n'apparaît). L'or et l'XP (Phase 2) — mécanismes sans dépendance visuelle — restent le terrain praticable de cette phase en attendant ; la réputation ou d'autres systèmes similaires suivraient le même chemin.
- **Objets magie-only** : bâtons, orbes, robes, amulettes, anneaux, grimoires — pas d'armure physique lourde, pas d'arme de corps-à-corps non magique (`devil_game_design_reference.md` §8). `items.txt` et le système d'équipement (`d2inventory.CharacterEquipment`) sont à retravailler avec ce nouveau set d'emplacements/types, pas juste réutilisés depuis D2.
- **Treasure Classes** : le concept (probabilités par TC, NoDrop, picks) est directement réutilisable depuis D2 — c'est un mécanisme, pas du contenu Blizzard.
- **Cube de Nexus** (remplace le Cube Horadrique) avec un **Codex progressif in-game** des recettes découvertes — amélioration sur D2 (recettes non documentées in-game), fonctionnalité UI à construire en plus de la logique de craft.
- **Respec** : voir Phase 2.
- **Events temporaires** (Tempête de loot, Boss corrompu, Nuit de l'Apocalypse, Quête d'urgence — §9) : système de scheduling/déclenchement, bon candidat pour le scripting WASM de la Phase 3 une fois en place.

## Phase 6 — Rendu, confort visuel & identité artistique
*Statut : filtrage des tuiles fait. Le reste dépend de production d'assets, pas seulement de code.*

- ✅ Filtrage linéaire sur les tuiles de sol/mur/ombre (`d2maprenderer`) — commit local `33f27a2`.
- **Identité visuelle du Mage** (`devil_mage_character_design.md`) : deux variantes (H/F), silhouette cramoisi/or/métal sombre, bâton à deux mains, aura colorée selon l'arbre de compétences dominant (orange-rouge Élémentalisme, cyan Arcane, violet Ésotérisme), 5 paliers de progression visuelle par niveau. **Aucun sprite existant ne correspond à ce design** — les sprites Sorceress/Necromancer de D2 ne sont pas un point de départ visuel valable (silhouette et costume différents). C'est un chantier de production d'assets (modélisation/pixel art + animation DC6/DCC), pas un problème d'ingénierie : à planifier comme un flux de travail séparé, en parallèle du code, pas comme une tâche de fin de phase.
- ✅ **Prompts de génération d'assets** (`docs/visual_asset_prompts.md`) : prompts prêts à l'emploi (style Final Fantasy, palette du design) pour un générateur d'images externe — Mage H/F par palier de progression, auras par arbre, icônes des 5 objets de départ, cadres par qualité (§8), monstres/gardiens des 5 régions et artworks de région. Aucun outil de génération d'image n'est disponible dans cet environnement de code : ces prompts sont à exécuter par un humain. Le document précise honnêtement leur portée réelle — icônes/portraits/artwork de menu sont directement exploitables, le sprite isométrique animé du Mage jouable reste un travail de production séparé (découpage en frames + encodage `.DC6`).
- Pipeline d'upscaling de sprites (xBRZ/HQx offline, métadonnées de taille cohérentes) — pertinent une fois qu'il y a de vrais sprites Devil à upscaler, pas seulement les sprites D2 d'origine.
- Post-traitement via les shaders Kage natifs d'ebiten (bloom, vignette) — bon candidat pour rendre les auras de compétences (voir plus haut) plutôt qu'un simple effet cosmétique.
- Résolution interne fixée à 800×600 — support HiDPI à traiter avec le reste.

## Phase 7 — Outillage de contenu 🚧
*Démarrée.*

- ✅ **Format de mod documenté** (`docs/modding.md`) : contrat hôte/invité, comment compiler un module Go avec le toolchain standard, la seule fonction hôte qui existe aujourd'hui, limites explicites (pas d'API quête/dialogue, types numériques/mémoire seulement, pas de sauvegarde côté script). Commit `4af83004`.
- **Hors de portée ici** : finir la communication `abysswrapper` ↔ moteur, et étendre les éditeurs HellSpawner — les deux exigent le dépôt `hellspawner`/`AbyssEngine` séparé et une interface graphique (giu/ImGui) qui a besoin d'un display, indisponibles dans cet environnement — même nature de mur que la production d'assets visuels.

## Phase 8 — Packaging & diffusion 🚧
*Démarrée.*

- ✅ **Build via `goreleaser`** (`.goreleaser.yml`, schéma v2) : vérifié de bout en bout — `goreleaser build --snapshot` produit un binaire réel, exécuté pour confirmer qu'il tourne. Scope honnête : macOS uniquement (amd64+arm64), vérifié. La compilation croisée CGO vers Linux échoue proprement depuis macOS sans toolchain C croisée (testé) — Windows aurait le même problème. La vraie solution (documentée dans le fichier) : un job CI par OS natif, pas de compilation croisée locale.
- ✅ **Rappel légal au démarrage** : `d2app.Create` affiche maintenant un message rappelant que le jeu nécessite les fichiers Diablo II + LoD achetés légalement par le joueur — vérifié en exécutant le binaire buildé par goreleaser. Commit `cc6d8289`.

## Ce qu'on ne fait pas maintenant

- Refonte 3D/PBR — incompatible avec des sprites 2D pré-rendus.
- Migrer vers AbyssEngine comme base — trop tôt, aucune logique de jeu à ce stade.
- Portage natif WebGPU/Vulkan — gain marginal pour le risque.
- Multijoueur en ligne — décision de scope confirmée par le game design (§1 : "action-RPG solo"). L'architecture interne d'OpenDiablo2 fait transiter même le solo par un client/serveur local ; on la laisse telle quelle.
- Les 5 régions avant d'avoir la Région I qui tient debout.
- Réutiliser tel quel le contenu D2 (skills.txt, monstats.txt, items.txt) — le design est trop différent (classe unique, magie-only, attributs différents) pour que ça vaille le coup ; seul le *pipeline* de chargement de ces formats est réutilisé.

## Risques à surveiller

- **Production d'assets visuels** : aucun sprite Devil n'existe encore, et les sprites D2 d'origine ne correspondent pas au design (voir Phase 6). C'est probablement le vrai goulot d'étranglement du projet, pas le code moteur.
- Solo ou petite équipe : la Phase 1 corrigée + Phase 2 (30 compétences) représentent plusieurs mois, pas des semaines.
- Le combat déjà commité (`0c9f823`, `1ebdbd3`) utilise une formule physique à corriger — ne pas construire d'autres systèmes par-dessus (IA, objets) avant d'avoir corrigé la formule de base, sous peine de devoir tout reprendre.
- wazero change la façon d'écrire des quêtes/dialogues ; prévoir un exemple de bout en bout (Sage Wyn) avant de scripter le reste du contenu narratif.

## Prochaine étape

Brancher *Trait de feu* comme premier sort réel (données de compétence + `base_sort` propre), pour valider la formule Energy sur un cas concret avant d'ouvrir davantage de compétences ou de contenu.
