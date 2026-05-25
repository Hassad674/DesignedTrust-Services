import 'package:flutter/material.dart';

import '../../../../core/theme/app_theme.dart';
import '../../../../l10n/app_localizations.dart';
import '../../domain/entities/feedback_type.dart';

/// Two-option segmented toggle for the report type (Bug / Security).
/// Soleil v2: active option gets the corail-soft background + corail
/// border, mirroring the apply-sheet persona picker.
class FeedbackTypeToggle extends StatelessWidget {
  const FeedbackTypeToggle({
    super.key,
    required this.value,
    required this.onChanged,
  });

  final FeedbackType value;
  final ValueChanged<FeedbackType> onChanged;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Row(
      children: [
        Expanded(
          child: _Option(
            label: l10n.feedbackTypeBug,
            icon: Icons.bug_report_outlined,
            isActive: value == FeedbackType.bug,
            onTap: () => onChanged(FeedbackType.bug),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _Option(
            label: l10n.feedbackTypeSecurity,
            icon: Icons.shield_outlined,
            isActive: value == FeedbackType.security,
            onTap: () => onChanged(FeedbackType.security),
          ),
        ),
      ],
    );
  }
}

class _Option extends StatelessWidget {
  const _Option({
    required this.label,
    required this.icon,
    required this.isActive,
    required this.onTap,
  });

  final String label;
  final IconData icon;
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
        borderRadius: BorderRadius.circular(AppTheme.radiusMd),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
          decoration: BoxDecoration(
            color: isActive ? soleil.accentSoft : cs.surfaceContainerLowest,
            border: Border.all(
              color: isActive ? cs.primary : soleil.borderStrong,
              width: isActive ? 1.5 : 1,
            ),
            borderRadius: BorderRadius.circular(AppTheme.radiusMd),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                icon,
                size: 18,
                color: isActive ? cs.primary : cs.outline,
              ),
              const SizedBox(width: 8),
              Text(
                label,
                style: SoleilTextStyles.bodyEmphasis.copyWith(
                  fontSize: 13,
                  color: isActive ? cs.primary : cs.onSurface,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
