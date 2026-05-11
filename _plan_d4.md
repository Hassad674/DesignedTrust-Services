# Plan D4 — GDPR Phase C : compliance documentation

Branche : `feat/gdpr-phase-c-docs`
Auteur : Hassad674
Cible : production-grade — RGPD prêt pour DPO + déploiement Vercel/Railway.

## Sortie attendue

1. Six (6) documents markdown canoniques sous `/legal/` (racine du repo) :
   - `legal/registre.md` (Registre des traitements — RGPD art. 30)
   - `legal/aipd.md` (Analyse d'Impact à la Protection des Données — RGPD art. 35)
   - `legal/dpa-template.md` (modèle Accord de Traitement — RGPD art. 28)
   - `legal/politique-confidentialite.md` (Politique de confidentialité publique — RGPD art. 13/14)
   - `legal/cgu.md` (Conditions Générales d'Utilisation)
   - `legal/cgv.md` (Conditions Générales de Vente B2B)
2. Sept (7) pages publiques web sous `web/src/app/[locale]/(public)/legal/` :
   - `page.tsx` (index — sommaire)
   - `registre/page.tsx`
   - `aipd/page.tsx`
   - `dpa-template/page.tsx`
   - `politique-confidentialite/page.tsx`
   - `cgu/page.tsx`
   - `cgv/page.tsx`
3. Sitemap (`web/src/app/sitemap.ts`) mis à jour pour inclure les 7 routes (`priority: 0.3, changeFrequency: yearly`).
4. Footer (`web/src/shared/components/legal/legal-footer.tsx`) augmenté d'une section "Documents légaux complets".
5. Traductions FR + EN dans `web/messages/{fr,en}.json` (titres, intros, nav). Le contenu détaillé est en FR avec une mention "English version available on request" en tête (anglais conservé pour les titres + politique de confidentialité).
6. Tests :
   - `web/e2e/legal-routes.spec.ts` (E2E Playwright : status 200 sur chaque page, H1 présent, lien footer, sitemap contient les URLs).
   - Tests unitaires Vitest : composant `LegalDocument` (renderer des sections markdown locales), `LegalFooter` (vérifier nouveaux liens), index `/legal`.

## Stratégie de rendu Markdown

Aucune dépendance markdown n'est installée ni ne sera ajoutée (la règle "pas de nouvelle dep sans nécessité"). Les documents `legal/*.md` sont la source canonique destinée au DPO (export Word direct). Les pages web rendent un équivalent **HTML sémantique direct** via un composant React serveur (`LegalDocument`) qui reçoit un tableau structuré de sections (heading + paragraphes + tableaux). Le contenu utilisateur reste donc strictement statique — aucun parsing dangereux, aucune dépendance ajoutée. Les fichiers markdown et les sections React partagent le même fond rédactionnel.

## Documents — taille cible

| Doc | Words target | Statut |
|-----|--------------|--------|
| registre.md | 2 500-3 500 | drafted |
| aipd.md | 2 000-3 000 | drafted |
| dpa-template.md | 1 800-2 500 | drafted |
| politique-confidentialite.md | 2 000-3 000 | drafted |
| cgu.md | 2 500-3 500 | drafted |
| cgv.md | 2 000-3 000 | drafted |

## Hors scope

- Aucun changement de schéma DB / endpoints API.
- Aucune signature à la place de l'utilisateur — tout est marqué `[À COMPLÉTER : ...]`.
- Pas de polyfill / dépendance markdown ajoutée.
- Pas de version EN complète des documents (la version EN ne couvre que les titres / nav / intro). Mention "English version available on request" affichée en haut.

## Pipeline de validation (mandatoire avant commit final)

```bash
cd web && npx tsc --noEmit
cd web && npx vitest run
cd web && npx playwright test legal-routes.spec.ts
cd web && npm run build 2>&1 | tail -20
```

## Commits prévus

1. `docs(gdpr-phase-c): plan` (ce fichier)
2. `docs(gdpr-phase-c): registre traitements + AIPD + DPA template` (markdown sources)
3. `docs(gdpr-phase-c): politique confidentialité + CGU + CGV` (markdown sources)
4. `feat(legal): routes /legal/* + composant LegalDocument` (pages web)
5. `feat(legal): footer + sitemap + i18n` (wiring)
6. `test(legal): vitest + playwright`

Tous portent l'identité `Hassad674` (51391731+Hassad674@users.noreply.github.com) avec trailer `Co-Authored-By: Claude`.
