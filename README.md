# DOCUMENT DE CONCEPTION DE JEU (PITCH & GDD SYNTHÉTIQUE)

* **Titre de travail :** *PROJECT : CATALYST* (Titre commercial : *Aura of the Fallen*)
* **Genre :** ARPG / Hack 'n' Slash isométrique
* **Cible :** Joueurs de *Path of Exile*, *Diablo*, *Grim Dawn*, appréciant la complexité de build et la narration déconstructive (façon *Spec Ops: The Line* ou *Nier*).

---

## 1. CONCEPT GÉNÉRAL & AMBIANCE

### Pitch
Dans un monde à l'agonie envahi par des aberrations chimériques, le joueur incarne le « Porteur du Sceau », dernier champion envoyé par les Séraphins pour allumer les Feux Sacrés et purger le royaume d’Aethelgard. Mais ici, le salut a un prix : chaque sort lancé arrache un morceau de la trame du monde. La rédemption promise n’est qu’un holocauste voilé.

### Tonalité & Direction Artistique
* **Visuel :** Un passage progressif d'une *High Dark Fantasy* gothique et majestueuse (marbre blanc, dorures, ciels d'aurore) vers une horreur cosmique et viscérale (matière organique, nécrose, ciels déchirés par le néant).
* **Audio :** Chants grégoriens purs et carillons cristallins au départ, qui se dégradent en dissonances microtonales, chuchotements binauraux et bruits de chair broyée.

### Le Twist Narratif Central
Le joueur ne « purifie » pas le monde : il consume sa bio-énergie vitale pour nourrir une entité extradimensionnelle qui l'utilise comme parasite d'annihilation. Les monstres abattus sont les Gardiens Primordiaux et les habitants difformes altérés pour tenter de survivre à l’atmosphère empoisonnée créée par la magie du héros.

---

## 2. SYSTÈME DE COMBAT & GAMEPLAY (100 % MAGIQUE)

### Zéro Attaque Physique
Aucune arme blanche, aucun arc, aucune parade au bouclier.
* **Armes équipables :** Uniquement des catalyseurs (Bâtons stellaires, Orbes gravitationnels, Grimoires interdits, Runes vivantes).
* **Fonction de l'équipement :** Les armes modifient la trajectoire, la vitesse d'incantation, la dispersion d'énergie ou ajoutent des affixes arcaniques passifs (ex. : *« L’Orbe de Ténèbres convertit 20 % des dégâts de Givre en dégâts de Dégénérescence Psychique »*).

```
[ ÉLÉMENT ]  +  [ FORME / VECTEUR ]  +  [ MODIFICATEURS D'AFFLICTION ]
    │                   │                             │
    └─────────┬─────────┘                             │
              ▼                                       ▼
      SORT DE BASE                     SORT CORROMPU DÉVASTATEUR
```

