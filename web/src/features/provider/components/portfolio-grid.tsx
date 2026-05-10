"use client"

import { useState } from "react"
import { Plus, Briefcase } from "lucide-react"
import { useTranslations } from "next-intl"
import { useHasPermission } from "@/shared/hooks/use-permissions"
import { useMyPortfolio, usePortfolioByOrganization, useDeletePortfolioItem } from "../hooks/use-portfolio"
import { PortfolioItemCard } from "./portfolio-item-card"
import { PortfolioDetailModal } from "./portfolio-detail-modal"
import { PortfolioFormModal } from "./portfolio-form-modal"
import type { PortfolioItem } from "../api/portfolio-api"
import { Button } from "@/shared/components/ui/button"

const MAX_ITEMS = 30

// --- Edit mode (profile dashboard) ---

// Soleil v2 shell-parity refactor: outer wrapper, header and empty
// state now mirror every other agency-profile section (social-links,
// languages, location, project-history). Behavior — fetch, mutations,
// modals — is untouched. The previous version painted the empty state
// with a corail wash + gradient CTA that drifted from the rest of the
// edit page; the new variant uses the canonical ivoire card +
// muted-tone empty illustration + corail outline CTA.
export function PortfolioSection() {
  const { data, isLoading } = useMyPortfolio()
  const deleteItem = useDeletePortfolioItem()
  const canEdit = useHasPermission("org_profile.edit")
  const t = useTranslations("portfolio")

  const [viewItem, setViewItem] = useState<PortfolioItem | null>(null)
  const [editItem, setEditItem] = useState<PortfolioItem | undefined>(undefined)
  const [showForm, setShowForm] = useState(false)

  const items = data?.data ?? []

  const handleDelete = (id: string) => {
    if (window.confirm(t("confirmDelete"))) {
      deleteItem.mutate(id)
    }
  }

  const openCreate = () => {
    setEditItem(undefined)
    setShowForm(true)
  }

  const openEdit = (item: PortfolioItem) => {
    setEditItem(item)
    setShowForm(true)
  }

  return (
    <section className="bg-card border border-border rounded-2xl p-7 shadow-[var(--shadow-card)]">
      {/* Header — matches social-links / languages / project-history */}
      <div className="flex items-center justify-between mb-4 gap-3">
        <div className="min-w-0">
          <h2 className="font-serif text-xl font-medium tracking-[-0.005em] text-foreground">
            {t("sectionTitle")}
          </h2>
          {items.length > 0 ? (
            <p className="mt-1 truncate text-sm text-muted-foreground">
              {t("publicItemCount", { count: items.length })}
            </p>
          ) : null}
        </div>

        {canEdit && items.length > 0 && items.length < MAX_ITEMS ? (
          <Button
            variant="ghost"
            size="auto"
            onClick={openCreate}
            aria-label={t("addProject")}
            className="inline-flex h-9 shrink-0 items-center gap-1.5 rounded-full border border-primary/30 bg-primary-soft px-3 text-sm font-medium text-primary transition-colors hover:bg-primary/10 focus-visible:outline-2 focus-visible:outline-ring focus-visible:outline-offset-2 sm:px-4"
          >
            <Plus className="h-4 w-4" aria-hidden="true" />
            <span className="hidden sm:inline">{t("addProject")}</span>
          </Button>
        ) : null}
      </div>

      {/* Content */}
      {isLoading ? (
        <PortfolioGridSkeleton />
      ) : items.length === 0 ? (
        canEdit ? <EmptyState onCreate={openCreate} /> : null
      ) : (
        <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-3">
          {items.map((item, index) => (
            <div
              key={item.id}
              className="animate-slide-up"
              style={{ animationDelay: `${Math.min(index * 50, 250)}ms` }}
            >
              <PortfolioItemCard
                item={item}
                readOnly={!canEdit}
                onView={() => setViewItem(item)}
                onEdit={canEdit ? () => openEdit(item) : undefined}
                onDelete={canEdit ? () => handleDelete(item.id) : undefined}
              />
            </div>
          ))}
        </div>
      )}

      <PortfolioDetailModal
        item={viewItem}
        open={!!viewItem}
        onClose={() => setViewItem(null)}
      />

      <PortfolioFormModal
        item={editItem}
        open={showForm}
        onClose={() => setShowForm(false)}
        nextPosition={items.length}
      />
    </section>
  )
}

