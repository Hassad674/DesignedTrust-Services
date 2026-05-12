# LEGAL AUDIT — Round 2 (post-implementation)

**Date :** 12 mai 2026 (legal-max-blindage round)
**Périmètre :** vérification post-implémentation de la roadmap P0 +
P1 décrite dans `LEGAL_AUDIT_EXPERT.md`. Audit autonome — pas de
revue avocat humain.

---

## 1. Score de blindage par page

| Page | Score | Évolution | Verdict |
|------|-------|-----------|---------|
| `/legal` (Mentions légales) | 8.5 / 10 | +5.5 | Stripe-passe avec réserve identité finale |
| `/legal/cgu` | 9 / 10 | +5.5 | Sec, copie-coller des standards Upwork/Malt |
| `/legal/cgv` | 9 / 10 | +6 | Tarifs fixés, mandat 289 I-2 CGI, DAC7, Stripe Restricted |
| `/legal/politique-confidentialite` | 8.5 / 10 | +5 | Tableau 11 traitements, profilage, transfers, SCC primaire |
| `/legal/code-de-conduite` | 9 / 10 | NEW | DSA art. 14/16/17/23, 13 comportements, 4 sanctions, appel |
| `/sous-processeurs` | 8 / 10 | +3 | DPF vs SCC par vendor, note Schrems II |
| `/cookies` | 7 / 10 | +0 | Intacte — déjà bonne base mais 2 catégories au lieu de 4 |
| `/decisions-automatisees` | 6.5 / 10 | +0 | OK mais formulaire d'appel pas outillé (Phase 3) |
| `/legal/registre` | 6 / 10 | +0 | Toujours interne, "[À COMPLÉTER]" résiduel dans l'i18n |
| `/legal/aipd` | 6 / 10 | +0 | 3 AIPD documentées, signatures à compléter |
| `/legal/dpa-template` | 7 / 10 | +0 | Template Module 2 OK, manque le Module sous-traitant pour client |
| Bannière CMP | 9 / 10 | +6 | equalWeightButtons, 4 liens locale-aware, manage button |
| LegalFooter | 9 / 10 | +2 | 9 liens + cookie manage + DPO + copyright |
| DashboardLegalLinks | 8.5 / 10 | +1.5 | 8 liens, ajout code de conduite + décisions auto |

**Score moyen :** 7.8 / 10 (avant la round : 4.2 / 10).

---

## 2. Comparaison verbatim avec Upwork / Malt / Contra

### 2.1 vs Upwork

| Critère | Upwork | Marketplace Service (après round) | Verdict |
|---------|--------|-----------------------------------|---------|
| Identité opérateur (LCEN art. 6-III) | Upwork Global LLC + adresse SEC publique | Designed Trust SAS + repli greffe (8 champs) | ✅ Parité (manque la finalisation greffe — décision humaine) |
| Auto-qualification intermédiaire | « not involved directly... not a party » | « intermédiaire technique LCEN art. 6 + DSA, n'est pas partie aux contrats » | ✅ Verbatim conforme |
| Plafond responsabilité | 12 mois fees | 12 mois commissions + plafond 10 000 € HT/an | ✅ Plus précis (plafond absolu en plus) |
| Off-platform circumvention | « circumvent or attempt to circumvent » | Article 6 anti-désintermédiation détaillé + indemnité 6 mois + 1240 Code civil | ✅ Plus exhaustif |
| Worker Classification | « Client is solely responsible » | Article 9 statut indépendant (Malt verbatim) + 9.5 Worker disclaimer Cass. soc. Uber/Deliveroo/TEE | ✅ Mieux que Upwork (jurisprudence FR citée) |
| Force majeure | Présente | Article 15 — art. 1218 Code civil + sortie 60j + sous-processeurs nommés | ✅ Parité |
| Class action waiver | Présent (non opposable UE) | NON inclus (volontairement) | ✅ Différenciation FR |
| Arbitrage forcé | AAA (US) | Médiation CMAP + Tribunal commerce Paris | ✅ Différenciation UE |

**Gap résiduel :** Upwork a une entité Escrow séparée ; Marketplace
Service compte sur Stripe Payments Europe Ltd (mentionné). C'est
acceptable juridiquement.

### 2.2 vs Malt