### Le Moteur de Fusion de Sorts (Grammaire Magique)
Le joueur assigne des **Éléments** (Feu primordial, Givre absolu, Arcane pur, Éther d'ombre) à des **Formes** (Projectile, Déferlement/Nova, Faille/Zone, Malédiction/Aura).
L'intérêt tactique repose sur les **Synergies d'Impact** :
* *Amorçage & Détonation :* Un sort applique un état élémentaire ; un sort d'un autre type consomme cet état pour déclencher un cataclysme cinétique.
* *Persistance réactive :* Deux zones élémentaires superposées créent une nouvelle réalité physique au sol (ex. : Faille d'Ombre + Brasier = Singularité Incendiaire qui aspire et incinère).

### L'Arbre d'Affliction (Évolution Visuelle et Mécanique)
Chaque point investi dans la puissance brute dégrade la nature même du sort :
* Un missile de lumière dorée devient un éclat de verre noir sanglant qui perfore les cadavres.
* Une onde de guérison commence à drainer la vie des entités neutres environnantes.
* Plus les compétences sont maximisées, plus les animations du héros perdent leur grâce pour devenir convulsives, le joueur devenant l’épicentre d'un déchirement spectral.

---

## 3. ÉVOLUTION NARRATIVE ET MÉCANIQUE DU HÉROS

```
   ACTE I : L'ÉLU SACRÉ            ACTE II : LE DOUTE             ACTE III : LE FLÉAU
[ Doré / Éthéré / Pur ]   ──►   [ Violet / Sang / Fissures ] ──► [ Noir / Nécrotique / Néant ]
 • Sauveur du Royaume            • Dégâts collatéraux             • Les "Monstres" = Protecteurs
 • Monstres hostiles             • Les PNJ s'enfuient             • Les PNJ deviennent les Boss
```

### Phase 1 — L'Élu Sacré
* **Esthétique :** Auras dorées, particules blanches, géométrie sacrée, sons de cloches et d'éclairs de foudre céleste. Le personnage porte des soies blanches immaculées.
* **Mécanique :** Sorts précis, ciblés, propres. Pas de dégâts collatéraux. Les jauges de mana sont claires et stables.
* **Narration :** Le héros est acclamé dans les bastions. Les quêtes consistent à « Purifier les Nids de Monstres » et « Restaurer les Pylônes de Vie ».

### Phase 2 — Le Doute
* **Esthétique :** Les sorts crépitent d'une lueur pourpre instable et d’éclairs noirs. Le sang des ennemis tués par magie se vaporise en fumée violette. Des murmures audibles uniquement au casque parasitent la musique.
* **Mécanique :** Apparition des **dégâts collatéraux**. Lancer un sort puissant détruit le décor, embrase la flore et blesse les PNJ alliés présents dans l'arène. Le coût en mana est parfois remplacé par des micro-sacrifices de points de vie.
* **Narration :** Les PNJ réfugiés reculent à votre approche. Les textes de quête deviennent flous ou contradictoires. L'avatar commence à avoir des hallucinations visuelles (les monstres ressemblent brièvement à des chevaliers ou à des prêtres en détresse).

### Phase 3 — La Révélation Maléfique
* **Esthétique :** Magie abyssale, vrilles nécrotiques, géométrie non-euclidienne. L'apparence du héros est monstrueuse : chair brûlée par l'arcanisme, yeux consumés, lévitation incontrôlée.
* **Mécanique :** Le mana est remplacé par **L'Entropie** : plus vous lancez de sorts, plus le monde autour de vous s'effondre (apparition de failles environnementales qui attaquent indistinctement tout ce qui vit).
* **Révélation & Climax :** Le joueur découvre que la « Corruption » combattue au début était le système immunitaire de la planète tentant de repousser l'énergie du joueur. Les boss de fin d'acte ne sont autres que vos mentors, les derniers Grands Mages et le Roi Sanctifié, terrifiés, luttant pour sauver ce qui reste de leur monde contre **vous**.

---

## 4. STRUCTURE DES SYSTÈMES DE JEU

### Trois Combinaisons de Sorts Uniques

#### 1. Singularité Boréale (Contrôle & Shatter)
* **Composants :** *Forme : Faille de Gravité (Arcane)* + *Élément : Givre Profane*.
* **Effet :** Crée un vortex d'énergie dense qui aspire violemment tous les ennemis dans un rayon de 12 mètres, gelant instantanément leurs tissus. Tout impact de projectile ultérieur brise les cibles gelées dans le vortex, libérant une explosion d'éclats osseux et de glace noire qui transperce les ennemis survivants.

#### 2. Suppuration Céleste (DoT & Contagion Épidémique)
* **Composants :** *Forme : Pluie Météorique (Zone)* + *Élément : Ombre Sanglante*.
* **Effet :** Fait s'abattre une averse d'aiguilles éthériques qui transpercent le sol et les ennemis. Les cibles touchées subissent l'affliction « Hémorragie de l'Âme ». À leur mort, leur cadavre explose dans un geyser de miasmes pourpres, répercutant l'effet aux ennemis proches et laissant une flaque caustique qui dissout les armures magiques.

#### 3. Châtiment Récursif (Burst monocible / Rayon)
* **Composants :** *Forme : Rayon Continu (Vecteur)* + *Élément : Feu Primordial corrompu*.
* **Effet :** Projette un faisceau de plasma noir hyper-concentré. Plus le rayon reste focalisé sur la même cible, plus sa largeur augmente, incinérant l'espace-temps autour de la cible. Si la cible survit plus de 3 secondes, une onde de surchauffe atomique éclate, repoussant violemment tous les assaillants adjacents et calcinant le sol de façon permanente.

---

### La Jauge d'Emprise Maléfique (Système Central d'Affliction)

La jauge d’Emprise Maléfique mesure l'empreinte corruptrice laissée par le joueur sur la réalité du monde. Elle augmente de manière irréversible au fil des boss tués et des nœuds de sorts interdits activés.

```
[ ÉTAT DE LA JAUGE D'EMPRISE MALÉFIQUE ]
 0% ────────────────────────── 50% ────────────────────────── 100%
 Sanctifié                   Instable                      Cataclysmique
  • HUD Doré                  • Fissures sur l'UI           • HUD Sanglant / Pulsatile
  • Faune pacifique           • La flore dépérit            • Décor détruit au passage
  • Marchands normaux         • Marchands méfiants          • Marchands hostiles / Fous
```

#### Impacts Visuels & UI :
* **À 25 % :** Les dorures de l'interface utilisateur (HUD) s'oxydent et se fissurent. Les orbes de vie et de mana perdent leur éclat cristallin et se troublent.
* **À 75 % :** L'interface palpite au rythme des incantations. Des runes sanglantes apparaissent sur les bords de l'écran en situation de combat intense. Les textes de lore des objets changent en temps réel dans l'inventaire pour révéler leur véritable nature tragique.

#### Impacts Environnementaux & Systémiques :
* **Dépérissement Passif :** L'herbe et la végétation meurent instantanément sous les pas du joueur.
* **Hostilité Globale :** Les PNJ marchands refusent de commercer, doublent leurs prix ou fuient les campements. Dans le dernier tiers du jeu, les villes jadis sûres sont abandonnées ou défendues par des milices d'anciens alliés que le joueur est contraint d'éradiquer.
* **Modificateurs de Gameplay (Risk/Reward) :** Plus l'Emprise est haute, plus la puissance magique brute augmente (+% Dégâts Arcaniques), mais plus la résistance passive du héros diminue et plus les failles spectrales (qui invoquent des assaillants revanchards) s'ouvrent spontanément.

---

### Arbre de Talents : « L'Altération de la Réalité »

L'arbre n'augmente pas de simples pourcentages statistiques génériques, il modifie les **lois physiques et logiques du jeu**. Il est divisé en 3 branches hérétiques :

```
                        [ CŒUR ARCANIQUE ]
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
  TRANSDIRECTION          ENTROPIE CHIMÉRIQUE       SCISSION VITALE
(Rupture Spatiale)       (Altération d'États)    (Sacrifice & Parasitisme)
```

#### Branche 1 : Transdirection (Espace & Géométrie)
* **Nœud Majeur : « Vecteur Paradoxal »**
  * *Mécanique :* Vos projectiles ne disparaissent plus au contact des murs ou des limites de l'écran. Ils traversent l'espace pour réapparaître du côté opposé de l'arène avec une vitesse et des dégâts critiques doublés.
* **Nœud Majeur : « Déplacement Quantique »**
  * *Mécanique :* Votre dash magique n'est plus une esquive, mais une dislocation : vous échangez instantanément votre position avec l'ennemi ciblé, déclenchant une implosion de vide à son ancien emplacement.

#### Branche 2 : Entropie Chimérique (Physique Élémentaire)
* **Nœud Majeur : « Zéro Absolu / Combustion Spontanée »**
  * *Mécanique :* Les ennemis subissant l'état *Gelé* ne sont plus ralentis : ils sont transformés en bombes thermiques. Tout dégat de feu appliqué sur une cible gelée vaporise l'entité, créant un nuage d'ébullition qui aveugle et asphyxie les cibles adjacentes.
* **Nœud Majeur : « Éther Réactif »**
  * *Mécanique :* Les flaques et zones magiques au sol réagissent à vos propres projectiles. Tirer à travers votre propre Faille d'Ombre charge vos sorts de micro-trous noirs perforants.

#### Branche 3 : Scission Vitale (Hémomancie & Parasitisme)
* **Nœud Majeur : « Autel Vivant »**
  * *Mécanique :* Le joueur peut sceller jusqu'à 50 % de sa barre de vie maximale pour supprimer le temps de recharge (cooldown) de son sort le plus destructeur. Chaque élimination restitue un fragment de bouclier d'os éphémère.
* **Nœud Ultime : « L’Hécatombe Inévitable »**
  * *Mécanique :* Vous ne pouvez plus mourir d'un coup fatal tant qu'une entité vivante (ennemi, allié ou créature passive) se trouve dans un rayon de 15 mètres. Le coup mortel est transféré instantanément à cette entité, la désintégrant pour régénérer 20 % de votre santé. Narrativement, le héros refuse la mort en vampirisant l'existence du monde qui l'entoure.
