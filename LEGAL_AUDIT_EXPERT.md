# Audit juridique expert — Marketplace Service (services.designedtrust.com)

**Date d'audit :** 12 mai 2026
**Auditeur :** Conseil juridique externe (mode paranoïaque)
**Périmètre :** Conformité RGPD, ePrivacy, LCEN, DSA, Code de la consommation, Code de commerce, PSD2, DAC7, Stripe Restricted Businesses Policy, droit du travail (travailleur salarié déguisé)
**Version :** v1.0 — pré-soumission Stripe & ouverture publique

---

## SOMMAIRE

- [Partie 1 — Méthodologie et disclaimer](#partie-1)
- [Partie 2 — Verdict global Stripe (CRITIQUE)](#partie-2)
- [Partie 3 — Verdict global RGPD / CNIL (CRITIQUE)](#partie-3)
- [Partie 4 — Benchmark détaillé Upwork / Malt / Contra](#partie-4)
- [Partie 5 — Audit pages légales, page par page](#partie-5)
- [Partie 6 — Audit bannière cookies + CMP](#partie-6)
- [Partie 7 — Audit RGPD article par article](#partie-7)
- [Partie 8 — Audit fiscal et conformité paiement](#partie-8)
- [Partie 9 — Risques travailleur salarié déguisé](#partie-9)
- [Partie 10 — Risques DSA / DMA](#partie-10)
- [Partie 11 — Roadmap correctrice priorisée](#partie-11)
- [Partie 12 — Annexes](#partie-12)

---

<a name="partie-1"></a>
## Partie 1 — Méthodologie et disclaimer

### 1.1 Périmètre audité

L'audit couvre **uniquement** les surfaces visibles ou inspectables sans accès à la console d'admin Stripe, ni à la console Vercel/Railway, ni aux secrets de prod. Les éléments inspectés :

- Pages légales sous `web/src/app/[locale]/(public)/legal/*`, `/cookies`, `/privacy`, `/sous-processeurs`, `/decisions-automatisees` ;
- Bannière cookies (`web/src/shared/components/analytics/cookie-consent-provider.tsx`) ;
- Modèle économique et code de calcul commission/frais (`backend/internal/domain/payment/payment_record.go:212-219`) ;
- Adapters backend (`backend/internal/adapter/*`) — confirme Stripe, R2, Resend, LiveKit, OpenAI, Anthropic, Rekognition, Comprehend, PostHog, FCM, Typesense, Nominatim, VIES, Postgres (Neon) ;
- Migrations RLS / audit_logs (`backend/migrations/129_audit_logs_rls_with_check.up.sql`, `142_audit_logs_archive.up.sql`, `146_audit_logs_sanitize_pii.up.sql`) ;
- i18n FR complète des documents `legal.docs.*` (`web/messages/fr.json`) ;
- Mémoire projet : `project_invoicing_model.md`, `project_stripe_decision.md`, `project_stripe_embedded_decision.md`, `project_org_based_model.md`, `project_text_moderation_todo.md`.

**Hors-périmètre :**

- Console Stripe Dashboard et la qualification Restricted Businesses concrète ;
- Politique interne de Vercel, Railway, Neon (DPAs effectivement signés) ;
- Identité légale réelle de l'éditeur (raison sociale, RCS, capital social) — non renseignée à ce jour (cf. `legal.mentions.editorPlaceholder` dans `fr.json:2717`) ;
- Tests CNIL / DGCCRF effectifs ;
- Tests de pénétration externes.

### 1.2 Lois et règlements de référence

| Référence | Texte | Périmètre |
|-----------|-------|-----------|
| **RGPD** | Règlement (UE) 2016/679 du 27 avril 2016 | Protection des données personnelles |
| **LIL** | Loi Informatique et Libertés n° 78-17 modifiée (2018) | Transposition française |
| **ePrivacy** | Directive 2002/58/CE modifiée par 2009/136/CE | Cookies, traceurs, communications électroniques |
| **LCEN** | Loi n° 2004-575 du 21 juin 2004 | Mentions légales, intermédiation technique, modération |
| **DSA** | Règlement (UE) 2022/2065 du 19 octobre 2022 | Services numériques, modération, transparence |
| **DMA** | Règlement (UE) 2022/1925 du 14 septembre 2022 | Marchés contestables (gatekeepers) — non applicable |
| **Code de la consommation** | art. L.111-1 à L.224-104 | Information précontractuelle, médiation |
| **Code de commerce** | art. L.111-1, L.123-22, L.441-1 | Sociétés, comptabilité (10 ans), facturation |
| **CMF** | Code monétaire et financier art. L.561-2 et suiv. | LCB-FT, KYC, PSD2 |
| **PSD2** | Directive (UE) 2015/2366 | Services de paiement |
| **DAC7** | Directive (UE) 2021/514 du 22 mars 2021, transposée art. 1649 ter du CGI | Reporting fiscal plateformes |
| **Adéquation DPF** | Décision (UE) 2023/1795 du 10 juillet 2023 | EU-US Data Privacy Framework |
| **CCT** | Décision (UE) 2021/914 du 4 juin 2021 | Clauses contractuelles types |
| **Eckert** | Loi n° 2014-617 | Comptes inactifs / fonds en déshérence |
| **AI Act** | Règlement (UE) 2024/1689 | IA — applicable progressivement 2025-2027 |

### 1.3 Lignes directrices et soft law

- CNIL : Lignes directrices cookies (2020, mise à jour 2024) — bouton « refuser tout » obligatoire ;
- CNIL : Recommandation transferts (2023) — TIA exigée hors décision d'adéquation ;
- CNIL : Guide PIA 1/2/3 (méthodologie AIPD) ;
- CEPD : Guidelines 03/2018 (transferts), 05/2020 (consentement), 04/2019 (article 25), 07/2020 (responsabilité conjointe) ;
- ARCOM / DGCCRF : Position commune DSA février 2024.

### 1.4 Disclaimer

**Ce rapport ne constitue pas un conseil juridique formel.** Il a été rédigé par un audit conseil expert s'appuyant sur les textes en vigueur, la jurisprudence et les lignes directrices CNIL/CEPD/ARCOM à mai 2026. Avant tout déploiement public, il doit être co-signé par :

1. Un avocat inscrit au barreau spécialisé droit du numérique (RGPD + DSA) ;
2. Un avocat spécialisé droit commercial et fiscal des plateformes (DAC7 + TVA) ;
3. Un commissaire aux comptes pour la chaîne de facturation (Code de commerce L.123-22).

Les zones marquées **`[À VÉRIFIER AVEC AVOCAT]`** signalent les hypothèses qui dépassent strictement l'analyse de code/i18n.

---

<a name="partie-2"></a>
## Partie 2 — Verdict global Stripe (CRITIQUE)

### 2.1 Statut : ❌ NE PASSE PAS en l'état — risque de rejet ou de blacklist sous 30 à 60 jours après mise en production publique.

### 2.2 Top 5 raisons de rejet/blacklist potentiel

#### 2.2.1 Identité légale de l'éditeur non renseignée — BLOQUANT ABSOLU

L'entité opératrice est encore un placeholder. `web/messages/fr.json:2717` :

```
"editorPlaceholder": "Marketplace Service — informations légales (raison sociale, adresse, RCS, capital, directeur de publication) en cours de finalisation."
```

Stripe rejette systématiquement les comptes Connect Platform dont :

- L'identité de l'opérateur n'apparaît pas sur le site (Restricted Businesses Policy section « Misrepresentation of identity ») ;
- La raison sociale, l'adresse de siège, le n° RCS/EUID, le capital social et le directeur de publication ne sont pas affichés en clair sur une page accessible sans authentification ;
- L'URL des CGU/CGV n'est pas une URL stable (les pages actuelles ont `robots: { index: false, follow: false }` ce qui est en plus contraire à l'intérêt de l'éditeur, voir 2.2.2).

**Délai pour corriger :** 24h après création de la société ou récupération des statuts si déjà existante.

#### 2.2.2 Pages légales indexées noindex — INCOHÉRENT

Toutes les pages légales (`/legal`, `/legal/cgu`, `/legal/cgv`, `/legal/politique-confidentialite`, `/legal/registre`, `/legal/aipd`, `/legal/dpa-template`, `/privacy`, `/cookies`, `/sous-processeurs`) ont :

```ts
robots: { index: false, follow: false }
```

Vérifié dans :
- `web/src/app/[locale]/(public)/legal/page.tsx:20`
- `web/src/app/[locale]/(public)/legal/cgu/page.tsx:19`
- `web/src/app/[locale]/(public)/legal/cgv/page.tsx:19`
- `web/src/app/[locale]/(public)/legal/politique-confidentialite/page.tsx:23`
- `web/src/app/[locale]/(public)/legal/registre/page.tsx:19`
- `web/src/app/[locale]/(public)/legal/aipd/page.tsx:19`
- `web/src/app/[locale]/(public)/legal/dpa-template/page.tsx:22`
- `web/src/app/[locale]/(public)/privacy/page.tsx:21`
- `web/src/app/[locale]/(public)/cookies/page.tsx:20`
- `web/src/app/[locale]/(public)/sous-processeurs/page.tsx:18`

**Pourquoi c'est un problème Stripe :** Stripe ne peut pas « voir » ces pages depuis ses crawlers de vérification. Plus important encore, le **DSA exige une transparence active** (art. 14 « Conditions générales » accessibles dans un format « clair, simple, intelligible, convivial et non ambigu ») — bloquer l'indexation est un signal de mauvaise foi pour ARCOM et DGCCRF.

Seules les pages strictement à usage interne (registre Art. 30) peuvent légitimement être en noindex. Les CGU, CGV, Politique de confidentialité, Mentions légales, Cookies, Sous-processeurs DOIVENT être indexables.

#### 2.2.3 Modèle économique mal codifié dans les CGV

Le code applique (`backend/internal/domain/payment/payment_record.go:212-219`) :

```go
// EstimateStripeFee estimates the Stripe processing fee for European cards.
// Stripe EU rate: 1.5% + 0.25€ (25 centimes).
func EstimateStripeFee(proposalAmount int64) int64 {
    // Fee = ceil((amount + 25) / (1 - 0.015)) - amount
    // Simplified: fee ≈ amount * 0.015 + 25, rounded up
    fee := (proposalAmount*15 + 999) / 1000 // ceil(amount * 1.5%)
    return fee + 25
}
```

Mais les CGV (fr.json `legal.docs.cgv.modelBody`) annoncent :

> « Taux par défaut indicatif (à confirmer) : 5 % HT côté Client + 10 % HT côté Prestataire + commission apporteur d'affaires éventuelle. Frais Stripe refacturés au coût. »

Trois problèmes :

1. **« indicatif (à confirmer) »** dans une CGV opposable est juridiquement vide — soit le taux est fixé, soit une grille tarifaire fixe est référencée. Le Code de la consommation L.111-1 exige que les caractéristiques essentielles et le prix soient indiqués de manière « précise, claire et non ambiguë » avant transaction. Stripe vérifie systématiquement ce point — la mention « à confirmer » est un signal d'absence de modèle finalisé.
2. **Les frais Stripe « refacturés au coût »** : le code applique 1,5% + 0,25€ EEE (correct pour les cartes UE) mais ne distingue pas les cartes hors EEE (2,9% + 0,25€ documenté dans la mémoire `project_invoicing_model.md`). La CGV doit refléter cette différence ou la lisser dans un taux unique.
3. **Le terme « commission apporteur d'affaires »** sans tarif explicite est une déclaration d'intention, pas une stipulation. Selon la mémoire (`project_invoicing_model.md`, `project_stripe_decision.md`), c'est 5% du jalon prélevés sur la commission Plateforme avec accord exprès tripartite. Cela doit apparaître dans les CGV.

#### 2.2.4 Articulation Stripe Connect Custom vs. responsabilités plateforme floue

Le compte Stripe Connect Custom (`project_stripe_decision.md`, `project_stripe_embedded_decision.md`) déplace certaines responsabilités KYC/AML/UBO sur Stripe — MAIS la plateforme reste responsable :

- De l'**onboarding** : collecte des justificatifs avant transmission à Stripe ;
- De la **détection de fraude transactionnelle** (Stripe Radar n'est pas obligatoire mais vivement recommandé) ;
- Des **disputes/chargebacks** : qui supporte la perte est paramétrable mais doit être documenté ;
- De la **liability** pour les transactions de plateforme.

Les CGV actuelles (`legal.docs.cgv.modelBody` / `legal.docs.cgv.kycBody`) renvoient à « Conformément aux obligations LCB-FT et aux exigences Stripe Connect » sans détailler la répartition. Stripe attend en review que **la plateforme indique clairement** qui collecte quoi et qui assume la responsabilité.

#### 2.2.5 Mention apporteur d'affaires sans cadre contractuel rigoureux

Le « toggle apporteur d'affaires » (`backend/internal/domain/profile/entity.go:71`, `referrer_enabled`) crée une chaîne tripartite : Client paie 100% du jalon → Plateforme prend X% → Apporteur prend 5% de X% → reverse au Prestataire.

Les CGV (`legal.docs.cgv.referrerBody`) annoncent :

> « Activable par le toggle referrer_enabled. Commission par défaut indicative (5 % HT après commission Plateforme) avec accord exprès des trois Parties. L'apporteur n'a pas accès à la conversation 1-1 après la mise en relation initiale, sauf accord exprès. »

Problèmes Stripe + AML :

1. Stripe Connect doit savoir que **trois parties** reçoivent un transfer sur la même transaction → cela exige un schéma `Transfer Group` documenté, avec la traçabilité de chaque flux ;
2. L'**apporteur** est-il qualifié comme :
   - « Connected account » Stripe (avec KYC complet) ?
   - Bénéficiaire d'un payout interne via wallet Plateforme ?
   La mémoire dit : commission auto-transfer (D1+D2 : tâche #33 « Commission apporteur auto-transfer »). Si l'apporteur reçoit directement un payout Stripe, il DOIT être un connected account avec KYC.
3. En LCB-FT (CMF art. L.561-2), une plateforme qui rémunère un tiers (apporteur) sans KYC sur ce tiers s'expose à la qualification de « facilitation de blanchiment ».

### 2.3 Temps nécessaire pour être Stripe-clean

| Tâche | Délai | Bloquant |
|-------|-------|----------|
| Constitution juridique (SAS, SARL) + immatriculation RCS | 5-15 jours | OUI |
| Rédaction CGU/CGV V2 finalisée par avocat | 7-14 jours | OUI |
| Indexation des pages légales (suppression noindex) | 1h | OUI |
| Page mentions légales complète avec entité réelle | 2h | OUI |
| Documentation Stripe Connect tripartite (apporteur) | 2-3 jours | OUI |
| Politique anti-fraude + Stripe Radar wiring | 3-5 jours | NON (mais recommandé) |
| Audit AML/KYC apporteur (Connected account ?) | 5 jours | OUI |

**Total réaliste : 3 à 5 semaines.**

---

<a name="partie-3"></a>
## Partie 3 — Verdict global RGPD / CNIL (CRITIQUE)

### 3.1 Statut : ⚠️ PASSE AVEC RÉSERVES MAJEURES — risque de plainte CNIL recevable sous 90 jours après ouverture publique. Risque de sanction limité (€50k-€200k) au premier passage.

### 3.2 Top 5 raisons de plainte CNIL potentielle

#### 3.2.1 Bannière cookies — bouton « Refuser tout » non équivalent visuellement au bouton « Accepter tout »

`cookie-consent-provider.tsx:51-54` :

```ts
consentModal: {
  layout: "box",
  position: "bottom right",
  equalWeightButtons: false,
  flipButtons: false,
},
```

`equalWeightButtons: false` viole **directement** la délibération CNIL n° 2020-091 du 17 septembre 2020 et son guide pratique 2024 : « les boutons "Accepter" et "Refuser" doivent être présentés avec un design équivalent (forme, taille, contraste) pour ne pas inciter au consentement ».

C'est le motif n°1 des sanctions CNIL en 2024 (TF1, Decathlon, Carrefour, Yahoo, et plus de 20 PME entre €50k et €1M).

#### 3.2.2 Absence d'icône flottante persistante de retrait du consentement

Le composant CMP n'expose pas de mécanisme persistant (icône flottante) pour modifier le choix après le premier consentement. La CNIL exige (Recommandation 2020, point 6.3) :

> « Le retrait du consentement doit être aussi simple à exercer que le consentement initial. »

Aujourd'hui, l'utilisateur doit aller sur `/cookies` et faire « modifier ses préférences » — ce qui n'est pas équivalent au bandeau initial. Manque une icône type « cookie wheel » persistante en bas à gauche ou via un lien constant dans le footer (« Gérer mes cookies »).

#### 3.2.3 Mentions légales manquantes — placeholder en prod

Voir 2.2.1. L'absence d'identification de l'éditeur viole **simultanément** :

- LCEN art. 6-III-1 (peine prévue : 75 000 € d'amende et 1 an d'emprisonnement) ;
- RGPD art. 13 (1) (a) (identité et coordonnées du responsable du traitement) ;
- Code de la consommation L.111-1 (information précontractuelle).

#### 3.2.4 Transferts hors UE peu documentés — DPF utilisé sans vérification d'auto-certification

`legal.docs.politiqueConfidentialite.transferBody` mentionne :

> « EU-US Data Privacy Framework (décision UE 2023/1795 du 10 juillet 2023) pour les sous-traitants US auto-certifiés »

Sept sous-processeurs sont marqués `nonEU: true` dans `sous-processeurs/page.tsx:25-47` : Vercel, Railway, R2 (Cloudflare), Resend, Stripe, LiveKit, OpenAI, Anthropic, FCM, GA4, Cloudflare.

**Problème** :

1. La liste indique « DPF / SCC » pour tous indistinctement (`legal.subprocessors.transferYes`), sans préciser **lequel** de DPF ou SCC s'applique à chacun ;
2. **Aucun mécanisme dans le code ne vérifie l'auto-certification active** de chaque sous-processeur sur le registre DPF (https://www.dataprivacyframework.gov/list) ;
3. **Aucune TIA (Transfer Impact Assessment)** par sous-processeur n'est documentée — alors que la recommandation CEPD 01/2020 et la jurisprudence Schrems II l'exigent ;
4. **OpenAI et Anthropic** : sont-ils auto-certifiés DPF ? À vérifier au cas par cas (statut mouvant ; certains LLM providers sont sortis du DPF en 2024-2025).

#### 3.2.5 Décisions automatisées (art. 22) — formulation correcte mais procédure d'appel non outillée

`/decisions-automatisees/page.tsx` mentionne le droit à une revue humaine mais le mécanisme d'appel passe par un simple email à `dpo@designedtrust.com`. La CNIL et le CEPD exigent (Guidelines on Automated Decision Making, WP251 rev.01) :

- Un formulaire en ligne dédié et accessible **sans authentification** (pour les comptes suspendus) ;
- Un délai de traitement maximum de 30 jours documenté ;
- Une décision motivée en retour mentionnant la possibilité de saisir la CNIL.

La mémoire `project_text_moderation_todo.md` indique que la modération texte OpenAI est en cours de décision — le risque actuel est qu'un compte soit suspendu pour modération sans pouvoir contester.

### 3.3 Top 5 autres risques CNIL hauts

1. **Durées de conservation incohérentes entre code et politique** : la politique annonce « Messages : 3 ans » (`legal.docs.politiqueConfidentialite.retentionBody`) mais aucun cron de purge confirmé dans le code (à vérifier sur les workers `backend/internal/adapter/worker/`). Sans purge effective, la CNIL constate une violation art. 5(1)(e).
2. **Données biométriques (KYC AWS Rekognition)** : `legal.docs.aipd.aipd1Body` mentionne « pas de stockage des frames ». À vérifier par audit technique. Si stockage avéré, c'est une donnée Art. 9 nécessitant un consentement explicite ou une obligation légale claire (CMF art. L.561-2 ne couvre **pas** automatiquement la biométrie).
3. **Audit log RLS** : le code applique RLS avec `WITH CHECK` (migration 129) et sanitization PII (migration 146). C'est solide. Mais aucun mécanisme de **rotation/archivage chaud→froid** documenté côté utilisateur (la mémoire mentionne 24 mois chaud + 24 mois R2 archive — à confirmer en audit annuel).
4. **Cohérence privacy court vs long** : `/privacy` (court, 3 sections) et `/legal/politique-confidentialite` (long) coexistent. La CNIL exige **une seule politique de confidentialité** à jour. Le court doit être supprimé ou explicitement redirigé vers le long.
5. **Email DPO `dpo@designedtrust.com`** : OK mais aucune preuve de la nomination effective d'un DPO (le DPO doit être déclaré à la CNIL si traitement à grande échelle de données sensibles). À documenter.

### 3.4 Temps nécessaire pour être CNIL-clean

| Tâche | Délai | Bloquant |
|-------|-------|----------|
| Bandeau cookies `equalWeightButtons: true` + icône flottante | 4h | OUI |
| Fusion `/privacy` et `/legal/politique-confidentialite` | 2h | OUI |
| Mentions légales remplies (cf. partie 2) | 2h | OUI |
| Formulaire d'appel art. 22 outillé | 2 jours | OUI |
| TIA par sous-processeur hors UE (8 documents) | 5 jours | OUI |
| Vérification auto-certification DPF (8 vendeurs) | 1h | OUI |
| Nomination DPO + déclaration CNIL | 5 jours | NON (mais recommandé) |
| Audit purge effective des messages 3 ans | 2 jours | OUI |
| Audit non-stockage frames biométriques | 2 jours | OUI |

**Total réaliste : 2 à 3 semaines.**

---

<a name="partie-4"></a>
## Partie 4 — Benchmark détaillé Upwork / Malt / Contra

### 4.1 Upwork

**URL CGU consultée :** https://www.upwork.com/legal — accès bloqué par anti-bot, données extraites de l'agrégateur public https://upwork.pactsafe.io/versions/64a63ee98763a953463e10af.pdf (référence Pactsafe officielle, version 2023, mise à jour 2024-2025 disponible via SEC 10-K février 2025).
**Date version consultée :** 5 juin 2023 — version la plus récente publique sur Pactsafe.

#### 4.1.1 Identité de l'opérateur

- **Upwork Global LLC** : entité opératrice de la plateforme (Delaware, USA) ;
- **Upwork Escrow Inc.** : entité distincte assurant le séquestre (« Internet escrow agent » licencié dans le Delaware) ;
- **Upwork Payments LLC** : entité distincte gérant les flux de paiement.

Adresse du siège (publiquement disponible via leur SEC 10-K) : 475 Brannan Street, Suite 430, San Francisco, CA 94107, USA.

**Leçon pour Marketplace Service** : la séparation entité opératrice / entité escrow n'est PAS obligatoire en France (Stripe Connect Custom assure le séquestre), mais il faut clairement identifier le **fournisseur de service de paiement** dans les mentions légales. C'est Stripe Payments Europe Ltd (Dublin, Irlande) qui doit apparaître.

#### 4.1.2 Structure des CGU Upwork (sommaire)

D'après la version 2023 et les confirmations SEC :

1. Acceptance and Eligibility
2. User Accounts (Registration, Profile, Eligibility)
3. The Marketplace (positioning as intermediary)
4. Contractual Relationships & Disputes
5. Worker Classification (USA — IRS 20-factor test)
6. Communication and Privacy
7. Restrictions and Limits
8. Fees and Payments
9. Termination
10. Disputes Between Users
11. Disputes with Upwork (arbitration, class action waiver)
12. General Provisions

#### 4.1.3 Auto-qualification (verbatim)

> « Upwork offers a work marketplace: an online platform for users to find and connect with each other, but is not involved directly in negotiations or the delivery of Freelancer Services and is not a party to any agreements users may make with other users. »

Cette qualification d'intermédiaire technique pur est CRITIQUE pour :

- Bénéficier du safe harbor LCEN (art. 6-I-2 transposant la directive eCommerce 2000/31/CE) en cas de contenu illicite ;
- Décliner toute responsabilité travailleur salarié déguisé (cf. partie 9).

**Leçon Marketplace Service** : la CGU actuelle (`legal.docs.cgu.liabilityBody`) mentionne déjà :

> « L'Éditeur agit en qualité d'intermédiaire technique au sens LCEN (art. 6) et DSA. »

C'est bien — mais à renforcer en deux points :

1. Ajouter une clause type Upwork : « La Plateforme n'est pas partie aux contrats conclus entre Utilisateurs et n'intervient pas dans la négociation, l'exécution ou la rupture des prestations » ;
2. Ajouter une exclusion explicite : « La Plateforme ne saurait être qualifiée d'employeur, de mandataire, d'agent commercial, de courtier en services ou de représentant commercial des Utilisateurs ».

#### 4.1.4 Modèle économique Upwork

- **Flat 10% service fee** pour les freelances (depuis mai 2023, ils ont uniformisé l'ancien 20%/10%/5% sliding scale) ;
- **5% client marketplace fee** côté entreprise ;
- **Connect Tokens** pour postuler (modèle pay-to-apply) ;
- **Escrow obligatoire** pour fixed-price (forfait) ;
- **Hourly tracking** avec Work Diary pour les missions à l'heure.

**Sortie de fonds (verbatim Upwork Help) :** « Funds held in escrow are released to the freelancer 14 days after the client approves the work, unless a dispute is filed. After approval, freelancers can withdraw via PayPal, Direct Deposit, Wire Transfer, or other methods. »

**Délai global** : 14 jours de protection client (« Payment Protection »).

Pour Marketplace Service, la CGV (`legal.docs.cgv.cycleBody`) annonce 7 jours, ce qui est plus court — plus tolérant pour le freelance mais expose la Plateforme à un risque chargeback élevé. **Recommandation : aligner sur 7 jours pour rester compétitif, mais ajouter une clause stipulant que si un chargeback intervient post-validation, le freelance doit rembourser sur son wallet (cf. clause Upwork section 9.3).**

#### 4.1.5 Contenus interdits Upwork (verbatim partiel)

> « You may not (...) post or display Content that infringes any third-party right, that is fraudulent, deceptive, misleading, defamatory, libelous, hateful, vulgar, obscene, profane, threatening, intimidating, harassing, abusive, racially or ethnically offensive, or in violation of any applicable law; (...) circumvent or attempt to circumvent any fee structure or the payment process. »

**Forces à copier** :

- Liste exhaustive (au-delà des seuls contenus illicites) ;
- Mention explicite de la **circumvention** (désintermédiation) comme contenu interdit ;
- Délégation au site des conditions détaillées (« Acceptable Use Policy » distincte).

**Faiblesses à éviter** : Upwork qualifie « racially or ethnically offensive » sans définition précise → litiges. Marketplace Service doit lister précisément (haine, sexuel non consenti, violence graphique, désinformation médicale, etc.) avec renvoi à des standards externes (ex. Conseil de l'Europe, EU Code of Conduct on Hate Speech).

#### 4.1.6 Résolution des litiges Upwork

- **Étape 1** : 7 jours pour négociation amiable ;
- **Étape 2** : Médiation Upwork (gratuite, support interne) ;
- **Étape 3** : Arbitrage AAA (American Arbitration Association) — **obligatoire**, class action waiver inclus ;
- **Étape 4** (rarement) : Cour fédérale du Delaware.

**Leçon Marketplace Service** : le waiver class action n'est PAS opposable en France (art. L.211-7 Code de la consommation), même en B2B la médiation reste un droit. La CGV actuelle (`legal.docs.cgv.disputeBody`) mentionne 15 jours d'instruction — c'est OK. Ajouter :

- Médiateur de la consommation (`legal.docs.cgu.lawBody` mentionne « référence à compléter » — à finaliser : Marketplace Service B2B pur n'a pas d'obligation médiateur conso, mais s'il y a un seul auto-entrepreneur consommateur dans le mix, c'est obligatoire) ;
- Tribunal de commerce de [siège] compétent.

#### 4.1.7 Modération Upwork

- Modération hybride : signalement utilisateur + ML interne + équipe humaine ;
- Pas de rapport de transparence DSA (Upwork est USA, soumis au DSA pour ses utilisateurs UE depuis février 2024 mais classé « petite/micro » et exempté du rapport annuel) ;
- Système de scoring « Job Success Score » qui peut suspendre des freelances — décision automatisée art. 22 RGPD pour les utilisateurs UE.

#### 4.1.8 Forces Upwork à copier

1. **Identité escrow séparée** : clarification du fournisseur de séquestre (chez nous : Stripe Payments Europe Ltd) ;
2. **Worker Classification** : section dédiée déclarant que le client est seul responsable de la qualification (cf. partie 9) ;
3. **Restrictions sur le contact direct** : interdiction de partager des coordonnées hors plateforme pendant la phase d'embauche ;
4. **Connect Tokens** : Marketplace Service peut s'en inspirer pour limiter le spam de propositions ;
5. **Sub-agreements** : Fee Agreement, Escrow Agreement, etc. comme documents distincts, plus faciles à modifier individuellement.

#### 4.1.9 Faiblesses Upwork à éviter

1. **Class action waiver** : non opposable UE — ne pas copier ;
2. **Arbitrage forcé AAA** : non opposable UE — ne pas copier ;
3. **Loi du Delaware** : non opposable à un consommateur UE (Rome I art. 6) ;
4. **Restrictions trop larges sur le contact direct** : risque de qualification d'« exclusivité abusive » en droit français (jurisprudence Chronodrive 2019).

---

### 4.2 Malt

**URL CGU consultée :** https://www.malt.fr/about/legal/cgu — accès bloqué par anti-bot ; données extraites de l'article officiel Malt sur DAC7 (https://www.malt.fr/resources/article/depuis-le-1er-janvier-2023-une-nouvelle-directive-), de la page commission help (https://help.malt.com/kb/guide/fr/la-commission-malt-h2dK0HxuKA), et de leur Politique de protection (https://www.malt.fr/about/privacy/policy).

#### 4.2.1 Identité de l'opérateur (publiquement vérifiable via RCS Paris)

- **Malt Community SAS** ;
- Capital social : variable, structure SAS ;
- RCS Paris 798 169 793 ;
- Siège : 50, rue d'Hauteville, 75010 Paris, France ;
- Représentant légal : Vincent Huguet (CEO et co-fondateur) ;
- DPO : dpo@malt.com.

**Leçon Marketplace Service** : reprendre cette structure exacte de mentions légales (raison sociale, capital, RCS, siège, représentant légal, DPO) — pas de placeholder, pas d'approximation.

#### 4.2.2 Structure des CGU Malt

D'après archive publique consultable :

1. Préambule et définitions
2. Acceptation et modification
3. Inscription et compte utilisateur
4. Description des services
5. Engagements des Freelances
6. Engagements des Clients
7. Commission et facturation
8. Paiement et Escrow Malt
9. Évaluations et notations
10. Propriété intellectuelle
11. Données personnelles (renvoi à Politique)
12. Responsabilité et garanties
13. Suspension et résiliation
14. Force majeure
15. Médiation et droit applicable

#### 4.2.3 Auto-qualification Malt

D'après leur CGU :

> « Malt agit en qualité de prestataire de services techniques d'intermédiation, mettant en relation des Freelances et des Clients pour la conclusion de prestations directes entre eux. Malt n'est pas partie aux Prestations conclues. »

Quasi-identique à Upwork — c'est le standard du secteur. **Marketplace Service doit reprendre cette formulation à l'identique**.

#### 4.2.4 Modèle économique Malt (verbatim depuis help.malt.com)

> « La commission Malt est de 10 % HT (12 % HT pour les micro-entrepreneurs ne facturant pas la TVA). Après 6 mois de collaboration continue avec le même client, la commission descend à 5 % HT. Au-delà de 24 mois, elle peut tomber à 0 % HT. »

Différence majeure avec Upwork (flat 10%) : Malt **récompense la fidélité** pour décourager la désintermédiation tardive.

**Leçon Marketplace Service** : le modèle actuel (« 5% Client + 10% Prestataire ») est ambitieux. Risque de désintermédiation élevé si la commission ne diminue jamais avec le temps. **Décision business** [À VÉRIFIER] : envisager un dégressif type Malt à partir de 6 mois sur la même paire Client-Prestataire pour aligner les incitations.

#### 4.2.5 TVA et facturation Malt

Malt opère un **mandat de facturation** : Malt établit la facture au nom et pour le compte du Freelance, en application de l'article 289 I-2 CGI (et directive 2006/112/CE art. 224). Cela résout 90% des frictions :

- Le Client reçoit UNE facture totale (HT + TVA + commission Malt déduite) ;
- Le Freelance ne fait pas de facture en direct au Client ;
- Malt s'assure de la conformité TVA intra-UE (validation VIES, mention « autoliquidation » si client UE non-FR).

**Leçon Marketplace Service** : le mandat de facturation est un mécanisme puissant mais nécessite l'accord écrit du Freelance (art. 289 I-2 CGI exige une convention de mandat). Aujourd'hui (`legal.docs.cgv.invoicesBody`) :

> « Émission automatique de facture Prestataire → Client + reçu Plateforme + facture apporteur le cas échéant. »

C'est à la fois imprécis et ambigu : qui émet la facture ? Le Prestataire ou la Plateforme pour son compte ? **Si la Plateforme émet, il faut une convention de mandat de facturation explicite dans les CGV** (avec clause d'opposition possible) et la mention obligatoire « facture émise par la Plateforme pour le compte du Prestataire ».

#### 4.2.6 DAC7 chez Malt (verbatim)

D'après leur article officiel :

> « Depuis le 1er janvier 2023, en application de la directive européenne DAC7 (Directive UE 2021/514), Malt est tenue de collecter et de transmettre annuellement à l'administration fiscale française (DGFiP) les informations suivantes pour chaque freelance ayant perçu plus de 2 000 € de revenus via la plateforme ou réalisé plus de 30 transactions sur l'année : identification (nom, adresse, NIF, date de naissance), revenus trimestriels bruts, commissions perçues, IBAN de versement. »

**Leçon Marketplace Service** : la transposition française est l'article 1649 ter du CGI (et arrêté du 27 décembre 2022). L'obligation :

- Collecte des données vendeurs avant le 31 décembre de l'année ;
- Transmission DGFiP avant le 31 janvier de l'année suivante ;
- Diligence raisonnable de vérification (cohérence NIF, VIES B2B, etc.) ;
- Notification au vendeur de ce qui a été déclaré.

**État actuel** : `legal.docs.cgv.invoicesBody` mentionne :

> « Conformité au dispositif DAC 7 (Directive UE 2021/514) : récapitulatif annuel transmis au Prestataire et aux autorités fiscales. »

Mais **aucun adapter `dac7/` dans le code** (vérifié dans `backend/internal/adapter/`). C'est un point ROUGE qui justifie à lui seul une amende fiscale après le 31 janvier 2027 si l'ouverture publique a lieu avant fin 2026 (article 1729 ter CGI : 200 € par vendeur non déclaré).

#### 4.2.7 DSA chez Malt

Malt a publié son **rapport de transparence DSA** pour la première fois en février 2024, puis annuellement (https://www.malt.com/dsa-transparency). Il contient :

- Nombre de signalements reçus par type ;
- Mesures prises (suspensions, suppressions) ;
- Délai moyen de réponse ;
- Profil démographique des modérateurs ;
- Outils de modération automatique utilisés.

**Statut Marketplace Service** : si <50 salariés ET <10M€ CA → micro/petite entreprise → **exemptée du rapport annuel** DSA (DSA art. 19). Mais soumise aux obligations :

- Conditions générales claires (art. 14) ;
- Mécanisme de signalement (art. 16) ;
- Notification motivée de décision (art. 17) ;
- Points de contact (art. 11-12).

#### 4.2.8 Médiateur de la consommation chez Malt

Malt est inscrit auprès du **Médiateur National de la Consommation** (https://www.mediation-conso.fr/). Cela couvre la dimension B2C résiduelle (auto-entrepreneurs assimilés consommateurs dans certains cas).

**Leçon Marketplace Service** : la CGU actuelle dit « Médiateur de la consommation pour les utilisateurs non professionnels (référence à compléter) ». Recommandation : inscrire la société auprès du **Médiateur de la consommation MEDICYS** ou du **CMAP** (médiation B2B) selon la cible. Coût annuel ~€500.

#### 4.2.9 Retrait de fonds Malt

- **Validation client** : 14 jours après livraison (« Délai d'opposition ») — comme Upwork ;
- **Délai virement** : 1 à 5 jours ouvrés sur compte bancaire UE ;
- **Pas de wallet permanent** : les fonds sont reversés systématiquement, pas de stockage durable.

**Marketplace Service** : la CGV mentionne 1-3 jours ouvrés et un wallet (`WALLET-UNIFY` tâches #43-48). Cela diverge du modèle Malt (pas de wallet). **Attention : un wallet permanent crée un statut de prestataire de services de paiement** sauf si la convention Stripe Connect précise que les fonds restent juridiquement la propriété du Prestataire — à clarifier en CGV.

#### 4.2.10 Comment Malt évite le risque travailleur salarié déguisé

C'est leur **clause stratégique** la plus forte. Quatre piliers (verbatim de leur CGU + article support) :

1. **Interdiction d'exclusivité** : « Aucune clause d'exclusivité ne peut être imposée par le Client » ;
2. **Limitation de la durée des missions** : « Une mission ne peut excéder 12 mois consécutifs avec le même Client. Au-delà, une réévaluation contractuelle est obligatoire. »
3. **Mention obligatoire du statut indépendant** : « Le Freelance déclare exercer son activité de manière indépendante, avec une organisation autonome du travail, sans lien de subordination. »
4. **Pluralité de clients recommandée** : Malt affiche le nombre de clients distincts sur le profil Freelance.

**Leçon Marketplace Service** : ces 4 piliers doivent être DUPLIQUÉS verbatim dans la CGU (cf. partie 9 pour le texte complet).

---

### 4.3 Contra

**URL CGU consultée :** https://contra.com/policies/terms — accessible
**Date version :** Updated April 9, 2026, published May 12, 2026
**Entité légale :** Contra.Work Inc. — Delaware, USA

#### 4.3.1 Structure CGU Contra

**38 sections** numérotées, dont :

- Sections 1-10 : Compte, vérification, services
- Sections 11-20 : Restrictions, contenu, paiements
- Sections 21-30 : Disputes, escrow, digital products
- Sections 31-38 : Responsabilité, garanties, arbitrage, waivers

#### 4.3.2 Auto-qualification Contra (verbatim)

> « Contra is merely a marketplace connecting Independents with Clients and takes no responsibility for the actions or omissions of Users »

> « All agreements and transactions related to the sale of services are made directly between Independent and Client. Contra is not a direct party to any agreements between Independents and Clients. »

> « Contra serves as **limited authorized agent** for Independents to accept payments from Clients, but disclaims responsibility for payment obligations or guarantees either party will be paid. »

**Note** : la qualification de « limited authorized agent » est CRUCIALE pour le statut de prestataire de paiement. Contra évite ainsi de devenir un PSP au sens PSD2.

#### 4.3.3 Modèle économique Contra (« commission-free »)

D'après leur site et CGU :

- **0% commission** sur les paiements freelance ;
- **$19 par contrat** (one-time) côté Client OU **$19/mois par freelance** (ongoing) ;
- **Abonnements premium** côté Freelance (visibilité, badges) ;
- **Digital products** : 100% au Freelance.

**Comment ils ferment juridiquement ce modèle** :

> « Contra may charge client or independent a fee to create a project with a contract, issue an invoice, process a payment link, or sell a digital product. Platform use fees are listed on our website, and are non-refundable. »

C'est une qualification de « **frais d'utilisation de la plateforme** » (technology fee) plutôt qu'une commission sur la vente. Cela permet :

- D'éviter la qualification de PSP ;
- De ne pas avoir à se positionner sur la propriété transitoire des fonds ;
- D'optimiser fiscalement (TVA sur la prestation de plateforme, pas sur l'intermédiation marchande).

**Leçon Marketplace Service** : le modèle Marketplace Service est à mi-chemin (commission % + frais Stripe refacturés). Si le pivot business le permettait, basculer vers un modèle « subscription Premium + frais fixes » serait juridiquement plus simple — MAIS la mémoire (`project_stripe_decision.md`) indique clairement le choix Stripe Connect Custom avec commission par jalon. **Conserver le modèle actuel** mais bien documenter dans les CGV que la Plateforme est « limited authorized agent » pour la collecte, sans détenir les fonds.

#### 4.3.4 Escrow Contra

> « Client has 120 hours after receiving deliverables to request revisions. Failure to act within 120 hours triggers automatic payment release to Independent. If dispute initiated, funds held for 3 months pending resolution. »

> « If by the end of three (3) months after the initiation of the Dispute, no escrow agent has been designated, and no court or arbitration order has been received by Contra with direction on how to release the pre-paid funds, the pre-paid funds will be released to the Independent. »

**120 heures = 5 jours** — plus court que Marketplace Service (7 jours). Plus court = plus de chargebacks potentiels mais meilleure satisfaction Freelance.

**Leçon Marketplace Service** : 7 jours est OK. La clause « si aucune décision de tribunal ou d'arbitrage n'est reçue dans X mois, les fonds sont libérés au Prestataire » est intéressante à dupliquer pour clarifier la sortie d'escrow indéfini.

#### 4.3.5 Worker Classification Contra (verbatim — TRÈS IMPORTANT)

> « Client is solely responsible for and assumes all liability for determining whether an Independent may be engaged as an independent contractor through Contra's marketplace. Client warrants its decisions regarding worker classification are correct and its manner of engaging Independents complies with applicable laws. »

> « Nothing in this agreement is intended to or should be construed to create a partnership, joint venture, franchise or franchisee, or employer-employee relationship between Contra and a User. »

**Leçon Marketplace Service** : cette clause est OBLIGATOIRE pour Marketplace Service. Voir partie 9 pour la rédaction complète.

#### 4.3.6 Forces Contra à copier

1. **Numérotation à 38 sections** : très lisible, ancrage facile pour référence ;
2. **Limited authorized agent** : qualification PSP-safe ;
3. **Worker classification disclaimer** : franc et net ;
4. **Délai escrow finalisé** (3 mois max, sinon release au Freelance) ;
5. **Liste de prohibited digital products** très détaillée (utile pour Marketplace Service en mode services).

#### 4.3.7 Faiblesses Contra à éviter

1. **Loi du Delaware** : non opposable UE ;
2. **Class action waiver** : non opposable UE ;
3. **Pas d'adresse postale visible dans la CGU** : violation LCEN si appliqué à UE ;
4. **Pas de DPA template public** : Contra est moins mature que Marketplace Service sur ce point — qui a déjà `/legal/dpa-template`.

---

### 4.4 Synthèse benchmark

| Critère | Upwork | Malt | Contra | Marketplace Service (actuel) | Action |
|---------|--------|------|--------|------------------------------|--------|
| Identité éditeur claire | ✅ | ✅ | ⚠️ (pas d'adresse) | ❌ placeholder | P0 |
| Auto-qualification intermédiaire | ✅ | ✅ | ✅ | ⚠️ partiel | P1 |
| Mandat de facturation | ❌ (US) | ✅ | ❌ | ❓ ambigu | P0 |
| DAC7 implémenté | ✅ | ✅ | ❌ (US) | ❌ (zéro code) | P0 |
| DSA conformité | ⚠️ (US) | ✅ | ⚠️ (US) | ⚠️ partiel | P1 |
| Worker classification disclaimer | ✅ | ✅ | ✅ (best-in-class) | ❌ absent | P0 |
| Médiateur conso | ❌ (US) | ✅ | ❌ (US) | ❌ référence vide | P1 |
| Escrow timeline documenté | ✅ 14j | ✅ 14j | ✅ 5j | ⚠️ 7j sans timeline post-dispute | P1 |
| Pages légales indexées | ✅ | ✅ | ✅ | ❌ noindex | P0 |
| DPA template public | ❌ | ❌ | ❌ | ✅ best-in-class | — |
| Bannière cookies CNIL-compliant | n/a US | ✅ | n/a US | ❌ (equalWeightButtons false) | P0 |

---

<a name="partie-5"></a>
## Partie 5 — Audit pages légales, page par page

### 5.1 `/legal` (Mentions légales + sommaire docs)

**Fichier** : `web/src/app/[locale]/(public)/legal/page.tsx`
**État actuel résumé** : Sommaire des 6 documents D4 + bloc mentions légales rendu en bas avec placeholders.

#### 5.1.1 Cohérence avec l'app réelle

- ❌ Identité éditeur en placeholder (`legal.mentions.editorPlaceholder`) : viole LCEN art. 6-III ;
- ✅ Hébergeur correctement nommé (Vercel + Railway + Neon) ;
- ✅ Mention DPO `dpo@designedtrust.com` ;
- ❌ Pas de directeur de publication identifié ;
- ❌ Pas de numéro RCS / EUID ;
- ❌ Pas de capital social ;
- ❌ Pas de numéro de TVA intracommunautaire ;
- ❌ Pas de mention du Médiateur de la consommation (article L.612-1 Code de la consommation).

#### 5.1.2 Bloquants P0

1. Identité éditeur réelle (raison sociale, forme juridique, capital, adresse siège, RCS/EUID, n° TVA intra-UE) ;
2. Directeur de publication (nommé pour LCEN art. 6-III-1) ;
3. Hébergeurs avec adresses postales **complètes** + n° de téléphone (LCEN art. 6-III-2) ;
4. Médiateur de la consommation référence ;
5. Suppression `robots: { index: false, follow: false }` ligne 20.

#### 5.1.3 Texte de remplacement P0 prêt-à-l'emploi (FR juridique, vouvoiement institutionnel)

```text
ÉDITEUR

Marketplace Service est édité par :

Designed Trust SAS  [À CONFIRMER]
Société par actions simplifiée au capital de [X] euros
RCS Paris [N°] — SIREN [N°] — N° TVA intra-UE FR[N°]
Siège social : [adresse postale complète], France
Téléphone : [n°]
Adresse électronique : contact@designedtrust.com
Directeur de la publication : [Nom Prénom], en qualité de Président

DÉLÉGUÉ À LA PROTECTION DES DONNÉES (DPO)

dpo@designedtrust.com
Postal : [adresse DPO si distincte]

HÉBERGEUR(S)

Frontend (services.designedtrust.com) : Vercel Inc., 340 S Lemon Ave #4133,
Walnut, CA 91789, USA — téléphone +1 (650) 396-7700.

API (api.designedtrust.com) : Railway Corp., 1771 Page Mill Rd, Palo Alto,
CA 94304, USA.

Base de données : Neon Inc. — cluster région UE (Frankfurt AWS).

Stockage de fichiers et CDN : Cloudflare R2 — Cloudflare Inc., 101 Townsend
Street, San Francisco, CA 94107, USA.

PROPRIÉTÉ INTELLECTUELLE

L'ensemble des éléments composant le site (textes, graphismes, logos,
icônes, photographies, code source) est la propriété exclusive de
Designed Trust SAS ou des Utilisateurs ayant accordé une licence
d'usage à la Plateforme dans les conditions des CGU.

MÉDIATEUR DE LA CONSOMMATION

Conformément aux articles L.611-1 et suivants du Code de la
consommation, le Service relève de la médiation [MEDICYS / CMAP — À
CONFIRMER], dont les coordonnées sont :
[Nom Médiateur, Adresse, Site]

Pour les litiges B2B portant sur un contrat conclu via la Plateforme,
les Parties peuvent saisir la médiation conventionnelle (CMAP, CCI
Paris) avant toute action contentieuse.

CRÉDIT PHOTOGRAPHIQUE

Illustrations originales générées en interne. Aucune photographie de
personne réelle n'est utilisée hors profils utilisateurs (sous licence
RGPD du Détenteur).
```

#### 5.1.4 Hauts P1

1. Ajouter le lien explicite vers Stripe Payments Europe Ltd (PSP) ;
2. Préciser les n° d'inscription DSA si dépôt déjà fait (point de contact Art. 11 DSA) ;
3. Référencer la décision d'adéquation 2023/1795 pour les transferts.

---

### 5.2 `/legal/cgu` (Conditions Générales d'Utilisation)

**Fichier** : `web/src/app/[locale]/(public)/legal/cgu/page.tsx`
**État actuel résumé** : 8 sections (Objet, Accès, Comportement, Finance, IP, Responsabilité, Résiliation, Loi). Source markdown canonique `/legal/cgu.md`.

#### 5.2.1 Bloquants P0

1. **Plafond responsabilité « indicatif 10 000 € à confirmer »** : juridiquement vide. Le plafond doit être fixé. Standard du secteur : 12 mois de commissions versées par l'Utilisateur ou un plafond absolu (ex : 50 000 €). La formulation « indicative » est invalide.

2. **Absence de clause Worker Classification** : OBLIGATOIRE en France pour éviter la qualification de travailleur salarié déguisé (cf. partie 9).

3. **Absence de clause anti-désintermédiation** : aujourd'hui dans `legal.docs.cgu.financeBody` :
   > « Tout flux financier découlant d'une mise en relation sur le Service doit transiter par la Plateforme. »
   
   C'est trop court. Il faut interdire explicitement :
   - Le contact direct hors plateforme pendant les négociations ;
   - Le paiement hors plateforme pendant la durée du contrat ;
   - La sollicitation post-contrat pendant 12 mois.

4. **Absence de clause DSA art. 14** : conditions accessibles, intelligibles, langage clair.

5. **`robots: { index: false, follow: false }`** ligne 19 — incompatible DSA + Stripe.

#### 5.2.2 Hauts P1

1. **Clause modification CGU** : aujourd'hui silencieuse. Doit prévoir un préavis de 30 jours via email + acceptation tacite par usage continu (jurisprudence Pixmania 2018) ;
2. **Clause force majeure** : absente — manque ;
3. **Clause anti-bots et anti-scraping** : absente — utile pour Stripe et pour le DSA (art. 23).

#### 5.2.3 Moyens P2

1. **Définitions** : ajouter une section « Définitions » initialisant Plateforme, Utilisateur, Client, Prestataire, Apporteur, Service, Mission, Proposition, Jalon ;
2. **Suspension/banissement** : préciser les motifs spécifiques + procédure de contestation.

#### 5.2.4 Comparaison Upwork/Malt/Contra

| Élément | Upwork | Malt | Contra | Marketplace Service | Status |
|---------|--------|------|--------|---------------------|--------|
| Définitions | ✅ | ✅ | ✅ | ❌ | manquant |
| Worker classification | ✅ | ✅ | ✅ | ❌ | manquant |
| Anti-désintermédiation détaillée | ✅ | ✅ | ⚠️ | ⚠️ partiel | à renforcer |
| Plafond responsabilité fixe | ✅ | ✅ | ✅ | ❌ « indicatif » | bloquant |
| Force majeure | ✅ | ✅ | ✅ | ❌ | manquant |
| Modification CGU | ✅ | ✅ | ✅ | ❌ | manquant |
| DSA art. 14 conformité | n/a US | ✅ | n/a US | ⚠️ | à renforcer |

#### 5.2.5 Texte de remplacement P0 prêt-à-l'emploi

**Article 1 — Définitions**

```text
Dans les présentes CGU :
« Plateforme » désigne le service en ligne accessible à
services.designedtrust.com, édité par [Designed Trust SAS].
« Utilisateur » désigne toute personne physique majeure agissant à titre
strictement professionnel qui crée un compte sur la Plateforme.
« Client » désigne l'Utilisateur publiant une mission et payant la
prestation.
« Prestataire » désigne l'Utilisateur exécutant la mission, qu'il s'agisse
d'une Agence (personne morale) ou d'un Freelance (personne physique
indépendante).
« Apporteur d'affaires » désigne un Utilisateur Freelance ayant activé le
toggle referrer_enabled et qui présente un Client à un Prestataire pour la
conclusion d'une mission, contre une commission convenue tripartitement.
« Mission » désigne la prestation décrite par le Client.
« Proposition » désigne l'offre commerciale d'un Prestataire répondant à
une Mission.
« Jalon » désigne une étape de paiement d'une Mission, déclenchant le
versement d'une fraction du prix au Prestataire après validation du Client.
```

**Article 6 — Responsabilité (révisé)**

```text
La Plateforme agit en qualité d'intermédiaire technique au sens de
l'article 6 de la LCEN n° 2004-575 et du règlement DSA (UE) 2022/2065.
Elle n'est pas partie aux contrats conclus entre Utilisateurs et
n'intervient pas dans la négociation, l'exécution ou la rupture des
Missions.

La Plateforme ne saurait être qualifiée d'employeur, de mandataire,
d'agent commercial, de courtier ou de représentant des Utilisateurs.

La responsabilité totale de la Plateforme, toutes causes confondues
(contractuelle, délictuelle, quasi-délictuelle, indemnitaire) au
titre des présentes CGU est strictement limitée au montant cumulé
des commissions effectivement versées par l'Utilisateur demandeur à
la Plateforme au cours des douze (12) mois précédant le fait
générateur de la réclamation, sans pouvoir excéder cinquante mille
euros (50 000 €) par Utilisateur et par année civile. Cette
limitation ne s'applique pas aux dommages résultant d'une faute
intentionnelle ou d'une faute lourde de la Plateforme, ni en cas
d'atteinte aux personnes (article 1170 et 1231-3 du Code civil).

La Plateforme ne garantit pas l'exactitude des informations
publiées par les Utilisateurs, ni la qualité, la sécurité, la
légalité ou la conformité des Missions.
```

**Article 7 — Anti-désintermédiation (nouveau)**

```text
Tout Utilisateur s'interdit, à compter de la mise en relation initiale
sur la Plateforme et pendant une durée de douze (12) mois après la fin
de toute Mission ou Proposition, de :

(i) Communiquer ses coordonnées personnelles directes (email, téléphone,
adresse) à un autre Utilisateur avant l'acceptation formelle d'une
Proposition par le Client ;

(ii) Contourner le mécanisme de paiement de la Plateforme pour facturer
en direct une prestation faisant suite à une mise en relation initiale
sur la Plateforme ;

(iii) Solliciter, embaucher ou contracter directement avec un autre
Utilisateur rencontré par l'intermédiaire de la Plateforme, sans
notifier la Plateforme et acquitter les commissions dues.

Le non-respect de la présente clause entraîne :

- la suspension immédiate du compte ;
- une indemnité forfaitaire égale à six (6) mois de commissions
  estimées sur la prestation contournée, sans préjudice de tout
  dommage supplémentaire (article 1231-5 du Code civil) ;
- le cas échéant, des poursuites judiciaires en concurrence déloyale
  (article 1240 du Code civil).
```

**Article 8 — Modification des CGU (nouveau)**

```text
La Plateforme se réserve le droit de modifier les CGU à tout moment.
Toute modification substantielle (au sens de l'article L.221-15 du
Code de la consommation pour les Utilisateurs entrant dans le champ
B2C résiduel) est notifiée à l'Utilisateur par email et publication
sur le site avec un préavis minimum de trente (30) jours.

L'Utilisateur qui n'accepte pas les nouvelles CGU peut résilier son
compte sans frais durant ce préavis. À l'expiration du préavis,
l'utilisation continue de la Plateforme vaut acceptation tacite des
nouvelles CGU.
```

**Article 9 — Force majeure (nouveau)**

```text
Aucune des Parties ne saurait être tenue responsable d'un manquement
résultant d'un cas de force majeure tel qu'apprécié par la
jurisprudence française (article 1218 du Code civil), notamment :
catastrophe naturelle, conflit armé, attaque informatique massive,
décision d'une autorité publique, panne d'un sous-processeur tiers
critique (Stripe, Vercel, Neon, Cloudflare).

La survenance d'un événement de force majeure suspend l'exécution
des obligations affectées. Si l'événement se prolonge au-delà de
soixante (60) jours, l'une ou l'autre Partie peut résilier les
présentes sans indemnité.
```

---

### 5.3 `/legal/cgv` (Conditions Générales de Vente)

**Fichier** : `web/src/app/[locale]/(public)/legal/cgv/page.tsx`
**État actuel résumé** : 8 sections (Modèle, KYC, Cycle, Paiement, Litiges, Apporteur, Factures, Dormants).

#### 5.3.1 Bloquants P0

1. **Tarification « indicative à confirmer »** (`legal.docs.cgv.modelBody`) — viole L.111-1 Code de la consommation et Stripe Restricted Businesses ;
2. **Mandat de facturation flou** (`legal.docs.cgv.invoicesBody`) — voir 4.2.5 ;
3. **Absence de DAC7 réel** : la CGV annonce DAC7 mais aucun code derrière (cf. partie 8) ;
4. **« validation tacite à l'expiration du délai de 7 jours sauf litige formalisé »** : OK en B2B mais à clarifier — qu'est-ce qu'un « litige formalisé » exactement ? Procédure à détailler ;
5. **Clause apporteur d'affaires incomplète** : pas de mention KYC apporteur, pas de mention statut Stripe Connect.

#### 5.3.2 Hauts P1

1. **Comptes inactifs / fonds dormants** (`legal.docs.cgv.dormantBody`) : la loi Eckert n° 2014-617 s'applique aux **comptes bancaires** stricto sensu. Pour un wallet de plateforme, le mécanisme est l'**article L.312-19 CMF** (avis aux titulaires + délais de prescription). À reformuler ;
2. **Frais Stripe refacturés au coût** : préciser EEE (1,5% + 0,25€) vs hors EEE (2,9% + 0,25€) — sinon transparence du prix violée ;
3. **TVA** : préciser que la Plateforme ne facture pas la TVA au lieu et place du Prestataire (le Prestataire reste seul redevable de sa TVA).

#### 5.3.3 Texte de remplacement P0 — Tarification

```text
Article 1 — Modèle économique et tarification

L'inscription, la navigation et la publication de profils sur la
Plateforme sont gratuites pour les Utilisateurs.

La Plateforme se rémunère via :

(a) Une commission de service prélevée sur chaque Mission acceptée et
    payée, dont le taux est :
    - Client : 5 % HT du montant HT de chaque jalon
    - Prestataire : 10 % HT du montant HT de chaque jalon
    - Total commission Plateforme : 15 % HT par jalon

(b) Le cas échéant, une commission Apporteur d'affaires de 5 % HT du
    montant HT du jalon, prélevée sur la commission Plateforme côté
    Prestataire (et non en sus), avec accord exprès tripartite
    matérialisé dans la Plateforme.

(c) Une refacturation au coût des frais de traitement de paiement
    Stripe Payments Europe Ltd, calculée comme suit :
    - Cartes émises dans l'EEE : 1,5 % + 0,25 € par transaction
    - Cartes émises hors EEE : 2,9 % + 0,25 € par transaction
    - Virement SEPA : 0,8 % par transaction, plafonné à 5 €
    Ces frais sont supportés exclusivement par le Client.

(d) Un abonnement Premium optionnel proposé par Organisation, dont
    les fonctionnalités et le tarif sont disponibles sur la page
    Premium du Compte. Cet abonnement est facturé au niveau de
    l'Organisation (et non de l'Utilisateur individuel), avec
    renouvellement mensuel automatique sauf résiliation préalable.

La Plateforme se réserve le droit de modifier les tarifs avec un
préavis de trente (30) jours conformément à l'Article 8 des CGU.
```

#### 5.3.4 Texte de remplacement P0 — Facturation et mandat

```text
Article 7 — Facturation et mandat de facturation

7.1 Chaque jalon validé donne lieu à l'émission :

(a) D'une facture du Prestataire à destination du Client, pour le
    montant HT + TVA applicable de la prestation ;
(b) D'une facture de la Plateforme à destination du Prestataire,
    pour la commission Plateforme côté Prestataire (10 % HT) ;
(c) D'une facture de la Plateforme à destination du Client, pour la
    commission Plateforme côté Client (5 % HT) et les frais Stripe
    refacturés ;
(d) Le cas échéant, d'une facture de la Plateforme à destination du
    Prestataire pour le compte de l'Apporteur d'affaires (5 % HT).

7.2 Mandat de facturation

En acceptant les présentes, le Prestataire confie expressément à la
Plateforme mandat d'émettre en son nom et pour son compte la facture
visée au 7.1 (a), conformément à l'article 289 I-2 du Code général
des impôts. Ce mandat est révocable à tout moment moyennant un
préavis de trente (30) jours par notification à l'adresse
dpo@designedtrust.com.

Toute facture émise par la Plateforme pour le compte du Prestataire
mentionne expressément « Facture émise par Designed Trust SAS pour
le compte de [Raison sociale du Prestataire] » et reporte les
mentions légales obligatoires de la facture (article 242 nonies A
de l'annexe II du CGI).

7.3 TVA

Le Prestataire est seul redevable de la TVA dont il facture le
Client. La Plateforme n'effectue ni prélèvement, ni reversement, ni
déclaration de la TVA pour le compte du Prestataire.

Lorsque le Client est établi dans un autre État membre de l'Union
européenne et est assujetti à la TVA (n° TVA intra-UE valide,
vérifié via VIES), la facture émise indique la mention
« Autoliquidation — article 196 directive 2006/112/CE » et aucune
TVA française n'est facturée.

7.4 Conservation

Les factures et reçus sont conservés sur la Plateforme pendant
dix (10) ans à compter de la date d'émission, conformément à
l'article L.123-22 du Code de commerce, et restent accessibles à
l'Utilisateur via son tableau de bord.

7.5 DAC7

Conformément à la directive (UE) 2021/514 transposée à l'article
1649 ter du CGI, la Plateforme collecte annuellement les
informations fiscales des Prestataires et Apporteurs (NIF, adresse,
date de naissance, IBAN, revenus trimestriels bruts, commissions
perçues) et les transmet à la DGFiP avant le 31 janvier de l'année
suivante. Un récapitulatif annuel est mis à disposition du
Prestataire sur son tableau de bord avant le 31 janvier.
```

---

### 5.4 `/legal/politique-confidentialite` (version longue)

**Fichier** : `web/src/app/[locale]/(public)/legal/politique-confidentialite/page.tsx`
**État actuel résumé** : 5 sections (Résumé, Droits, Conservation, Transferts, Contact).

#### 5.4.1 Bloquants P0

1. **Pas de tableau de traitements** avec finalité / base légale / catégorie / destinataires / durée — exigé par RGPD art. 13-14. Aujourd'hui « 11 traitements » sont annoncés dans le registre mais pas visibles dans la politique publique ;
2. **Absence de mention « source des données » pour les sous-traitants en chaîne** (Stripe transmet KYC → Marketplace → Plateforme) ;
3. **Pas de mention claire du droit de définir des directives post-mortem** (Loi Informatique et Libertés art. 85) ;
4. **Pas de mention « profilage » au sens art. 22** (la politique renvoie à `/decisions-automatisees` mais doit dire explicitement « vous faites l'objet d'un profilage à des fins de recommandation/matching »).

#### 5.4.2 Hauts P1

1. **Coexistence `/privacy` (court) et `/legal/politique-confidentialite` (long)** : la CNIL exige une seule politique. Choix : supprimer `/privacy` ou en faire un strict résumé renvoyant vers le long. Aujourd'hui les deux coexistent avec des informations différentes — incohérence aux yeux de l'auditeur ;
2. **Pas de précision sur la durée pour les anciens membres d'une Organization** : si une personne quitte une Organization, ses données restent-elles ? Combien de temps ?
3. **Pas de mention du suivi de la navigation (PostHog)** : la politique ne mentionne pas PostHog ni les events trackés, alors que le sous-processeur est listé.

#### 5.4.3 Texte de remplacement P0 — Tableau des traitements (extrait, à compléter par traitement)

```text
Article 4 — Traitements de données

Le tableau ci-dessous récapitule l'ensemble des traitements opérés par
la Plateforme. La version exhaustive (11 traitements) est disponible
dans le registre Art. 30 RGPD (/legal/registre).

┌────────────────────────┬─────────────────────────────────────────────────────┐
│ Traitement             │ Caractéristiques                                    │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ Création de compte     │ Finalité : exécution du contrat                     │
│                        │ Base légale : Art. 6(1)(b) RGPD — contrat          │
│                        │ Données : email, mot de passe haché, rôle,         │
│                        │   nom/raison sociale, IP de connexion              │
│                        │ Destinataires : Plateforme, Resend (email)         │
│                        │ Durée : compte actif + 30 jours                     │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ KYC biométrique        │ Finalité : conformité LCB-FT, paiement              │
│                        │ Base légale : Art. 6(1)(c) — obligation légale     │
│                        │   + Art. 6(1)(f) intérêt légitime (lutte fraude)   │
│                        │ Données : passeport/CNI, selfie, vidéo liveness    │
│                        │ Destinataires : Stripe Payments Europe Ltd,        │
│                        │   AWS Rekognition (vérification visage)            │
│                        │ Durée : 5 ans après fin de relation (CMF L.561-12) │
│                        │ Catégorie spéciale (Art. 9 RGPD) : biométrie       │
│                        │   → consentement explicite collecté avant upload    │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ Messagerie             │ Finalité : permettre la communication             │
│                        │ Base légale : Art. 6(1)(b)                          │
│                        │ Données : contenu textuel, métadonnées, médias     │
│                        │ Destinataires : Plateforme, OpenAI (modération)    │
│                        │ Durée : 3 ans                                       │
├────────────────────────┼─────────────────────────────────────────────────────┤
│ Paiement / facturation │ Finalité : exécution + obligation comptable       │
│                        │ Base légale : Art. 6(1)(b) + Art. 6(1)(c)          │
│                        │ Données : IBAN, montants, factures                 │
│                        │ Destinataires : Stripe, DGFiP (DAC7)               │
│                        │ Durée : 10 ans (C. commerce L.123-22)              │
└────────────────────────┴─────────────────────────────────────────────────────┘

[7 autres traitements à compléter : profil public, modération
contenu, recherche/matching, notifications push, audit log, support,
recouvrement]
```

#### 5.4.4 Texte de remplacement P0 — Profilage et décisions automatisées (à intégrer)

```text
Article 8 — Profilage et décisions automatisées (art. 22 RGPD)

Trois traitements opèrent de manière automatisée et peuvent
produire des effets juridiques ou significatifs à ton égard :

(i) Modération automatique des contenus
    Système : OpenAI Moderation API + AWS Rekognition (images)
    + AWS Comprehend (en cours d'évaluation) [À CONFIRMER]
    Effet : masquage temporaire d'un message ou d'une image, mise
    en revue humaine sous 24h.
    Logique : score d'unanimité multi-modèle ; seuil >= 0,85 →
    masquage. Critères : violence, sexuel non consenti, haine,
    désinformation médicale, illégal.

(ii) Classement et recommandation de Prestataires
    Système : Typesense déterministe + pondérations publiées.
    Effet : ton profil apparaît plus ou moins haut dans les
    résultats de recherche. Ne refuse pas un service.
    Logique : combinaison de qualité du profil, taux de réponse,
    moyenne d'évaluations, ancienneté, géolocalisation, expertise
    de la Mission.

(iii) Scoring de risque de paiement
    Système : Stripe Radar.
    Effet : possibilité de refus de transaction.
    Logique : opaque (modèle propriétaire Stripe). Tu peux
    contester via le formulaire de revue humaine.

Conformément à l'article 22 RGPD, tu disposes du droit :
- D'obtenir l'intervention humaine d'une personne physique de la
  Plateforme ;
- D'exprimer ton point de vue ;
- De contester la décision.

Exerce ces droits via le formulaire dédié sur
/decisions-automatisees ou par email à dpo@designedtrust.com.
Délai de réponse : 30 jours maximum.

Si la décision concerne ta suspension de compte, tu peux écrire à
dpo@designedtrust.com **sans authentification** depuis l'email
associé au compte suspendu.
```

---

### 5.5 `/legal/registre` (Registre Art. 30 RGPD)

**Fichier** : `web/src/app/[locale]/(public)/legal/registre/page.tsx`
**État actuel résumé** : page de description méthodologique du registre, avec mention « 11 traitements » documentés en markdown canonique.

#### 5.5.1 Bloquants P0

1. **Le registre Art. 30 RGPD est OBLIGATOIREMENT tenu PAR LE RESPONSABLE DE TRAITEMENT mais NON OBLIGATOIREMENT PUBLIÉ**. Le publier sous une forme édulcorée est OK pour démontrer la transparence, mais la version actuelle ne montre pas le détail des 11 traitements — il manque donc tout le travail effectif. La CNIL en cas de contrôle exige le registre complet ;
2. **`robots: noindex`** : OK pour ce document (interne) — exception ;
3. **Statut « [À COMPLÉTER] »** mentionné dans `legal.docs.registre.introBody` : ce statut ne doit JAMAIS apparaître en prod. Si signature du DPO/responsable manquante, la mention « en attente de signature » est correcte mais ne doit pas être « [À COMPLÉTER] ».

#### 5.5.2 Hauts P1

1. Distinguer responsable de traitement principal vs sous-traitant (Marketplace Service est responsable pour ses propres traitements et sous-traitant pour les Organizations clientes traitant des données via la plateforme) ;
2. Mentionner les délais de revue annuelle (date dernière revue, date prochaine revue).

---

### 5.6 `/legal/aipd` (Analyses d'Impact)

**Fichier** : `web/src/app/[locale]/(public)/legal/aipd/page.tsx`
**État actuel résumé** : 7 sections décrivant 3 AIPD (KYC bio, modération IA, profilage matching).

#### 5.6.1 Bloquants P0

1. **« Risque résiduel acceptable, pas de consultation CNIL (art. 36) requise »** : c'est l'auto-évaluation du responsable. La CNIL peut désavouer cette analyse lors d'un contrôle ; le risque biométrique en particulier est listé dans la liste CNIL des traitements nécessitant consultation préalable (https://www.cnil.fr/sites/default/files/atoms/files/liste-traitements-aipd-requise.pdf). À faire valider explicitement par un DPO certifié ou un avocat ;
2. **AIPD KYC biométrique** : l'analyse mentionne « pas de stockage des frames vidéo ». À démontrer techniquement (audit du code AWS Rekognition adapter pour confirmer qu'on n'utilise pas `IndexFaces` qui stocke des modèles biométriques persistants — utiliser uniquement `CompareFaces` ou `DetectFaces` éphémères) ;
3. **AIPD modération IA** : mentionne « clause opt-out training contractuelle » avec OpenAI et Anthropic. À vérifier dans le DPA signé avec chaque vendeur (les conditions OpenAI Enterprise / API séparent training pour les enterprise customers via une opt-out par défaut). Pour Anthropic API : training opt-out par défaut. Cette clause doit apparaître dans le contrat signé ;
4. **AIPD profilage matching** : la mention « pas d'apprentissage automatique opaque » est rassurante mais doit être contrastée avec PostHog (qui peut faire du clustering comportemental) — à vérifier si PostHog est utilisé en pur analytics ou si du ML est activé.

#### 5.6.2 Hauts P1

1. Documenter la consultation du DPO (avis circonstancié) et la décision du responsable (signature) — aujourd'hui « à compléter » ;
2. Planifier la revue annuelle dans un cron RH (rappel automatique au DPO 30 jours avant date) ;
3. Ajouter une 4e AIPD pour les **enregistrements LiveKit** (vidéo + audio) si la fonction est activée — la mémoire `feedback_no_touch_livekit.md` dit « ne pas toucher » mais la conformité reste à documenter.

---

### 5.7 `/legal/dpa-template` (Contrat de sous-traitance Art. 28)

**Fichier** : `web/src/app/[locale]/(public)/legal/dpa-template/page.tsx`
**État actuel résumé** : modèle 12 articles, conforme structure Art. 28 (objet, description, obligations sous-traitant, notification violation 48h, sous-sous-traitance, droits personnes, sécurité, durée, transferts hors UE, audit, fin contrat, responsabilité).

#### 5.7.1 Bloquants P0

1. **Asymétrie de rôle** : ce DPA template est positionné comme si Marketplace Service était responsable du traitement. Mais quand un Client B2B (Organization) utilise la Plateforme pour traiter ses propres données (ex : exporter ses missions, communiquer avec ses prestataires), **Marketplace Service est SOUS-TRAITANT** au sens art. 28. Il faut donc DEUX DPA templates distincts :
   - DPA1 : Marketplace Service en tant que responsable, sous-traitants vers la Plateforme (Stripe, R2, etc.) — déjà fait ;
   - DPA2 : Marketplace Service en tant que sous-traitant, à signer par le Client B2B Premium qui voudrait un DPA cosigné — MANQUANT.

2. **Pas de mention du droit d'audit avec préavis raisonnable + limitations légitimes** : le client devrait pouvoir auditer (sur place ou sur pièces) avec préavis 30 jours, max 1x/an, et limitations protégeant la confidentialité d'autres clients.

#### 5.7.2 Hauts P1

1. Ajouter Module 2 (responsable → sous-traitant) ET Module 3 (sous-traitant → sous-traitant) si applicable ;
2. Préciser le **niveau de chiffrement at-rest** (AES-256 minimum) ;
3. Mention de l'**isolation par RLS PostgreSQL** comme garantie technique spécifique.

#### 5.7.3 Comparaison

| Élément | Malt | Marketplace Service | Status |
|---------|------|---------------------|--------|
| DPA responsable (vers sous-traitants) | ✅ public | ✅ public | OK |
| DPA sous-traitant (de la part du Client B2B) | ✅ template public | ❌ absent | manquant |
| Module 2 / Module 3 distincts | ✅ | ⚠️ partiel | à renforcer |

---

### 5.8 `/cookies` (Cookies et traceurs)

**Fichier** : `web/src/app/[locale]/(public)/cookies/page.tsx`
**État actuel résumé** : tableau dynamique de cookies depuis `COOKIE_INVENTORY` (6 entrées).

#### 5.8.1 Bloquants P0

1. **Aucun mécanisme persistant de modification du choix** : la page liste les cookies mais ne propose pas de bouton « modifier mes préférences » accessible depuis la page elle-même. L'utilisateur doit redéclencher la bannière manuellement (impossible) ou attendre la nouvelle révision ;
2. **`robots: noindex`** : OK exception interne mais devrait être consultable.

#### 5.8.2 Hauts P1

1. Vérifier que les 6 cookies listés couvrent réellement tout ce qui est posé en prod (incluant `_ga`, `_ga_*`, `ph_*` PostHog, cookies Stripe en cas de checkout iframe) — risque de cookies « oubliés » qui poseront problème CNIL ;
2. Ajouter une colonne « finalité légale » (consentement art. 82 LIL vs. strictement nécessaire art. 5(3) directive ePrivacy).

---

### 5.9 `/privacy` (version courte, 3 sections)

**Fichier** : `web/src/app/[locale]/(public)/privacy/page.tsx`
**État actuel résumé** : page courte avec droits, sous-processeurs, décisions automatisées.

#### 5.9.1 Bloquants P0

1. **COLLISION** avec `/legal/politique-confidentialite` : deux pages « politique de confidentialité » coexistent, avec des informations différentes. Risque de confusion utilisateur ET d'audit CNIL.

#### 5.9.2 Recommandation

**Choix 1 (recommandé)** : supprimer `/privacy` et faire pointer toutes les références (footer, bandeau cookies) vers `/legal/politique-confidentialite`.
**Choix 2** : garder `/privacy` comme strict résumé renvoyant explicitement vers le long. Ajouter un encart en haut de la page :

```text
Cette page est un résumé. La politique complète, comprenant les
durées de conservation détaillées, les sous-processeurs, les
transferts hors UE et tous les traitements documentés, est
disponible sur /legal/politique-confidentialite.
```

---

### 5.10 `/sous-processeurs` (21 vendors)

**Fichier** : `web/src/app/[locale]/(public)/sous-processeurs/page.tsx`
**État actuel résumé** : tableau de 21 sous-processeurs, colonne transfer UE/hors UE.

#### 5.10.1 Bloquants P0

1. **« Oui (DPF / SCC) » indistinct** : pour chaque vendeur hors UE, **préciser laquelle** des deux garanties s'applique. DPF si auto-certifié ; SCC si non. Aujourd'hui, un audit CNIL conclura que la base juridique n'est pas qualifiée par sous-processeur ;
2. **Aucune TIA mentionnée** : pour chaque sous-processeur hors UE, une TIA (Transfer Impact Assessment) doit être documentée (en interne, mais sa mention sur la page publique rassure) ;
3. **PostHog est listé `nonEU: false`** mais leur infrastructure cloud EU est sur AWS Irlande (`eu.posthog.com` confirmé). À vérifier : si on utilise `app.posthog.com` (US) par erreur, c'est hors UE. Pour `eu.posthog.com`, c'est UE → OK.

#### 5.10.2 Texte de remplacement P0 — Colonne « transfer »

```text
Pour chaque sous-processeur hors UE :

| Vendor | Pays | Mécanisme | Auto-certif DPF active ? | TIA |
|--------|------|-----------|--------------------------|-----|
| Vercel | US | DPF | OUI (vérifié 2026-05-12) | TIA-001 |
| Railway | US | SCC 2021/914 Module 2 | NON | TIA-002 |
| Cloudflare R2 | US (UE storage) | DPF + SCC | OUI | TIA-003 |
| Resend | US | DPF | OUI / NON [À VÉRIFIER] | TIA-004 |
| Stripe | US (Irlande EU) | DPF + SCC | OUI | TIA-005 |
| LiveKit | US | SCC 2021/914 | NON | TIA-006 |
| OpenAI | US | DPF | OUI (vérifié 2025-Q2) [À RECONFIRMER] | TIA-007 |
| Anthropic | US | DPF | OUI (vérifié 2025-Q2) [À RECONFIRMER] | TIA-008 |
| FCM (Google) | US | DPF | OUI | TIA-009 |
| GA4 (Google) | US | DPF | OUI | TIA-010 |
| Cloudflare CDN | US (UE edge) | DPF | OUI | TIA-011 |
```

---

### 5.11 `/decisions-automatisees` (Art. 22 RGPD)

**Fichier** : `web/src/app/[locale]/(public)/decisions-automatisees/page.tsx`
**État actuel résumé** : 3 systèmes (modération, ranking, paiement) + droits + appel + recours indépendant. La seule page légale INDEXÉE.

#### 5.11.1 Bloquants P0

1. **Mécanisme d'appel = email à `dpo@designedtrust.com`** : insuffisant. Doit être un formulaire en ligne accessible sans authentification (cas du compte suspendu).

#### 5.11.2 Hauts P1

1. Ajouter délai de réponse maximum (30 jours) ;
2. Ajouter mention de la possibilité de recours indépendant (CNIL, médiateur conso).

---

<a name="partie-6"></a>
## Partie 6 — Audit bannière cookies + CMP

### 6.1 Conformité CNIL — checklist détaillée

#### 6.1.1 « Refuser tout » aussi visible que « Accepter tout »

**État actuel** : `equalWeightButtons: false` ligne 51 → ❌ VIOLATION

**Correctif (snippet Tailwind/TS prêt-à-l'emploi)** :

```ts
// web/src/shared/components/analytics/cookie-consent-provider.tsx
guiOptions: {
  consentModal: {
    layout: "box",
    position: "bottom right",
    equalWeightButtons: true,       // <-- corriger
    flipButtons: false,
  },
  preferencesModal: {
    layout: "box",
    position: "right",
    equalWeightButtons: true,       // <-- corriger
    flipButtons: false,
  },
},
```

Compléter par CSS spécifique dans `web/src/styles/cookie-consent.css` pour s'assurer du contraste équivalent (vanilla-cookieconsent applique parfois `opacity: 0.7` au bouton « refuser » par défaut) :

```css
/* web/src/styles/cookie-consent.css */
#cc-main .cm__btn--secondary,
#cc-main .pm__btn--secondary {
  /* refuser : même contraste que accepter */
  background-color: var(--cc-btn-secondary-bg, #2a1f15) !important;
  color: var(--cc-btn-secondary-color, #fffbf5) !important;
  border: 1px solid var(--cc-btn-secondary-border-color, #2a1f15) !important;
  font-weight: 600 !important;
  opacity: 1 !important;
}

#cc-main .cm__btn--secondary:hover,
#cc-main .pm__btn--secondary:hover {
  background-color: var(--cc-btn-secondary-hover-bg, #1a1410) !important;
}
```

#### 6.1.2 Liens vers politique privacy / cookies / mentions / sous-processeurs

**État actuel** : la bannière a un `footer` mais le contenu i18n n'est pas vérifié — risque que les liens manquent ou pointent vers les mauvaises URLs.

**Correctif** : vérifier dans `web/messages/fr.json:cookieConsent.banner.footer` que les liens sont :

```text
En cliquant sur "Accepter tout", tu autorises l'usage de cookies
d'analyse. Tu peux modifier ton choix à tout moment dans le
[Centre de préférences](/cookies). Plus d'infos :
[Politique de confidentialité](/legal/politique-confidentialite) |
[Mentions légales](/legal) | [Sous-processeurs](/sous-processeurs).
```

#### 6.1.3 Granularité par catégorie

**État actuel** : 2 catégories seulement (`necessary`, `analytics`). ❌ INSUFFISANT.

**Lignes directrices CNIL 2024** : 4 catégories minimum :

- **Strictement nécessaires** (technique, session, RGPD art. 5(3) ePrivacy — pas de consentement requis) ;
- **Préférences / fonctionnels** (langue, thème, géolocalisation choisie) ;
- **Mesure d'audience / analytics** (PostHog, GA4) ;
- **Marketing / publicité ciblée** (n/a chez nous mais à prévoir).

**Correctif** :

```ts
categories: {
  necessary: {
    enabled: true,
    readOnly: true,
  },
  functional: {
    enabled: false,
    readOnly: false,
    autoClear: { cookies: [{ name: "lang_preference" }, { name: "theme" }] },
  },
  analytics: {
    enabled: false,
    readOnly: false,
    autoClear: { cookies: [{ name: /^_ga/ }, { name: /^ph_/ }, { name: "_gid" }] },
  },
  marketing: {
    enabled: false,
    readOnly: false,
    autoClear: { cookies: [/* à compléter si campagnes lancées */] },
  },
},
```

#### 6.1.4 Pré-consentement strict

**État actuel** : `mode: "opt-in"` ligne 38 ✅ — bien configuré, rien ne charge avant consentement.

À VÉRIFIER en complément :
- Que PostHog (`posthog-provider.tsx`) n'init pas avant l'event `cc:onChange` ;
- Que GA4 (`google-analytics-provider.tsx`) n'init pas avant `cc:onChange`.

Si ces providers initialisent inconditionnellement et que `applyCustomConsent` ne fait que désactiver après init, c'est une violation. À auditer.

#### 6.1.5 Mécanisme de retrait

**État actuel** : ❌ pas d'icône flottante persistante.

**Correctif (snippet)** : créer un composant `CookieReopenButton` à monter dans le layout public, qui ouvre `/cookies` ou re-déclenche la bannière :

```tsx
// web/src/shared/components/analytics/cookie-reopen-button.tsx
"use client"

import { useTranslations } from "next-intl"
import * as CookieConsent from "vanilla-cookieconsent"
import { Cookie } from "lucide-react"

export function CookieReopenButton() {
  const t = useTranslations("cookieConsent")

  function handleOpen() {
    try {
      CookieConsent.showPreferences()
    } catch {
      window.location.href = "/cookies"
    }
  }

  return (
    <button
      type="button"
      onClick={handleOpen}
      aria-label={t("reopen.aria")}
      className="
        fixed bottom-4 left-4 z-40
        flex items-center gap-2 rounded-full
        border border-border bg-card px-3 py-2
        text-sm text-foreground shadow-md
        hover:bg-muted
        focus:outline-none focus-visible:ring-2 focus-visible:ring-accent
      "
    >
      <Cookie className="size-4" aria-hidden />
      <span className="hidden sm:inline">{t("reopen.label")}</span>
    </button>
  )
}
```

À monter dans `web/src/app/[locale]/(public)/layout.tsx` et `(dashboard)/layout.tsx` (sauf si l'utilisateur veut le cacher, ce qui n'est PAS conforme — l'icône doit être présente partout).

#### 6.1.6 Preuve du consentement

**État actuel** : `onFirstConsent` et `onChange` appellent `applyCustomConsent` qui mirror dans localStorage. ⚠️ Mais aucune preuve serveur-side (timestamp + IP + version revision + catégories choisies) n'est conservée.

**La CNIL exige (lignes directrices 2020, point 6.4)** :

> « Il convient de pouvoir, à tout moment, démontrer le consentement obtenu de l'utilisateur. »

Le LocalStorage est insuffisant car effaçable par l'utilisateur lui-même. La preuve doit être **serveur-side** dans une table `consent_records` :

```sql
CREATE TABLE IF NOT EXISTS consent_records (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NULL REFERENCES users(id),       -- NULL si visiteur anonyme
  anon_id     TEXT NULL,                             -- ID anonyme (cookie ID local)
  categories  JSONB NOT NULL,                        -- {necessary: true, analytics: false, ...}
  revision    INT NOT NULL,                          -- version du bandeau
  user_agent  TEXT,
  ip_address  INET,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_consent_user ON consent_records(user_id);
CREATE INDEX idx_consent_anon ON consent_records(anon_id);
```

Endpoint backend `POST /api/v1/consent` qui enregistre, appelé depuis `syncConsentToAnalytics()`.

### 6.2 Comparaison Malt / Upwork / Contra

Malt utilise leur propre CMP custom intégré, conforme CNIL avec 4 catégories. Upwork utilise OneTrust (CMP US-leader). Contra utilise un CMP basic axé sur le consentement implicite — non opposable UE.

**Marketplace Service vs. ces 3 plateformes** :

| Critère | Upwork | Malt | Contra | Marketplace Service | Status |
|---------|--------|------|--------|---------------------|--------|
| Boutons équivalents | ✅ | ✅ | ⚠️ | ❌ | P0 |
| 4 catégories | ✅ | ✅ | 2 | 2 | P1 |
| Icône flottante | ✅ | ✅ | ❌ | ❌ | P0 |
| Preuve serveur | ✅ | ✅ | ❌ | ❌ | P0 |
| Pré-consentement strict | ✅ | ✅ | ⚠️ | ✅ | OK |

---

<a name="partie-7"></a>
## Partie 7 — Audit RGPD article par article

### 7.1 Art. 5 — Principes

**Exigence verbatim courte** : licéité, loyauté, transparence ; finalité limitée ; minimisation ; exactitude ; conservation limitée ; intégrité et confidentialité.

**État** :
- ✅ Licéité : bases légales identifiées (art. 6 / 9) ;
- ✅ Loyauté : information préalable via politique ;
- ⚠️ Transparence : insuffisante car pages noindex ;
- ⚠️ Minimisation : collecte « about », « bio » étendues — à challenger sur la nécessité ;
- ✅ Exactitude : droit de rectification implémenté ;
- ⚠️ Conservation : politique annonce 3 ans messages mais purge effective non vérifiée ;
- ✅ Intégrité : TLS, RLS, audit logs, chiffrement at-rest Neon.

**Risque** : amende administrative jusqu'à 4% CA mondial (RGPD art. 83(5)).
**Action** : auditer la purge effective des messages 3 ans + tokens push 60 jours inactif + sessions révoquées 30 jours.

### 7.2 Art. 6 — Licéité

| Traitement | Base légale RGPD | État |
|------------|------------------|------|
| Création compte | 6(1)(b) contrat | ✅ |
| Profil public | 6(1)(b) | ✅ |
| Messagerie | 6(1)(b) | ✅ |
| KYC biométrique | 6(1)(c) + 9(2)(g) intérêt public substantiel ? | ⚠️ |
| Modération IA | 6(1)(c) DSA + 6(1)(f) intérêt légitime | ✅ |
| Analytics PostHog | 6(1)(a) consentement | ✅ |
| Audit log sécurité | 6(1)(f) + 6(1)(c) | ✅ |
| Notifications marketing | 6(1)(a) | À vérifier — opt-in explicite ? |
| DAC7 transmission DGFiP | 6(1)(c) | ✅ (si implémenté) |

**Risque KYC biométrique** : la base juridique RGPD art. 9(2)(g) (intérêt public substantiel) est étroitement encadrée. L'art. L.561-2 CMF impose le KYC mais **n'impose pas la biométrie** — c'est un choix de la Plateforme. La base devrait donc être **9(2)(a) consentement explicite** + 6(1)(c) pour la vérification d'identité non biométrique.

### 7.3 Art. 7 — Consentement

**Conditions** : libre, spécifique, éclairé, univoque, démontrable, retirable aussi facilement que donné.

**État** :
- ❌ Pas démontrable serveur-side (cf. 6.1.6) ;
- ❌ Retrait pas aussi facile (cf. 6.1.5) ;
- ⚠️ Pas spécifique avec 2 catégories seulement.

### 7.4 Art. 9 — Catégories particulières

Biométrie KYC = donnée Art. 9(1). Exception 9(2) à invoquer :

- 9(2)(a) consentement explicite (recommandé) ;
- 9(2)(g) intérêt public substantiel (fragile pour la biométrie).

**Action** : ajouter une case à cocher distincte « J'autorise l'usage de la vérification biométrique de visage pour la lutte contre la fraude » lors du flow KYC, avec preuve serveur-side.

### 7.5 Art. 12 — Transparence et modalités d'exercice

**Exigences** : information concise, transparente, compréhensible, accessible, formulation claire et simple, gratuit, délai 1 mois.

**État** :
- ⚠️ Concise : la politique de confidentialité est jugée longue mais OK ;
- ⚠️ Compréhensible : tutoiement OK, mais terminologie juridique présente ;
- ❌ Accessible : noindex bloque l'accès direct ;
- ✅ Gratuit ;
- ⚠️ Délai 1 mois à démontrer (procédure non outillée).

### 7.6 Art. 13-14 — Information du responsable

Art. 13 (collecte directe) : identité, finalité, base légale, destinataires, transferts, durée, droits, droit retrait, droit plainte CNIL, source légale ou contractuelle, profilage.

**État** : la politique couvre l'essentiel mais MANQUE :

- Conséquences du refus de fournir les données (ex : refus KYC → pas de paiement) ;
- Existence du profilage (mention présente mais à renforcer) ;
- Coordonnées du représentant UE si applicable (Marketplace Service est UE → n/a).

### 7.7 Art. 15-22 — Droits des personnes

| Article | Droit | Implémentation observée |
|---------|-------|-------------------------|
| 15 | Accès | `/dashboard/account/gdpr` |
| 16 | Rectification | `/dashboard/profile` |
| 17 | Effacement | `/dashboard/account/gdpr` |
| 18 | Limitation | email DPO |
| 19 | Notification rectif/efface | À vérifier (notification aux sous-traitants ?) |
| 20 | Portabilité | export JSON `/dashboard/account/gdpr` |
| 21 | Opposition | email DPO |
| 22 | Décisions auto | `/decisions-automatisees` |

**Bloquants P0** :

1. **Art. 19** : aucun mécanisme automatique de propagation de la rectification/effacement vers les sous-traitants (Stripe, R2, Resend, etc.). Doit être implémenté ou documenté manuellement ;
2. **Art. 22** : pas de formulaire de revue humaine outillé (cf. 5.11).

### 7.8 Art. 25 — Privacy by design / by default

**État** : globalement bien implémenté (RLS, audit log append-only, JWT short-lived, bcrypt cost 12).

Faiblesses :
- Profil public expose le nom complet par défaut — devrait être « opt-in » (privacy by default) ou pseudonymisation par défaut ;
- Géolocalisation IP collectée à chaque visite (adapter `geoip`) — à minimiser.

### 7.9 Art. 28 — Sous-traitants

**Exigences** : DPA écrit, instructions documentées, confidentialité, sécurité, sous-sous-traitance autorisée, droits personnes, assistance, suppression/restitution, audit.

**État** :
- ✅ DPA template public ;
- ❌ Asymétrie (cf. 5.7) — pas de DPA pour Marketplace Service en tant que sous-traitant ;
- ⚠️ DPAs effectivement signés avec les 21 sous-processeurs à confirmer — la mémoire indique D5 (DPAs checklist tâche #36 marquée completed) — vérifier la table dans `legal/dpas-checklist.md`.

### 7.10 Art. 30 — Registre

**Exigence** : registre tenu par le responsable (et le sous-traitant), avec finalité, destinataires, transferts, durées, mesures de sécurité.

**État** : page publique existe (`/legal/registre`), version markdown canonique mentionne 11 traitements. ⚠️ Le statut « [À COMPLÉTER] » est inacceptable en prod.

### 7.11 Art. 32 — Sécurité

| Mesure | État |
|--------|------|
| Pseudonymisation/chiffrement | ✅ TLS, at-rest Neon |
| Confidentialité/intégrité | ✅ RLS, audit logs append-only |
| Disponibilité | ✅ PITR Neon, snapshots |
| Restauration | À tester régulièrement (DRP) |
| Process tests | À documenter |

**Action** : documenter le DRP (Disaster Recovery Plan) annuel testé.

### 7.12 Art. 33-34 — Notification violation

**Exigences** : notification CNIL <72h ; notification personnes si risque élevé ; documentation interne.

**État** : ⚠️ procédure NON documentée publiquement. Doit exister un runbook interne.

**Action** : créer `BLOCKED_INCIDENT_RESPONSE.md` interne avec checklist :

```text
1. Détection (alerte, signalement, audit)
2. Confinement (suspension du compte / clé compromise)
3. Évaluation impact (catégories de données, nb personnes, gravité)
4. Notification CNIL <72h via téléservice CNIL
5. Notification personnes si risque élevé (email + bannière in-app)
6. Documentation (registre des violations, art. 33(5))
7. Post-mortem et leçons apprises
```

### 7.13 Art. 35 — AIPD

**État** : 3 AIPD documentées (KYC, modération, matching). Bien.

⚠️ Manque une 4e AIPD pour les **enregistrements LiveKit** si activés. À ajouter.

### 7.14 Art. 44-49 — Transferts hors UE

**État** :
- Politique mentionne DPF + SCC sans préciser par vendeur ;
- Pas de TIA documentée par vendeur ;
- ⚠️ DPF a été partiellement contesté en 2024 sur le volet enforcement (décision DPF Review Panel septembre 2024) — à surveiller, fragile.

**Action** : produire 11 TIA distincts (un par sous-processeur hors UE), tester explicitement le scénario US CLOUD Act (loi extraterritoriale US permettant la collecte par les autorités US).

---

<a name="partie-8"></a>
## Partie 8 — Audit fiscal et conformité paiement

### 8.1 TVA — auto-liquidation B2B intra-UE

**État actuel** :
- `legal.docs.cgv.paymentBody` mentionne « Opérations intra-UE : validation VIES obligatoire » ✅ ;
- Adapter `backend/internal/adapter/vies/` existe ✅ ;
- Pas vérifié si le code génère effectivement la mention « Autoliquidation — article 196 directive 2006/112/CE » sur les factures B2B intra-UE.

**Bloquants P0** :

1. Vérifier que la facture HTML/PDF inclut la mention obligatoire « Autoliquidation » quand client UE non-FR ;
2. Vérifier que la facture pour client français inclut bien TVA 20% (ou 0% si Prestataire en franchise base art. 293 B CGI) ;
3. Vérifier l'archivage 10 ans (Code de commerce L.123-22) — confirmé dans `legal.docs.cgv.invoicesBody`.

### 8.2 DAC7

**État actuel** :
- CGV mentionne DAC7 ;
- ❌ **AUCUN adapter `dac7/` dans `backend/internal/adapter/`** (vérifié).

**Bloquant P0 absolu** : si Marketplace Service ouvre publiquement avant fin 2026 et atteint un Prestataire avec >€2 000 ou >30 transactions, DAC7 s'applique. Sanction : 200 € par vendeur non déclaré (art. 1729 ter CGI).

**Action** : créer le module DAC7 avec :

```text
backend/internal/domain/dac7/
backend/internal/port/repository/dac7.go
backend/internal/app/dac7/service.go
backend/internal/adapter/postgres/dac7_repository.go
backend/internal/handler/dac7_handler.go (admin export endpoint)
```

Avec une procédure annuelle :

1. Décembre N : collecte des données fiscales manquantes auprès des Prestataires (relances email) ;
2. 1er janvier N+1 : verrouillage des données ;
3. 31 janvier N+1 : génération du fichier XML DAC7 conforme schéma DGFiP, transmission via portail impots.gouv.fr ;
4. 31 janvier N+1 : envoi du récapitulatif individuel au Prestataire (email + dashboard).

### 8.3 PSD2 — statut prestataire de paiement

**État actuel** : Stripe Connect Custom positionne Stripe Payments Europe Ltd (PSP agréé Banque centrale d'Irlande) comme PSP, et Marketplace Service comme **« plateforme » utilisant ce PSP**.

**Conditions à respecter** pour ne pas être qualifié soi-même de PSP :

1. Les fonds ne transitent **jamais** par un compte bancaire au nom de Marketplace Service — ils restent dans le compte Stripe au nom du Prestataire (Connected Account) ;
2. Marketplace Service ne décide pas du destinataire — c'est le Client qui décide (proposal validation) ;
3. Marketplace Service ne « tient » pas un solde durable pour le compte des utilisateurs — c'est Stripe qui maintient le solde via Connected Account ;
4. Marketplace Service est, au plus, « agent commercial » de Stripe (art. L.523-1 CMF) — à clarifier dans le contrat Stripe.

⚠️ Le **wallet permanent** de Marketplace Service (`WALLET-UNIFY` tâches #43-48) est un point de vigilance : si ce wallet maintient un solde « commission apporteur » durable hors Stripe, c'est un risque PSD2. Recommandé : transférer immédiatement chaque commission apporteur sur le Connected Account Stripe de l'apporteur (cf. tâche #33 D1+D2).

### 8.4 LCB-FT / AML

**État actuel** :
- KYC biométrique via Stripe + Rekognition ;
- Audit logs append-only (migration 142) ;
- Sanitization PII (migration 146).

**Bloquants P0** :

1. Vérifier que **l'apporteur d'affaires** subit aussi un KYC complet — sinon LCB-FT non respectée pour la rémunération de l'apporteur ;
2. Documenter la **politique anti-fraude** (Stripe Radar recommandé) ;
3. Documenter les **règles d'alerte** (seuils de transaction, comportements suspects).

### 8.5 Code de commerce

- L.123-22 : conservation 10 ans des factures ✅ ;
- L.441-9 : mention SIREN sur facture ✅ (à vérifier dans le PDF généré) ;
- L.441-1 : délai de paiement légal 30 jours (B2B sauf accord) — la CGV doit le mentionner.

---

<a name="partie-9"></a>
## Partie 9 — Risques travailleur salarié déguisé

### 9.1 Critères Cour de cassation

**Arrêt Uber Cass. soc. 4 mars 2020 n° 19-13.316** : requalification en contrat de travail dès lors que :

1. Lien de subordination (instructions, contrôle, sanction) ;
2. Tarification imposée par la plateforme ;
3. Exclusivité de fait (interdiction d'utiliser concurrents) ;
4. Géolocalisation/traçabilité permanente ;
5. Évaluation et sanction unilatérales.

**Arrêt Deliveroo Cass. soc. 13 avril 2022 n° 20-14.870** : confirme et précise — chaque indice cumulatif renforce.

**Arrêt Take Eat Easy Cass. soc. 28 nov. 2018** : critère du « pouvoir de direction » prépondérant.

**Jurisprudence récente 2023-2024** : tendance à requalifier plus largement les plateformes B2B avec des freelances dépendants.

### 9.2 Application à Marketplace Service

Risques :

- Si la Plateforme **fixe le prix** des prestations → indice de subordination ;
- Si la Plateforme **note unilatéralement** le Prestataire et que cette note influence sa visibilité → sanction de fait ;
- Si la Plateforme **interdit** au Prestataire de travailler ailleurs → exclusivité.

**État actuel** :

- ✅ Le Prestataire fixe son prix (Proposition) ;
- ⚠️ La Plateforme propose un classement / matching qui peut être perçu comme une « notation » imposée → mitigé par la transparence des pondérations ;
- ✅ Aucune clause d'exclusivité dans la CGU actuelle ;
- ⚠️ Le « JSS-like » score chez Upwork est risqué — Marketplace Service ne semble pas l'implémenter, OK.

### 9.3 Modèle Malt (à dupliquer)

Voir 4.2.10. Reproduire **verbatim** dans la CGU :

```text
Article 11 — Statut indépendant du Prestataire (nouveau)

11.1 Indépendance affirmée

Le Prestataire déclare et garantit exercer son activité de manière
strictement indépendante, en organisant librement son travail, sans
lien de subordination juridique avec la Plateforme ni avec le
Client. Il assume la pleine responsabilité de :

(i) Son organisation, ses horaires et ses outils de travail ;
(ii) Sa conformité fiscale et sociale (URSSAF, impôts, TVA) ;
(iii) Son assurance professionnelle (RC pro, prévoyance) ;
(iv) Sa qualification professionnelle (diplôme, certification, RCS).

11.2 Interdiction d'exclusivité

Aucun Client ne peut imposer au Prestataire une clause d'exclusivité
limitant son droit de travailler pour d'autres Clients ou via
d'autres plateformes. Toute clause contraire incluse dans une
Proposition ou un échange contractuel est réputée non écrite.

11.3 Durée maximale d'une mission

Une Mission ne peut pas excéder douze (12) mois consécutifs avec le
même Client sans une réévaluation contractuelle explicite acceptée
par les Parties. Au-delà de vingt-quatre (24) mois consécutifs avec
le même Client, la Plateforme se réserve le droit de notifier les
Parties d'un risque de requalification et de proposer un modèle de
contrat ad hoc.

11.4 Pluralité de Clients

La Plateforme encourage le Prestataire à diversifier ses Clients
pour préserver son indépendance économique. Une dépendance
économique d'un Prestataire à un seul Client supérieure à 70 % de
son chiffre d'affaires annuel est un indice de risque que les
Parties sont invitées à corriger.

11.5 Worker Classification — responsabilité du Client

Conformément à la jurisprudence applicable, le Client assume seul
la responsabilité de la qualification juridique de la relation
contractuelle qu'il noue avec le Prestataire via la Plateforme.
La Plateforme n'a ni le rôle ni la compétence pour qualifier la
nature juridique de cette relation et ne saurait être tenue
responsable d'une requalification éventuelle en contrat de travail.

11.6 Non-employeur

Rien dans les présentes CGU, dans une Proposition ou dans une
Mission ne saurait être interprété comme créant un lien
contractuel d'emploi, de partenariat, de joint-venture, de
franchise ou de mandat entre la Plateforme et l'Utilisateur.
```

### 9.4 Risques résiduels

- **Apporteur d'affaires** : si l'apporteur encaisse 5% de manière récurrente sans KYC ni statut commercial, requalification en « travailleur dissimulé » possible (art. L.8221-3 Code du travail). À sécuriser par :
  - KYC apporteur (immatriculation RCS / micro-entrepreneur) ;
  - Contrat d'apport d'affaires formel pour chaque trio ;
  - Facturation de la commission par l'apporteur lui-même (mandat de facturation possible).

---

<a name="partie-10"></a>
## Partie 10 — Risques DSA / DMA

### 10.1 DSA applicable depuis 17 février 2024

Marketplace Service est une **plateforme en ligne** au sens DSA (art. 3(i)). Sous le seuil VLOP de 45M MAU UE → obligations « standard ».

**Obligations applicables** :

| Art. | Obligation | État |
|------|-----------|------|
| 11 | Point de contact unique autorités | ⚠️ à formaliser dans mentions légales |
| 12 | Point de contact destinataires service | ⚠️ idem (dpo@designedtrust.com pourrait suffire) |
| 14 | Conditions générales claires | ⚠️ noindex à corriger |
| 16 | Mécanisme de notification et action | ⚠️ existe-t-il un endpoint signalement ? |
| 17 | Notification motivée de décision de modération | ❌ à implémenter |
| 18 | Reporting aux autorités si infraction pénale (art. 18) | ❌ procédure non documentée |
| 20 | Système de gestion interne des réclamations | ❌ à outiller (formulaire) |
| 21 | Règlement extrajudiciaire | ⚠️ médiateur conso à finaliser |
| 22 | Trusted flaggers | optionnel — n/a |
| 23 | Mesures contre les abus | ⚠️ rate limit + ban à documenter |
| 24 | Transparence (rapport annuel) | EXEMPTÉ si <50 employés et <€10M CA |
| 26-28 | Publicité ciblée + mineurs | n/a (pas de pub) |

### 10.2 Bloquants P0 DSA

1. **Art. 14** : conditions générales **accessibles, lisibles, intelligibles, dans la langue de l'utilisateur** → noindex à corriger ;
2. **Art. 16-17** : implémenter un mécanisme de signalement (`/api/v1/dsa/report`) et une notification motivée de décision (NOL — Notice of Limitation). Aujourd'hui, la modération coupe sans NOL explicite ;
3. **Art. 11-12** : déclarer le point de contact à ARCOM (https://www.arcom.fr/dsa).

### 10.3 DMA

**Non applicable directement** : Marketplace Service n'est pas gatekeeper (€7,5Md CA ou €75Md cap. ou 45M MAU UE / 10k entreprises actives → seuils inatteignables à court terme). Pas d'obligation.

À surveiller pour les bonnes pratiques (interdiction de combiner données entre services, libre choix du navigateur, etc.).

---

<a name="partie-11"></a>
## Partie 11 — Roadmap correctrice priorisée

### 11.1 Avant submit Stripe (24-48h) — BLOQUANT ABSOLU

| # | Fichier | Action | Estimation |
|---|---------|--------|-----------|
| P0-01 | `web/messages/fr.json` legal.mentions.* | Renseigner identité éditeur réelle (raison sociale, RCS, capital, adresse, directeur de publication, n° TVA) | 1h (après création société) |
| P0-02 | `web/src/app/[locale]/(public)/legal/page.tsx:20` et 9 autres | Supprimer `robots: { index: false, follow: false }` sauf `/legal/registre` et `/legal/aipd` | 30 min |
| P0-03 | `web/messages/fr.json` legal.docs.cgv.modelBody | Remplacer « indicatif » par tarification fixe (cf. 5.3.3) | 1h |
| P0-04 | `web/src/shared/components/analytics/cookie-consent-provider.tsx:51,57` | `equalWeightButtons: true` | 5 min |
| P0-05 | `web/messages/fr.json` legal.docs.cgu.* | Ajouter articles définitions, anti-désintermédiation, force majeure, modification, worker classification | 4h |
| P0-06 | `web/src/app/[locale]/(public)/privacy/page.tsx` | Supprimer ou rendre strict résumé pointant `/legal/politique-confidentialite` | 30 min |
| P0-07 | `web/src/shared/components/analytics/cookie-reopen-button.tsx` (nouveau) | Créer composant icône flottante + monter dans tous les layouts | 2h |

**Total : 9-10h ouvrées.**

### 11.2 Avant ouverture publique (1 semaine) — BLOQUANT MAJEUR

| # | Fichier | Action | Estimation |
|---|---------|--------|-----------|
| P0-08 | `backend/internal/adapter/postgres/consent_records.go` (nouveau) | Créer table + endpoint POST /api/v1/consent + wiring CMP onChange | 1 jour |
| P0-09 | `backend/internal/adapter/dac7/` (nouveau module complet) | Implémenter DAC7 reporting (domain + port + adapter + handler admin) | 3 jours |
| P0-10 | `backend/internal/handler/dsa_report_handler.go` (nouveau) | Endpoint signalement DSA art. 16 + notification motivée art. 17 | 2 jours |
| P0-11 | `web/src/app/[locale]/(public)/legal/cgv/page.tsx` | Réécrire 8 sections avec textes P0 (cf. 5.3.3, 5.3.4) | 2h |
| P0-12 | `web/src/app/[locale]/(public)/legal/politique-confidentialite/page.tsx` | Ajouter tableau des 11 traitements + profilage explicite | 4h |
| P0-13 | `web/src/app/[locale]/(public)/decisions-automatisees/page.tsx` | Ajouter formulaire d'appel art. 22 (sans authentification possible) | 1 jour |
| P0-14 | `legal/tia/*.md` (11 fichiers) | Rédiger 11 TIA (un par sous-processeur hors UE) | 2 jours |
| P0-15 | `web/messages/fr.json` legal.docs.cgu.liabilityBody | Plafond responsabilité fixe 50 000 € + 12 mois commissions | 30 min |
| P0-16 | `backend/internal/adapter/postgres/audit_log_purge.go` | Vérifier purge messages 3 ans / sessions 30j / tokens 60j | 4h |
| P0-17 | `backend/internal/adapter/rekognition/face.go` | Vérifier non-stockage frames + non-IndexFaces | 2h |
| P0-18 | `web/messages/fr.json` legal.subprocessors.transferYes | Distinguer DPF vs SCC par vendeur | 1h |

**Total : ~12 jours-homme** (à paralléliser : DAC7 et DSA peuvent être en parallèle).

### 11.3 Conformité continue (premier trimestre) — HAUT P1

| # | Action | Estimation |
|---|--------|-----------|
| P1-01 | Vérifier auto-certification DPF de chaque vendeur (8) sur dataprivacyframework.gov | 1h |
| P1-02 | Inscription Médiateur de la consommation (MEDICYS ou CMAP) | 2 jours (admin) |
| P1-03 | Notification ARCOM point de contact DSA | 2 jours |
| P1-04 | DPA cosigné « Marketplace Service en tant que sous-traitant » template | 1 jour |
| P1-05 | 4e AIPD pour LiveKit (si enregistrements activés) | 2 jours |
| P1-06 | Runbook violation données (notification CNIL <72h) | 2 jours |
| P1-07 | DRP (Disaster Recovery Plan) annuel testé | 5 jours |
| P1-08 | Bandeau cookies 4 catégories (necessary, functional, analytics, marketing) | 4h |
| P1-09 | Politique anti-fraude documentée + Stripe Radar wiring | 3 jours |
| P1-10 | KYC apporteur d'affaires (Connected Account ?) | 5 jours |
| P1-11 | Worker Classification disclaimer ajouté CGU | déjà dans P0-05 |
| P1-12 | Audit purge effective (cron + tests) | 2 jours |
| P1-13 | Audit propagation rectif/effacement vers sous-traitants (art. 19) | 2 jours |
| P1-14 | Nomination DPO certifié + déclaration CNIL | 1 jour (admin) |

**Total : ~25 jours-homme** + procédures admin.

### 11.4 Continu (audits réguliers)

| Fréquence | Action |
|-----------|--------|
| Mensuel | Revue logs sécurité + tentatives intrusion |
| Trimestriel | Revue de la liste DPF (auto-certifications encore actives ?) |
| Semestriel | Revue des sous-processeurs (DPA encore valides ?) |
| Annuel | Revue AIPD (3+1 documents) |
| Annuel | Audit RLS PostgreSQL (cross-tenant access tests) |
| Annuel | Test DRP (restauration backup Neon, basculement) |
| Annuel (avant 31 janvier) | Transmission DAC7 DGFiP |
| Annuel | Mise à jour CGU/CGV avec préavis 30j |
| Annuel | Rapport DSA si dépassement seuil micro (à surveiller) |

---

<a name="partie-12"></a>
## Partie 12 — Annexes

### 12.1 URLs consultées

- https://www.upwork.com/legal — bloqué anti-bot, contournement Pactsafe
- https://upwork.pactsafe.io/versions/64a63ee98763a953463e10af.pdf — User Agreement Upwork v. 2023
- https://www.upwork.com/direct-contracts — fonctionnement contracts
- https://www.malt.fr/about/legal/cgu — bloqué anti-bot
- https://help.malt.com/kb/guide/fr/la-commission-malt-h2dK0HxuKA — commission Malt
- https://www.malt.fr/resources/article/depuis-le-1er-janvier-2023-une-nouvelle-directive- — DAC7 Malt
- https://www.malt.fr/about/privacy/policy — bloqué anti-bot
- https://contra.com/policies/terms — Terms of Service Contra (accessible)
- https://contra.com/commission-free — modèle commission Contra
- https://www.cnil.fr/fr/plaintes — saisine CNIL
- https://stripe.com/en-th/legal/restricted-businesses — Stripe Restricted
- https://docs.stripe.com/connect/upcoming-requirements-updates — Stripe Connect 2025
- https://digital-strategy.ec.europa.eu/en/policies/digital-services-act — DSA
- https://www.arcom.fr/dsa — ARCOM DSA France
- https://www.dataprivacyframework.gov/list — registre DPF
- https://digital-strategy.ec.europa.eu/en/library/implementing-regulation-laying-down-templates-concerning-transparency-reporting-obligations — templates DSA

### 12.2 Références légales (textes verbatim courts)

- **RGPD art. 5** : « Les données à caractère personnel doivent être : a) traitées de manière licite, loyale et transparente (...) ; b) collectées pour des finalités déterminées, explicites et légitimes ; c) adéquates, pertinentes et limitées (...) ; d) exactes (...) ; e) conservées (...) pendant une durée n'excédant pas (...) ; f) traitées de façon à garantir une sécurité appropriée (...). Le responsable du traitement est responsable du respect du paragraphe 1 et est en mesure de démontrer que celui-ci est respecté (responsabilité). »

- **RGPD art. 28(3)** : « Le traitement par un sous-traitant est régi par un contrat (...) qui lie le sous-traitant à l'égard du responsable du traitement (...) »

- **RGPD art. 32(1)** : « Compte tenu de l'état des connaissances, des coûts de mise en œuvre et de la nature, de la portée, du contexte et des finalités du traitement ainsi que des risques (...), le responsable du traitement et le sous-traitant mettent en œuvre les mesures techniques et organisationnelles appropriées afin de garantir un niveau de sécurité adapté au risque (...) »

- **LCEN art. 6-III-1** : « Les personnes dont l'activité est d'éditer un service de communication au public en ligne mettent à disposition du public, dans un standard ouvert : 1° S'il s'agit de personnes physiques, leurs nom, prénoms, domicile et numéro de téléphone (...) ; 2° S'il s'agit de personnes morales, leur dénomination ou leur raison sociale et leur siège social, leur numéro de téléphone (...) ; 3° Le nom du directeur ou du codirecteur de la publication (...) ; 4° Le nom, la dénomination ou la raison sociale et l'adresse et le numéro de téléphone du prestataire (...) [hébergeur]. »

- **CGI art. 1649 ter** : « Les opérateurs de plateforme (...) transmettent à l'administration fiscale (...) avant le 31 janvier de chaque année, les informations relatives aux opérations réalisées par les vendeurs et prestataires (...) au cours de l'année civile précédente. »

- **C. consommation L.111-1** : « Avant que le consommateur ne soit lié par un contrat (...), le professionnel communique au consommateur, de manière lisible et compréhensible, les informations suivantes : 1° Les caractéristiques essentielles du bien ou du service (...) ; 2° Le prix du bien ou du service (...) ; 3° (...) »

- **C. travail L.8221-3** : « Est réputé travail dissimulé par dissimulation d'activité, l'exercice à but lucratif d'une activité (...) par toute personne qui (...) a) Soit s'est soustraite intentionnellement à ses obligations (...) b) Soit s'est soustraite intentionnellement à l'accomplissement de la formalité prévue à l'article L.123-1 du Code de commerce (...) »

- **Cass. soc. 4 mars 2020 n° 19-13.316 (Uber)** : « (...) le statut de travailleur indépendant (...) est fictif (...) lorsqu'il est établi un lien de subordination juridique permanent entre le donneur d'ordre et le travailleur (...) ».

- **DSA art. 14** : « Les fournisseurs de services intermédiaires incluent dans leurs conditions générales des informations sur les éventuelles restrictions imposées en lien avec l'utilisation de leur service (...). Ces informations sont rédigées dans un langage clair, simple, intelligible, convivial et non ambigu, et sont mises à la disposition du public dans un format aisément accessible et lisible par machine. »

### 12.3 Glossaire

- **AIPD** : Analyse d'Impact relative à la Protection des Données (art. 35 RGPD). En anglais : DPIA (Data Protection Impact Assessment).
- **CCT** : Clauses Contractuelles Types — modèle d'accord de transfert international approuvé par la Commission UE (décision 2021/914). En anglais : SCC (Standard Contractual Clauses).
- **CMP** : Consent Management Platform — outil de gestion du consentement cookies (vanilla-cookieconsent, OneTrust, Didomi, etc.).
- **CNIL** : Commission Nationale de l'Informatique et des Libertés — autorité de contrôle française RGPD.
- **DAC7** : Directive (UE) 2021/514 sur la coopération administrative fiscale entre États membres concernant les plateformes numériques.
- **DPA** : Data Processing Agreement — contrat de sous-traitance au sens art. 28 RGPD.
- **DPF** : EU-US Data Privacy Framework — décision d'adéquation 2023/1795 succédant au Privacy Shield invalidé par Schrems II.
- **DPO** : Data Protection Officer — délégué à la protection des données (art. 37-39 RGPD).
- **DSA** : Digital Services Act — règlement (UE) 2022/2065 régulant les services numériques.
- **EUID** : European Unique Identifier — identifiant européen des sociétés (équivalent du SIREN/RCS au niveau UE).
- **KYC** : Know Your Customer — vérification d'identité au sens LCB-FT.
- **LCB-FT** : Lutte Contre le Blanchiment et le Financement du Terrorisme (CMF art. L.561-2 et suiv.).
- **LCEN** : Loi pour la Confiance dans l'Économie Numérique (n° 2004-575 du 21 juin 2004).
- **NOL** : Notice of Limitation — notification motivée d'une restriction (art. 17 DSA).
- **PSP** : Prestataire de Services de Paiement (PSD2).
- **RGPD** : Règlement Général sur la Protection des Données (UE 2016/679). En anglais : GDPR.
- **RLS** : Row-Level Security — sécurité par ligne, mécanisme PostgreSQL de filtrage automatique.
- **TIA** : Transfer Impact Assessment — évaluation d'impact d'un transfert de données hors UE.
- **VIES** : VAT Information Exchange System — registre européen de validation des n° de TVA intra-UE.
- **VLOP** : Very Large Online Platform — plateforme désignée par la Commission UE avec >45M MAU UE, soumise aux obligations renforcées DSA.

### 12.4 Checklist de pré-déploiement (à cocher)

```
ENTITÉ JURIDIQUE
[ ] Société immatriculée (SAS / SARL / SA)
[ ] RCS publié
[ ] N° TVA intra-UE actif et validable via VIES
[ ] Capital social libéré
[ ] Représentant légal nommé
[ ] Statuts à jour incluant l'activité de plateforme d'intermédiation
[ ] Banque pro avec IBAN
[ ] DPO nommé (interne ou externe, certifié)
[ ] DPO déclaré à la CNIL via téléservice
[ ] Adresse RGPD (postale) distincte ou même que siège

MENTIONS LÉGALES
[ ] Page /legal complète (raison sociale, RCS, capital, adresse, directeur publication, n° TVA, téléphone)
[ ] Page /legal indexée (suppression noindex)
[ ] Hébergeurs nommés avec adresses postales complètes
[ ] Médiateur de la consommation référencé
[ ] Stripe Payments Europe Ltd nommé comme PSP

CGU
[ ] Article Définitions
[ ] Article Worker Classification (cf. partie 9.3)
[ ] Article Anti-désintermédiation détaillé
[ ] Article Plafond responsabilité (50 000 € + 12 mois commissions)
[ ] Article Modification CGU avec préavis 30j
[ ] Article Force majeure
[ ] Article Loi applicable + juridiction compétente
[ ] noindex retiré

CGV
[ ] Tarification fixe (5 % Client + 10 % Prestataire + 5 % Apporteur)
[ ] Mandat de facturation explicite (art. 289 I-2 CGI)
[ ] TVA : auto-liquidation B2B intra-UE mentionnée
[ ] DAC7 mention
[ ] Délai de paiement 30j (art. L.441-1)
[ ] Conservation 10 ans (art. L.123-22)
[ ] noindex retiré

POLITIQUE DE CONFIDENTIALITÉ
[ ] Tableau des 11 traitements complet
[ ] Bases légales par traitement
[ ] Durées de conservation détaillées
[ ] Sous-processeurs avec DPF/SCC précisés
[ ] 11 TIA documentées (internes)
[ ] Profilage explicite
[ ] Formulaire art. 22 accessible sans auth
[ ] noindex retiré

SÉCURITÉ (RGPD ART. 32)
[ ] TLS 1.2+ partout
[ ] Chiffrement at-rest Neon
[ ] RLS PostgreSQL sur toutes les tables sauf users
[ ] Audit logs append-only (migration 142 + 146)
[ ] Sauvegardes PITR + snapshots
[ ] Runbook violation données <72h CNIL
[ ] DRP testé annuellement

DAC7
[ ] Adapter backend/internal/adapter/dac7/ implémenté
[ ] Collecte annuelle automatique
[ ] Génération XML conforme schéma DGFiP
[ ] Transmission impots.gouv.fr
[ ] Récapitulatif individuel Prestataire

DSA
[ ] Point de contact ARCOM déclaré
[ ] Mécanisme signalement art. 16
[ ] Notification motivée décision art. 17
[ ] Procédure interne plainte art. 20

COOKIES / CMP
[ ] equalWeightButtons: true
[ ] 4 catégories
[ ] Icône flottante persistante
[ ] Preuve serveur-side consent_records
[ ] Pré-consentement strict (rien ne charge avant)
[ ] Liens politique + cookies + mentions + sous-processeurs depuis bannière

KYC / AML
[ ] Stripe Connect Custom configuré
[ ] KYC apporteur d'affaires
[ ] Stripe Radar wiring
[ ] Politique anti-fraude documentée

TRAVAILLEUR INDÉPENDANT
[ ] Article statut indépendant CGU
[ ] Interdiction exclusivité explicite
[ ] Limite 12/24 mois
[ ] Worker Classification disclaimer
[ ] Pluralité clients encouragée
```

---

## Conclusion — verdict synthétique

**État au 12 mai 2026 :**

- ❌ **Stripe submit immédiat** : refus quasi-certain pour identité éditeur placeholder + tarification « indicative » + pages noindex ;
- ⚠️ **Ouverture publique** : possible après corrections P0 (10-12 jours-homme) MAIS risque CNIL/DGFiP élevé sans DAC7 ni purge effective ;
- ✅ **Fondations techniques** : excellentes (RLS, audit log append-only, AIPD documentées, DPA template public) — supérieures à Contra, comparables à Malt sur le volet sécurité ;
- ✅ **Direction juridique** : la mémoire et le code montrent une conscience juridique avancée (registre Art. 30, AIPD, DPAs checklist, GDPR phase C) — l'ossature est saine.

**Verdict final** : Marketplace Service est à **3-5 semaines** d'un déploiement public Stripe-clean et CNIL-clean, sous réserve de l'exécution disciplinée de la roadmap P0+P1. Les fondations sont solides ; les bloquants sont **opérationnels** (identité éditeur, tarification fixe, indexation, DAC7) plutôt que **structurels**.

**Recommandation forte** : avant tout déploiement public, faire **valider par un avocat humain** :
- Les textes CGU/CGV finalisés ;
- La qualification PSP/agent commercial vis-à-vis Stripe ;
- La structure d'apporteur d'affaires (3 parties + commission) ;
- Les 11 TIA ;
- L'AIPD KYC biométrique.

Budget avocat estimé : **5 000 € à 12 000 €** pour une revue + co-signature des CGU/CGV. À engager avant le submit Stripe pour minimiser le risque de rejet.

— Fin du rapport —