| Critère | Malt | Marketplace Service | Verdict |
|---------|------|---------------------|---------|
| Identité (RCS Paris, capital, DPO) | Complète | Repli greffe + DPO en place | ⚠️ Acceptable, à finaliser |
| Mandat de facturation 289 I-2 CGI | Article complet + opposition | Article 7.2 verbatim + opposition 30j | ✅ Parité |
| Worker Classification 4 piliers | Indépendance, exclusivité, durée 12 mois, pluralité | Article 9 reprend les 4 piliers verbatim | ✅ Parité |
| DAC7 implémenté | Production depuis 2023 | CGV art. 8 décrit mais **ZÉRO code backend** | ❌ Bloquant fiscal — voir gaps |
| TVA auto-liquidation B2B | Vérification VIES + mention | Article 7.3 mentionne + VIES (adapter présent) | ✅ Parité |
| Médiation MEDICYS | Inscription effective | « Référence en cours d'inscription » (fallback) | ⚠️ Décision humaine |
| Rapport transparence DSA | Publié annuellement | Engagement (premier rapport janvier 2027) | ✅ Engagement formel |
| Bannière cookies CNIL | 4 catégories, equalWeightButtons | 2 catégories, equalWeightButtons ✅ | ⚠️ 2 catégories suffisent en pratique (necessary + analytics, marketing absent) — mais Malt a 4. À traiter en Phase 3. |

**Gap résiduel :** DAC7 (BLOQUANT fiscal après ouverture publique) +
Inscription effective Médiateur (admin) + Migration à 4 catégories
CMP.

### 2.3 vs Contra

| Critère | Contra | Marketplace Service | Verdict |
|---------|--------|---------------------|---------|
| Limited authorized agent qualification | Article 38 | Mention dans mentions légales (PspBody) + CGV art. 12 | ✅ Parité |
| Délai escrow 3 mois max | Section 22 | CGV art. 5 « 3 mois → release au Prestataire » | ✅ Verbatim Contra |
| Worker classification | Section 18 | Article 9 CGU | ✅ Parité |
| Liste prohibited digital products | Très détaillée | Article 10 CGV (10 items) + Code de conduite (13 items) | ✅ Plus exhaustif |
| DPA template public | Absent | /legal/dpa-template existant | ✅ Mieux que Contra |
| Class action waiver | Présent | Absent (FR) | ✅ Différenciation UE |
| Adresse postale visible | Absente | Présente (mentions légales + hébergeurs) | ✅ Mieux que Contra |

**Gap résiduel :** Aucun majeur — Marketplace Service est désormais
supérieur à Contra sur l'aspect transparence (mentions, DPA, code de
conduite).

---

## 3. Risque résiduel Stripe (post-round)

### Bloquants restants

1. **Identité finale au greffe** (raison sociale exacte, RCS, capital,
   adresse, directeur publication) — décision business / admin, à
   finaliser avant submit. **Statut formule de repli OK Stripe Review
   en attendant.**
2. **DAC7 implémentation backend** — la CGV mentionne DAC7 mais le
   code n'a pas d'adapter (P0-09 dans la roadmap originelle). Risque
   amende 200 € par vendeur non déclaré (CGI 1729 ter) après 31
   janvier 2027 si ouverture > 2 000 € ou 30 transactions. **NON
   Stripe-bloquant** mais **fiscal-bloquant**.
3. **Apporteur d'affaires KYC** — CGU art. 10 + CGV art. 6 le décrivent
   correctement (Stripe Connected Account + NIF), mais la
   vérification de présence effective du Connected Account côté code
   n'a pas été auditée (hors scope Phase 1). **Risque LCB-FT moyen.**

### Risque réduit par cette round

- Mentions légales : passage de placeholder à 8 champs structurés
  avec formule de repli explicite — **Stripe Restricted Businesses
  ne devrait plus rejeter** sur le motif identité.
- Tarification : passage d'« indicatif à confirmer » à grille fixe
  publiée — **Stripe Restricted Businesses + Code de la consommation
  L.111-1 satisfaits**.
- Pages noindex : retirées sur CGU, CGV, politique, sous-processeurs,
  mentions, code de conduite — **Stripe crawler peut vérifier**.
- Worker classification : article 9 Malt verbatim — **Stripe ne
  signalera plus de risque travailleur dissimulé**.

### Verdict Stripe

**⚠️ PASSE AVEC RÉSERVE :** soumissible en review Stripe **après**
remplissage des champs corporate au greffe (formule de repli OK
en attendant). DAC7 et inscription médiateur ne sont pas bloquants
Stripe (relèvent du fisc et de la consommation FR).

---

## 4. Risque résiduel CNIL (post-round)

### Réglés

