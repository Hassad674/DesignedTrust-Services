"use client"

import { useTranslations } from "next-intl"
import type {
  FieldErrors,
  UseFormRegister,
} from "react-hook-form"
import { Input } from "@/shared/components/ui/input"
import { cn } from "@/shared/lib/utils"

// The subset of the report form values these fields own. Kept local so
// the component does not import the parent's schema type (which would
// create a circular dependency between form ↔ fields).
type ReportTextValues = {
  title: string
  description: string
  reporterEmail: string
  hp: string
}

interface ReportTextFieldsProps {
  register: UseFormRegister<ReportTextValues>
  errors: FieldErrors<ReportTextValues>
  /** Anonymous reporters get an optional contact-email field. */
  showEmailField: boolean
}

const TEXTAREA_CLASS = cn(
  "block w-full rounded-lg border border-border bg-card px-3 py-2 text-sm shadow-xs",
  "transition-all duration-200 ease-out placeholder:text-muted-foreground",
  "focus:border-primary focus:outline-none focus:ring-4 focus:ring-primary/15",
)

// ReportTextFields — title (required), description (required), and an
// optional contact email shown only to anonymous reporters. The
// honeypot input is rendered here, visually hidden and aria-hidden, so a
// human never sees it but a form-filling bot trips it.
export function ReportTextFields({
  register,
  errors,
  showEmailField,
}: ReportTextFieldsProps) {
  const t = useTranslations("feedback")
  return (
    <>
      <Input
        label={t("title_label")}
        placeholder={t("title_placeholder")}
        error={errors.title ? t("title_error") : undefined}
        autoComplete="off"
        {...register("title")}
      />

      <div className="flex flex-col gap-1">
        <label
          htmlFor="feedback-description"
          className="text-sm font-medium text-foreground"
        >
          {t("description_label")}
        </label>
        <textarea
          id="feedback-description"
          rows={4}
          placeholder={t("description_placeholder")}
          aria-invalid={errors.description ? true : undefined}
          aria-describedby={
            errors.description ? "feedback-description-error" : undefined
          }
          className={TEXTAREA_CLASS}
          {...register("description")}
        />
        {errors.description && (
          <p
            id="feedback-description-error"
            role="alert"
            className="text-xs text-primary-deep"
          >
            {t("description_error")}
          </p>
        )}
      </div>

      {showEmailField && (
        <Input
          type="email"
          label={t("email_label")}
          hint={t("email_hint")}
          placeholder={t("email_placeholder")}
          error={errors.reporterEmail ? t("email_error") : undefined}
          autoComplete="email"
          {...register("reporterEmail")}
        />
      )}

      {/* Honeypot — hidden from humans + assistive tech, never focusable. */}
      <input
        type="text"
        tabIndex={-1}
        autoComplete="off"
        aria-hidden="true"
        className="sr-only"
        {...register("hp")}
      />
    </>
  )
}
