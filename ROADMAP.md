# OpenDiablo2 — Feuille de route

De « ça compile » à « on y joue » : plan en six phases pour rendre OpenDiablo2 jouable en solo — architecture, priorités, et ce qu'on ne fait volontairement pas tout de suite.

## Décision de départ

On construit sur **OpenDiablo2** (Go/ebiten), pas sur AbyssEngine. AbyssEngine est l'architecture visée à terme par le projet historique, mais au moment de l'analyse elle n'avait qu'un menu principal et des vidéos d'intro — aucune logique de jeu. OpenDiablo2 fait déjà marcher un personnage dans l'Acte 1 : c'est la base la plus proche de « jouable ». Le clone local d'AbyssEngine a été supprimé (voir plus bas) ; il pourra être re-cloné si le projet en a de nouveau besoin (ex. prévisualisation live depuis HellSpawner via un binaire compilé séparément).

**Cible concrète de fin de parcours** : un Acte 1 solo complet — créer un personnage, combattre, looter et équiper, monter des compétences, finir une quête de bout en bout, sauvegarder et recharger. Pas la parité des 5 actes tout de suite.

## Phase 0 — Remise en état des fondations ✅
*Fait — commit local `597702f` sur la branche `worktree-sprite-upscale`*

Avant d'ajouter du gameplay, s'assurer que ce qui existe ne casse pas silencieusement.