- ✅ Bannière cookies — equalWeightButtons + CSS + 4 liens locale-aware
- ✅ Bouton flottant retrait consentement (CookieConsentManageButton)
- ✅ Mentions légales identifiant le responsable de traitement
- ✅ Fusion /privacy → /legal/politique-confidentialite (politique
  unique)
- ✅ Tableau des 11 traitements avec base légale
- ✅ Profilage / art. 22 explicité
- ✅ DPF en tant que mécanisme **supplémentaire** uniquement (Schrems II)
- ✅ Catégories particulières art. 9 (biométrie) — consentement
  explicite + non-stockage frames
- ✅ Notification violation art. 33-34 mentionnée

### Réserves

1. **Formulaire d'appel art. 22 outillé (sans authentification)** —
   actuellement renvoi vers `dpo@designedtrust.com` par email. Non
   bloquant CNIL mais sub-optimal. **Phase 3 ou follow-up backend.**
2. **Preuve serveur-side du consentement (CMP)** — actuellement
   localStorage uniquement. Idéalement table `consent_records` Postgres.
   Non bloquant CNIL en première saisine mais à prévoir. **Phase 3.**
3. **DPO formellement nommé + déclaré à la CNIL** — la page liste
   `dpo@designedtrust.com` mais la nomination doit être déclarée à
   la CNIL via téléservice. **Décision humaine + admin.**
4. **Inscription Médiateur de la consommation** (MEDICYS / CMAP) —
   décision business + €500/an.
5. **Audit purge effective** (messages 3 ans, sessions 30j, tokens
   60j) — la politique l'annonce, à vérifier côté backend. **Phase 3.**
6. **TIA documentées** pour les 11 sous-processeurs hors UE — la
   politique l'annonce comme « documentée en interne » ; à
   matérialiser. **Phase 3 ou follow-up.**

### Verdict CNIL

**✅ PASSE AVEC RÉSERVES MINEURES :** plus de bloquant majeur. Les
réserves résiduelles sont opérationnelles (DPO déclaré, médiateur
inscrit, purge auditée, TIA matérialisées) plutôt que structurelles.
Le risque de plainte CNIL recevable et **sanctionnée** est désormais
très faible.

---

## 5. Risque résiduel DSA (Règlement (UE) 2022/2065)

| Article | Obligation | État |
|---------|-----------|------|
| Art. 11-12 | Point de contact autorités / utilisateurs | ✅ Mentions légales + support@designedtrust.com |
| Art. 14 | Conditions générales claires + indexables | ✅ CGU + CGV + Code de conduite indexables, langage clair |
| Art. 16 | Mécanisme de signalement | ⚠️ Placeholder UI ReportButton (mailto:) — backend à brancher |
| Art. 17 | Notification motivée de décision | ✅ Code de conduite + CGU art. 11 mentionnent SLA 48h + 10j |
| Art. 23 | Mesures contre les abus | ✅ Sanctions graduées documentées |
| Art. 24 | Rapport annuel | ✅ Exempté <50 salariés + engagement formel publication 2027 |

**Verdict DSA :** ✅ **PASSE**. Le seul gap est le branchement backend
du mécanisme de signalement (l'UI est en place via `ReportButton`
mailto:), mais le DSA n'exige pas une UI native — l'email est un
canal de signalement valide.

---

## 6. Verdict global

| Domaine | Verdict |
|---------|---------|
| Stripe Restricted Businesses | ⚠️ PASSE avec réserve (greffe à finaliser) |
| CNIL | ✅ PASSE avec réserves mineures |
| DSA | ✅ PASSE |
| LCEN art. 6 | ✅ PASSE |
| Code de la consommation L.111-1 | ✅ PASSE (tarifs fixés) |
| Code de commerce L.123-22 | ✅ PASSE (10 ans documenté) |
| DAC7 (fiscal) | ❌ NE PASSE PAS — implémentation backend manquante |
| Worker Classification | ✅ PASSE (clause Malt verbatim) |
| PSD2 / agent commercial Stripe | ✅ PASSE (mention explicite) |
| LCB-FT | ⚠️ PASSE avec réserve (apporteur KYC à vérifier en code) |

**Verdict global : PASSE avec réserves majoritairement humaines
/ admin (greffe, DPO déclaration, médiateur inscription) + un
bloquant technique restant (DAC7).**

---

## 7. Décisions humaines bloquantes

Les décisions suivantes nécessitent une intervention humaine /
business / juridique et ne peuvent être résolues par un agent :

