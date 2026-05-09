import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/theme/app_theme.dart';
import '../../../../l10n/app_localizations.dart';
import '../../domain/entities/job_application_entity.dart';
import '../providers/job_provider.dart';
import '../widgets/candidate_card.dart';

/// Standalone candidates screen — Soleil v2 visual port.
///
/// Reachable from the legacy "/candidates" route (now superseded by
/// the M-08 candidates tab inside `JobDetailScreen`). Keeps a Soleil
/// AppBar + the same Soleil-styled empty/loading/error states used in
/// the tabbed flow.
///
/// 2026-05-09 — Persona filter (Fix 3): segmented chips above the list
/// narrow the rows to a single applicant_kind. Each filter is a
/// separate cache entry via [filteredJobApplicationsProvider].
class CandidatesScreen extends ConsumerStatefulWidget {
  const CandidatesScreen({super.key, required this.jobId});

  final String jobId;

  @override
  ConsumerState<CandidatesScreen> createState() => _CandidatesScreenState();
}

class _CandidatesScreenState extends ConsumerState<CandidatesScreen> {
  ApplicantKind? _kindFilter;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;
    final l10n = AppLocalizations.of(context)!;
    final args = JobApplicationsArgs(jobId: widget.jobId, kindFilter: _kindFilter);
    final candidates = ref.watch(filteredJobApplicationsProvider(args));

    return Scaffold(
      backgroundColor: cs.surface,
      appBar: AppBar(
        backgroundColor: cs.surfaceContainerLowest,
        scrolledUnderElevation: 0,
        elevation: 0,
        title: Text(
          l10n.applications,
          style: SoleilTextStyles.titleLarge.copyWith(
            color: cs.onSurface,
            fontSize: 18,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      body: SafeArea(
        top: false,
        child: Column(
          children: [
            _CandidateFilterBar(
              active: _kindFilter,
              onChange: (next) => setState(() => _kindFilter = next),
            ),
            Expanded(
              child: RefreshIndicator(
                onRefresh: () async => ref.invalidate(filteredJobApplicationsProvider(args)),
                child: candidates.when(
                  loading: () => const Center(child: CircularProgressIndicator()),
                  error: (e, _) => _ErrorView(
                    args: args,
                    message: l10n.somethingWentWrong,
                    retryLabel: l10n.retry,
                  ),
                  data: (items) {
                    if (items.isEmpty) {
                      return _EmptyView(
                        title: l10n.jobDetail_m08_emptyTitle,
                        body: l10n.jobDetail_m08_emptyBody,
                      );
                    }
                    return ListView.separated(
                      padding: const EdgeInsets.fromLTRB(20, 16, 20, 28),
                      itemCount: items.length,
                      separatorBuilder: (_, __) => const SizedBox(height: 12),
                      itemBuilder: (context, index) => CandidateCard(
                        item: items[index],
                        jobId: widget.jobId,
                        candidates: items,
                        candidateIndex: index,
                      ),
                    );
                  },
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _CandidateFilterBar extends StatelessWidget {
  const _CandidateFilterBar({
    required this.active,
    required this.onChange,
  });

  final ApplicantKind? active;
  final ValueChanged<ApplicantKind?> onChange;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final entries = <_FilterEntry>[
      _FilterEntry(value: null, label: l10n.candidateFilterAll),
      _FilterEntry(
        value: ApplicantKind.freelance,
        label: l10n.candidateFilterFreelances,
      ),
      _FilterEntry(
        value: ApplicantKind.agency,
        label: l10n.candidateFilterAgencies,
      ),
      _FilterEntry(
        value: ApplicantKind.referrer,
        label: l10n.candidateFilterReferrers,
      ),
    ];
    return SizedBox(
      height: 48,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
        itemCount: entries.length,
        separatorBuilder: (_, __) => const SizedBox(width: 8),
        itemBuilder: (context, i) {
          final entry = entries[i];
          final isActive = entry.value == active;
          return _FilterChip(
            label: entry.label,
            isActive: isActive,
            onTap: () => onChange(entry.value),
          );
        },
      ),
    );
  }
}

class _FilterEntry {
  const _FilterEntry({required this.value, required this.label});
  final ApplicantKind? value;
  final String label;
}

class _FilterChip extends StatelessWidget {
  const _FilterChip({
    required this.label,
    required this.isActive,
    required this.onTap,
  });

  final String label;
  final bool isActive;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final soleil = Theme.of(context).extension<AppColors>()!;
    return Semantics(
      button: true,
      label: label,
      selected: isActive,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(AppTheme.radiusFull),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
          decoration: BoxDecoration(
            color: isActive ? soleil.accentSoft : cs.surfaceContainerLowest,
            border: Border.all(
              color: isActive ? cs.primary : cs.outline,
              width: isActive ? 1.5 : 1,
            ),
            borderRadius: BorderRadius.circular(AppTheme.radiusFull),
          ),
          child: Center(
            child: Text(
              label,
              style: SoleilTextStyles.bodyEmphasis.copyWith(
                fontSize: 12.5,
                color: isActive ? soleil.primaryDeep : cs.onSurfaceVariant,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _ErrorView extends ConsumerWidget {
  const _ErrorView({
    required this.args,
    required this.message,
    required this.retryLabel,
  });

  final JobApplicationsArgs args;
  final String message;
  final String retryLabel;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.fromLTRB(20, 60, 20, 28),
      children: [
        Icon(
          Icons.error_outline,
          size: 48,
          color: theme.colorScheme.onSurfaceVariant,
        ),
        const SizedBox(height: 12),
        Text(
          message,
          textAlign: TextAlign.center,
          style: theme.textTheme.bodyMedium?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
        const SizedBox(height: 8),
        Center(
          child: TextButton(
            onPressed: () => ref.invalidate(filteredJobApplicationsProvider(args)),
            child: Text(retryLabel),
          ),
        ),
      ],
    );
  }
}

class _EmptyView extends StatelessWidget {
  const _EmptyView({required this.title, required this.body});

  final String title;
  final String body;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;
    final soleil = theme.extension<AppColors>()!;

    return ListView(
      padding: const EdgeInsets.fromLTRB(20, 40, 20, 28),
      children: [
        Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: soleil.accentSoft,
            borderRadius: BorderRadius.circular(AppTheme.radius2xl),
            border: Border.all(color: cs.outline),
          ),
          child: Column(
            children: [
              Container(
                width: 56,
                height: 56,
                decoration: BoxDecoration(
                  color: cs.surfaceContainerLowest,
                  borderRadius: BorderRadius.circular(AppTheme.radiusFull),
                ),
                child: Icon(
                  Icons.groups_outlined,
                  color: soleil.primaryDeep,
                  size: 28,
                ),
              ),
              const SizedBox(height: 16),
              Text(
                title,
                textAlign: TextAlign.center,
                style: SoleilTextStyles.titleMedium.copyWith(
                  color: cs.onSurface,
                  fontSize: 18,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                body,
                textAlign: TextAlign.center,
                style: SoleilTextStyles.body.copyWith(
                  color: cs.onSurfaceVariant,
                  height: 1.5,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