- ✅ `github.com/hajimehoshi/ebiten/v2` monté de v2.0.2 (2021) à v2.9.9 ; go.mod passé à Go 1.24 (le plancher que réclame la nouvelle version d'ebiten). Build/vet/test complets vérifiés propres avec Go 1.26.5.
- ✅ Corrigé plusieurs vrais bugs révélés par `go vet` (format strings non constantes dans `d2term`/`d2gamescreen`/`d2player`, canal `os.Signal` non bufferisé dans `d2app`) — bloquants pour `go test` depuis que le vet par défaut inclut ces checks.
- ✅ Corrigé `TestRandFunction` : le test s'appuyait sur `rand.Seed()` pour rendre deux évaluations reproductibles — depuis Go 1.24, `rand.Seed()` est un no-op documenté dès qu'un module cible `go 1.24+`. Réécrit pour vérifier le vrai contrat de `rand(v1,v2)` (retourne toujours l'une des deux bornes).
- ✅ CI GitHub Actions (`.github/workflows/ci.yml`) : build + vet + test + `golangci-lint` sur push/PR. `.golangci.yml` datait d'avant le schéma v2 de golangci-lint → CI épinglée sur v1.64.2. Le lint tournait apparemment pour la première fois : des centaines de violations de style (`wsl`) préexistantes sont sorties — pas corrigées (hors scope), la CI n'échoue que sur le code nouveau/modifié (`only-new-issues`).
- ✅ Test de fumée ajouté (`d2core/d2asset/d2asset_smoke_test.go`) : boot du gestionnaire d'assets sans fenêtre ni fichiers MPQ. Un vrai test headless du moteur complet (`d2app.Create`/`Run`) reste hors de portée de la CI publique — il lui faudrait un display et les fichiers de jeu originaux, protégés par le droit d'auteur.

**Pourquoi en premier :** sans ça, chaque phase suivante casse la précédente sans qu'on s'en aperçoive.

## Phase 1 — Boucle de jeu minimale 🚧
*Le gros du travail — plusieurs mois. Démarré : commit local `0c9f823`.*

**Fait :**
- Une première tranche verticale du combat, de bout en bout — un monstre tuable (`killable=1` dans monstats.txt) a des PV, un cast de compétence proche de lui déclenche une résolution de dégâts côté serveur, et le client reçoit et applique la mise à jour (PV ou suppression à la mort). Commit local `0c9f823`.
- Les dégâts viennent maintenant de la vraie arme équipée par le joueur (`Equipment.RightHand`/`LeftHand` → lookup dans `Records.Item.Weapons` → `MinDamage`/`MaxDamage`, avec repli sur les dégâts à deux mains puis sur un dégât "à mains nues" si rien n'est trouvé) — remplace le placeholder à plat 2–6. Commit local `1ebdbd3`.

**Simplifications volontaires restantes** (documentées en commentaire `ponytail:` dans `d2networking/d2server/game_server.go`) :
- Ciblage par proximité au point de cast, pas de clic-sur-monstre.
- Pas de jet de précision (attack rating vs défense) — un coup dans le rayon touche toujours.
- Pas de bonus de dégâts de compétence, pas de montée en puissance liée à la force, pas de zone d'effet, une seule cible touchée.

**Reste à faire :**

C'est ici que se joue « jouable ou pas » — tout le reste est du confort autour.

- **Combat** : brancher réellement les formules de dégâts/résistances/chance-de-toucher sur les tables déjà parsées dans `d2records` (aujourd'hui lues, pas forcément exploitées en jeu).
- **Compétences** : au moins un arbre complet et une build jouable par classe, pour valider le pipeline plutôt que viser l'exhaustivité tout de suite.
- **Inventaire/équipement** : les effets des objets doivent modifier les stats du personnage, pas seulement s'afficher dans la grille.
- **IA** : le pathfinding existe déjà dans `d2mapengine` — il manque le comportement de combat des monstres.
- **Sauvegarde** : commencer par un format JSON maison, plus simple à faire évoluer ; la compatibilité `.d2s` (format binaire du jeu original, documenté par la communauté) peut venir après si on veut échanger des sauvegardes avec le vrai client.
- Une quête de l'Acte 1 scriptée de bout en bout pour valider tout le pipeline avant d'en écrire dix de plus.

**Pourquoi un seul acte, une seule build par classe :** valider le pipeline complet sur un périmètre étroit coûte moins cher que la moitié du contenu sur cinq actes.

## Phase 2 — Scripting moderne
*2–3 semaines*

Remplacer une dépendance morte par l'approche que les moteurs moddables utilisent en 2026.

- `d2script` embarque `otto`, un interpréteur JS pur Go sans activité depuis des années. Le remplacer par `wazero` (runtime WASM pur Go, zéro cgo, activement maintenu) : la logique de quêtes/événements devient des modules WASM sandboxés.
- Bénéfice concret : les mods peuvent être écrits dans n'importe quel langage qui compile vers WASM (Rust, AssemblyScript, TinyGo) et tournent dans un bac à sable strict — pas d'accès arbitraire au process hôte comme avec un `Eval()` JS.
- Exposer une API minimale et stable aux scripts plutôt que toute la surface de `d2interface` — moins de surface de rupture à chaque refactor du moteur.

**Pourquoi pas avant la Phase 1 :** pas la peine de moderniser le scripting avant d'avoir des quêtes à scripter.

## Phase 3 — Rendu & confort visuel
*Statut : filtrage des tuiles fait, reste à faire ci-dessous*

- ✅ Filtrage linéaire sur les tuiles de sol/mur/ombre (`d2maprenderer`) — déjà fait, en commit local.
- Un vrai pipeline d'upscaling de sprites, fait proprement : un outil *offline* qui régénère les DC6/DCC en xBRZ/HQx et met à jour **toutes** les métadonnées de taille en cohérence (une première tentative en direct sur la texture a été annulée car elle cassait le positionnement des sprites).
- Post-traitement via les shaders Kage natifs d'ebiten (bloom, vignette, correction colorimétrique) — support déjà intégré, aucune dépendance à ajouter.
- Résolution interne fixée à 800×600 dans `ebiten_renderer.go` — support HiDPI et redimensionnement propre de l'UI à traiter ensemble.

## Phase 4 — Outillage de contenu
*En parallèle*

- Finir la communication `abysswrapper` ↔ moteur pour un vrai aperçu live depuis HellSpawner (le `Read()` actuel est un stub non terminé).
- Documenter le format de mod une fois le scripting WASM (Phase 2) en place — pas avant, pour ne pas documenter une API qui va bouger.

## Phase 5 — Packaging & diffusion
*En dernier*

- Builds cross-plateforme via `goreleaser`.
- Rappel légal explicite au premier lancement : le joueur doit fournir ses propres fichiers MPQ Diablo II + LoD achetés légalement — le binaire n'en contient aucun.

## Ce qu'on ne fait pas maintenant

- Refonte 3D/PBR — incompatible avec des sprites 2D pré-rendus, reviendrait à refaire le jeu.
- Migrer vers AbyssEngine comme base — trop tôt, aucune logique de jeu à ce stade.
- Portage natif WebGPU/Vulkan — ebiten abstrait déjà bien le rendu, gain marginal pour le risque.
- Multijoueur en ligne — décision de scope : le jeu visé est solo uniquement. L'architecture interne d'OpenDiablo2 fait transiter même le solo par un client/serveur local (`d2networking`) ; on la laisse telle quelle sans investir dans du netcode pour du jeu à distance.
- Les 5 actes avant d'avoir un Acte 1 qui tient debout.

## Risques à surveiller

- Solo ou petite équipe : la Phase 1 seule représente plusieurs mois, pas des semaines.
- Le format `.d2s` réel est plus complexe qu'il n'y paraît (checksums, versions) — ne l'attaquer qu'une fois le format maison éprouvé.
- wazero change la façon d'écrire des quêtes ; prévoir un exemple de bout en bout avant de migrer le contenu existant.

## Prochaine étape

Phase 0 (mise à jour ebiten + CI), puis un seul combat contre un seul monstre en Phase 1 pour valider la chaîne complète avant d'en écrire davantage.