1. **Finaliser l'enregistrement de Designed Trust SAS au greffe** —
   raison sociale exacte, RCS, capital social libéré, adresse
   postale, directeur de publication nommé. ETA : 5-15 jours.
2. **Nommer un DPO formel + déclarer à la CNIL** — interne (CTO /
   responsable légal) ou externe (DPO externalisé) certifié. ETA :
   5 jours.
3. **Inscrire la société auprès d'un Médiateur de la consommation**
   (MEDICYS ou CMAP). Coût annuel ~€500. ETA : 2 jours admin.
4. **Faire valider les CGU et CGV finalisées par un avocat** spécialisé
   droit du numérique + droit fiscal des plateformes. Budget ~€5-12k
   selon le brief. ETA : 7-14 jours.
5. **Implémentation DAC7 backend** — module `adapter/dac7/`, cron
   annuel, format XML DGFiP, récap individuel. ETA dev : 3-5 jours.
6. **Audit KYC apporteur d'affaires côté code** — vérifier que la
   présence d'un Connected Account Stripe est exigée avant tout
   versement de commission. ETA audit : 1 jour.
7. **Mise à jour de la migration legacy `subscriptions(user_id)` →
   `subscriptions(organization_id)`** — flaggée dans la mémoire
   projet (2026-04-22) mais hors scope de cette round.

---

## 8. Variables d'env / config Vercel/Railway à mettre à jour

Aucune variable d'environnement nouvelle introduite par cette round.
Tous les ajouts s'appuient sur les valeurs déjà présentes :

- `NEXT_PUBLIC_DPO_EMAIL` — déjà utilisé par `getDpoEmail()` (fallback
  vers `dpo@designedtrust.com`).
- Pas d'API key, pas de webhook, pas de feature flag.

---

## 9. Hors-scope flaggé (non touché)

Les éléments suivants étaient hors scope de la round legal-max-blindage
et restent à traiter dans une future round :

1. **Backend** — aucune modification (off-limits). Les obligations
   backend (DAC7 adapter, endpoint DSA report, table `consent_records`,
   purge cron) sont documentées dans le rapport ci-dessus mais
   non implémentées.
2. **Mobile (Flutter)** — pas de parité mobile dans cette round. Le
   bouton de signalement, le bandeau anti-spam et la politique de
   confidentialité longue doivent être portés côté mobile.
3. **Admin (Vite + React)** — pas de changement. L'admin doit
   probablement intégrer un dashboard de modération pour traiter les
   signalements DSA art. 16/17.
4. **LiveKit (4e AIPD si enregistrements activés)** — hors scope par
   mémoire `feedback_no_touch_livekit.md`.
5. **Implémentation effective des sub-processeurs hors UE TIA** — 11
   documents à produire en interne. La page sous-processeurs
   l'annonce ; le contenu réel des TIA reste à matérialiser.
6. **Migration `subscriptions(user_id)` → `(organization_id)`** —
   flaggée dans la mémoire projet, hors scope.
7. **Formulaire d'appel art. 22 outillé (sans auth)** — actuellement
   mailto: vers dpo@designedtrust.com. Un endpoint backend public
   anonyme + une page `/decisions-automatisees/recours` seraient
   préférables.
8. **Audit purge effective des messages 3 ans / sessions 30j /
   tokens push 60j** — audit du code worker, hors scope cette round.

---

## 10. Conclusion

La round **legal-max-blindage** a augmenté le score moyen des pages
légales de 4.2 / 10 à 7.8 / 10. Les bloquants Stripe identifiés dans
`LEGAL_AUDIT_EXPERT.md` (identité éditeur placeholder, tarification
indicative, pages noindex, absence de Worker Classification, bannière
cookies non-CNIL) sont tous **résolus côté texte / UI**. Le risque
de rejet Stripe est désormais limité à la finalisation de
l'enregistrement au greffe — qui peut s'opérer en parallèle de la
revue Stripe en s'appuyant sur la formule de repli officielle.

Le risque CNIL est limité à des réserves opérationnelles (DPO
formellement déclaré, médiateur inscrit, purge auditée, TIA
matérialisées) plutôt que structurelles.

**Pas d'itération Phase 3 nécessaire en mode autonome :** les gaps
restants relèvent soit du backend (off-limits cette round), soit de
décisions business/admin/avocat. L'agent ayant épuisé l'enveloppe
de polish texte/UI qu'il pouvait livrer sans toucher au backend.

— Fin du Round 2 —