// --- Empty state ---

// Mirrors the project-history empty state: muted icon chip, neutral
// title + italic muted-foreground description, corail-outline CTA.
// No more full corail wash / decorative blurred circles / white text
// on red — that pattern was unique to this card and broke parity with
// every other section on the agency edit shell.
function EmptyState({ onCreate }: { onCreate: () => void }) {
  const t = useTranslations("portfolio")
  return (
    <div className="flex flex-col items-center justify-center py-10 text-center">
      <div className="w-12 h-12 rounded-full bg-primary-soft flex items-center justify-center mb-3">
        <Briefcase
          className="w-6 h-6 text-primary"
          aria-hidden="true"
        />
      </div>
      <p className="text-base font-medium text-foreground mb-1">
        {t("emptyTitle")}
      </p>
      <p className="max-w-sm text-sm text-muted-foreground italic">
        {t("emptyDescription")}
      </p>
      <Button
        variant="ghost"
        size="auto"
        onClick={onCreate}
        className="mt-4 inline-flex h-10 items-center gap-1.5 rounded-full border border-primary/30 bg-primary-soft px-4 text-sm font-medium text-primary transition-colors hover:bg-primary/10 focus-visible:outline-2 focus-visible:outline-ring focus-visible:outline-offset-2"
      >
        <Plus className="h-4 w-4" aria-hidden="true" />
        {t("addFirstProject")}
      </Button>
    </div>
  )
}

// --- Skeleton loading ---

function PortfolioGridSkeleton() {
  return (
    <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-3">
      {[0, 1, 2, 3].map((i) => (
        <div
          key={i}
          className="aspect-[4/5] animate-shimmer rounded-2xl bg-gradient-to-br from-muted via-muted/60 to-muted"
        />
      ))}
    </div>
  )
}

// --- Read-only mode (public profile) ---

interface PublicPortfolioSectionProps {
  orgId: string
}

export function PublicPortfolioSection({ orgId }: PublicPortfolioSectionProps) {
  const { data, isLoading } = usePortfolioByOrganization(orgId)
  const [viewItem, setViewItem] = useState<PortfolioItem | null>(null)
  const t = useTranslations("portfolio")

  const items = data?.data ?? []

  if (!isLoading && items.length === 0) return null

  return (
    <section className="rounded-2xl border border-border bg-card p-4 shadow-sm sm:p-6">
      <div className="mb-4 flex items-center gap-3 sm:mb-5">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-primary-soft to-primary-soft/60 sm:h-10 sm:w-10">
          <Briefcase className="h-5 w-5 text-primary-deep" />
        </div>
        <div className="min-w-0">
          <h2 className="truncate text-base font-semibold tracking-tight text-foreground sm:text-lg">
            {t("sectionTitle")}
          </h2>
          {items.length > 0 && (
            <p className="mt-0.5 truncate text-xs text-muted-foreground">
              {t("publicItemCount", { count: items.length })}
            </p>
          )}
        </div>
      </div>

      {isLoading ? (
        <PortfolioGridSkeleton />
      ) : (
        <div className="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-3">
          {items.map((item, index) => (
            <div
              key={item.id}
              className="animate-slide-up"
              style={{ animationDelay: `${Math.min(index * 50, 250)}ms` }}
            >
              <PortfolioItemCard
                item={item}
                readOnly
                onView={() => setViewItem(item)}
              />
            </div>
          ))}
        </div>
      )}

      <PortfolioDetailModal
        item={viewItem}
        open={!!viewItem}
        onClose={() => setViewItem(null)}
      />
    </section>
  )
}
