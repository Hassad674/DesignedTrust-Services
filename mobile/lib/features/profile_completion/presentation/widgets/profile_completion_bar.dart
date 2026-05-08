import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/theme/theme_colors.dart';
import '../../../../l10n/app_localizations.dart';
import '../../domain/entities/profile_completion_report.dart';
import '../providers/profile_completion_providers.dart';

/// ProfileCompletionBar — Soleil v2 progress card surfacing the
/// "Profil rempli à X%" report on every profile-related screen.
///
/// Composes three states:
///
///   * loading / empty data -> renders nothing (no skeleton flash on
///     screens where the bar is a secondary affordance).
///   * complete -> hidden when [hideWhenComplete] is true to avoid a
///     dead UI block on a fully completed profile.
///   * partial -> renders the corail-filled bar plus a chevron pill
///     showing the missing-section count. Tapping opens a modal with
///     a list of every section, marked filled / empty.
class ProfileCompletionBar extends ConsumerWidget {
  const ProfileCompletionBar({super.key, this.hideWhenComplete = false});

  /// When true, the widget collapses to `SizedBox.shrink()` once the
  /// report reaches 100%. Defaults to false: surfaces that want to
  /// celebrate completion (e.g. the profile page) keep the bar visible.
  final bool hideWhenComplete;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(profileCompletionProvider);
    return async.when(
      data: (report) =>
          _buildContent(context, ref, report, hideWhenComplete: hideWhenComplete),
      loading: () => const SizedBox.shrink(),
      error: (_, __) => const SizedBox.shrink(),
    );
  }

  Widget _buildContent(
    BuildContext context,
    WidgetRef ref,
    ProfileCompletionReport report, {
    required bool hideWhenComplete,
  }) {
    if (hideWhenComplete && report.isComplete) {
      return const SizedBox.shrink();
    }
    if (report.totalSections == 0) {
      return const SizedBox.shrink();
    }

    final theme = Theme.of(context);
    final colors = theme.extension<AppColors>()!;
    final l10n = AppLocalizations.of(context)!;

    return InkWell(
      onTap: () => _openMissingModal(context, report),
      borderRadius: BorderRadius.circular(16),
      child: Ink(
        decoration: BoxDecoration(
          color: theme.colorScheme.surface,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: colors.border),
        ),
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        l10n.profileCompletionTitle(report.percent),
                        style: theme.textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        report.isComplete
                            ? l10n.profileCompletionSubtitleComplete
                            : l10n.profileCompletionSubtitle(
                                report.filledSections,
                                report.totalSections,
                              ),
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: colors.mutedForeground,
                        ),
                      ),
                    ],
                  ),
                ),
                if (!report.isComplete)
                  _MissingPill(count: report.missingCount),
              ],
            ),
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: LinearProgressIndicator(
                value: (report.percent / 100).clamp(0, 1).toDouble(),
                minHeight: 8,
                backgroundColor: colors.muted,
                valueColor: AlwaysStoppedAnimation<Color>(
                  report.isComplete
                      ? colors.success
                      : theme.colorScheme.primary,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _openMissingModal(
    BuildContext context,
    ProfileCompletionReport report,
  ) {
    return showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => _ProfileCompletionMissingSheet(report: report),
    );
  }
}

class _MissingPill extends StatelessWidget {
  const _MissingPill({required this.count});
  final int count;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.extension<AppColors>()!;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: colors.accentSoft,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            '$count',
            style: theme.textTheme.labelSmall?.copyWith(
              color: colors.primaryDeep,
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(width: 4),
          Icon(Icons.chevron_right, size: 14, color: colors.primaryDeep),
        ],
      ),
    );
  }
}

class _ProfileCompletionMissingSheet extends StatelessWidget {
  const _ProfileCompletionMissingSheet({required this.report});
  final ProfileCompletionReport report;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.extension<AppColors>()!;
    final l10n = AppLocalizations.of(context)!;

    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    l10n.profileCompletionModalTitle(report.percent),
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  tooltip: l10n.profileCompletionModalCloseLabel,
                  onPressed: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              l10n.profileCompletionModalSubtitle,
              style: theme.textTheme.bodyMedium?.copyWith(
                color: colors.mutedForeground,
              ),
            ),
            const SizedBox(height: 12),
            ConstrainedBox(
              constraints: BoxConstraints(
                maxHeight: MediaQuery.of(context).size.height * 0.6,
              ),
              child: ListView.separated(
                shrinkWrap: true,
                itemCount: report.sections.length,
                separatorBuilder: (_, __) => const SizedBox(height: 8),
                itemBuilder: (context, index) {
                  final section = report.sections[index];
                  return _SectionTile(section: section);
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _SectionTile extends StatelessWidget {
  const _SectionTile({required this.section});
  final ProfileCompletionSection section;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.extension<AppColors>()!;
    final label = _labelFor(section.key, AppLocalizations.of(context)!);

    if (section.filled) {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          color: colors.successSoft,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Icon(Icons.check, size: 18, color: colors.success),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                label,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: colors.mutedForeground,
                  decoration: TextDecoration.lineThrough,
                ),
              ),
            ),
          ],
        ),
      );
    }

    return Container(
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        border: Border.all(color: colors.border),
        borderRadius: BorderRadius.circular(12),
      ),
      child: ListTile(
        title: Text(label, style: theme.textTheme.bodyMedium),
        trailing: const Icon(Icons.chevron_right, size: 18),
        // The router-aware navigation lives in the host screen; here
        // we just close the sheet so the user lands back on the profile
        // page where they can tap the section to edit. Surfacing each
        // path as a deep-link would require a router map equal to the
        // web's; the current mobile build keeps the navigation surface
        // simple by leaving the editor entry points on the profile
        // screen itself.
        onTap: () => Navigator.of(context).pop(),
      ),
    );
  }

  String _labelFor(String key, AppLocalizations l10n) {
    switch (key) {
      case 'photo':
        return l10n.profileCompletionSectionPhoto;
      case 'title':
        return l10n.profileCompletionSectionTitle;
      case 'about':
        return l10n.profileCompletionSectionAbout;
      case 'expertises':
        return l10n.profileCompletionSectionExpertises;
      case 'skills':
        return l10n.profileCompletionSectionSkills;
      case 'pricing':
        return l10n.profileCompletionSectionPricing;
      case 'availability':
        return l10n.profileCompletionSectionAvailability;
      case 'location':
        return l10n.profileCompletionSectionLocation;
      case 'languages':
        return l10n.profileCompletionSectionLanguages;
      case 'video':
        return l10n.profileCompletionSectionVideo;
      case 'social_links':
        return l10n.profileCompletionSectionSocialLinks;
      case 'billing_profile':
        return l10n.profileCompletionSectionBillingProfile;
      case 'kyc':
        return l10n.profileCompletionSectionKyc;
      case 'portfolio':
        return l10n.profileCompletionSectionPortfolio;
      case 'client_about':
        return l10n.profileCompletionSectionClientAbout;
    }
    // Unknown keys fall back to the raw machine identifier — never a
    // crash, and a follow-up release adds the missing translation.
    return key;
  }
}
