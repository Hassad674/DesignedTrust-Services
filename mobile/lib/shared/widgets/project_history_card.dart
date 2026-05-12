import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../core/models/review.dart';
import 'review_card_widget.dart';

import '../../core/theme/app_theme.dart';
/// Shared card rendering one completed project — amount pill + date
/// header, optional title, embedded review (or "Awaiting review"
/// placeholder), and an optional [footer] slot consumers can use to
/// append context such as a counterparty chip.
///
/// Used both by the provider-side project history (under
/// `features/project_history/`) and by the client-profile project
/// history (under `features/client_profile/`) so the visual pattern
/// stays in sync across surfaces.
class ProjectHistoryCard extends StatelessWidget {
  const ProjectHistoryCard({
    super.key,
    required this.title,
    required this.amountCents,
    required this.completedAt,
    this.review,
    this.footer,
    this.onLeaveReview,
  });

  /// Proposal title. Renders only when non-empty (callers may hide
  /// the title when the other party opted out of sharing it).
  final String title;

  /// Amount in cents. Formatted as a grouped euro value for the pill.
  final int amountCents;

  /// Timestamp the proposal was marked completed.
  final DateTime completedAt;

  /// Optional review to embed inside the card. When null the card
  /// renders an "Awaiting review" placeholder.
  final Review? review;

  /// Optional widget appended after the review body. Used on the
  /// client surface to attach the provider chip.
  final Widget? footer;

  /// When non-null AND the entry has no review yet, the awaiting-
  /// review placeholder becomes a tappable button that fires this
  /// callback. Parents pass it on profiles where the viewer is the
  /// counterparty and is eligible to leave their review. Anonymous
  /// visitors (or already-reviewed entries) do not pass a callback.
  final VoidCallback? onLeaveReview;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final amountEuros = amountCents / 100;
    final formattedAmount = NumberFormat.currency(
      locale: 'fr_FR',
      symbol: '€',
      decimalDigits: 0,
    ).format(amountEuros);
    final formattedDate =
        DateFormat('d MMM yyyy', 'fr_FR').format(completedAt);

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: theme.dividerColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header: amount pill + date
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 5,
                ),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      Theme.of(context).colorScheme.primaryContainer,
                      Theme.of(context).colorScheme.errorContainer,
                    ],
                  ),
                  borderRadius: BorderRadius.circular(99),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      Icons.euro,
                      size: 13,
                      color: (Theme.of(context).extension<AppColors>()?.primaryDeep ?? Theme.of(context).colorScheme.error),
                    ),
                    const SizedBox(width: 3),
                    Text(
                      formattedAmount.replaceAll('€', '').trim(),
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                        color: (Theme.of(context).extension<AppColors>()?.primaryDeep ?? Theme.of(context).colorScheme.error),
                      ),
                    ),
                  ],
                ),
              ),
              Row(
                children: [
                  Icon(
                    Icons.schedule,
                    size: 13,
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    formattedDate,
                    style: theme.textTheme.labelSmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ],
          ),
          if (title.isNotEmpty) ...[
            const SizedBox(height: 10),
            Text(
              title,
              style: theme.textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
          const SizedBox(height: 12),

          // Body: review or awaiting state (clickable when eligible)
          if (review != null)
            ReviewCardWidget(review: review!)
          else
            AwaitingReviewBox(onTap: onLeaveReview),

          if (footer != null) ...[
            const SizedBox(height: 10),
            footer!,
          ],
        ],
      ),
    );
  }
}

/// Placeholder rendered inside [ProjectHistoryCard] when no review has
/// been submitted yet. Exposed publicly so feature-level widgets can
/// reuse the same visual outside of a card body if ever needed.
///
/// When [onTap] is non-null, the placeholder becomes a tappable
/// "Laisser ton avis" button — that's the entry point of the
/// double-blind reviews flow from a partner's profile project history.
class AwaitingReviewBox extends StatelessWidget {
  const AwaitingReviewBox({super.key, this.onTap});

  /// Tap handler. When null, the placeholder stays static (read-only).
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isInteractive = onTap != null;
    final iconColor = isInteractive
        ? theme.colorScheme.primary
        : theme.colorScheme.onSurfaceVariant;
    final textColor = isInteractive
        ? theme.colorScheme.primary
        : theme.colorScheme.onSurfaceVariant;
    final label = isInteractive
        ? 'Laisser ton avis sur ce projet'
        : 'Awaiting review';

    final content = Row(
      children: [
        Icon(
          isInteractive ? Icons.edit_outlined : Icons.schedule,
          size: 16,
          color: iconColor,
        ),
        const SizedBox(width: 8),
        Text(
          label,
          style: theme.textTheme.bodySmall?.copyWith(
            color: textColor,
            fontWeight: isInteractive ? FontWeight.w600 : FontWeight.w400,
          ),
        ),
      ],
    );

    final container = Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: isInteractive
            ? theme.colorScheme.primaryContainer
            : theme.colorScheme.surfaceContainerLow,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isInteractive
              ? theme.colorScheme.primary.withValues(alpha: 0.4)
              : theme.dividerColor,
          style: BorderStyle.solid,
        ),
      ),
      child: content,
    );

    if (!isInteractive) return container;

    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: container,
      ),
    );
  }
}
