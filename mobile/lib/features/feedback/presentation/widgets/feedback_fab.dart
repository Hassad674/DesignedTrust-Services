import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/router/app_router.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../l10n/app_localizations.dart';
import 'feedback_sheet.dart';

/// Small always-visible "Signaler" pill that opens the feedback sheet.
///
/// Mounted globally (above the GoRouter navigator) by [FeedbackOverlay]
/// so it follows the user across screens. Anchored bottom-right, lifted
/// above the bottom-nav so it never collides with it. Compact (icon +
/// label) and styled with Soleil v2 corail.
///
/// The current route is read from the [appRouterProvider] GoRouter
/// (the FAB sits above the navigator, so it cannot read the location
/// from its own context) and passed to the sheet as `page_url`.
class FeedbackFab extends ConsumerWidget {
  const FeedbackFab({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    if (l10n == null) return const SizedBox.shrink();
    final cs = Theme.of(context).colorScheme;

    return Semantics(
      button: true,
      label: l10n.feedbackButtonTooltip,
      child: Material(
        color: cs.primary,
        elevation: 3,
        shadowColor: Colors.black.withValues(alpha: 0.2),
        borderRadius: BorderRadius.circular(AppTheme.radiusFull),
        child: InkWell(
          onTap: () => _open(context, ref),
          borderRadius: BorderRadius.circular(AppTheme.radiusFull),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.bug_report_outlined, size: 18, color: cs.onPrimary),
                const SizedBox(width: 6),
                Text(
                  l10n.feedbackButtonLabel,
                  style: SoleilTextStyles.button.copyWith(
                    color: cs.onPrimary,
                    fontSize: 13,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void _open(BuildContext context, WidgetRef ref) {
    // Resolve the current location from the router's delegate — the FAB
    // is above the navigator so `GoRouterState.of(context)` is unavailable.
    final router = ref.read(appRouterProvider);
    final pageUrl =
        router.routerDelegate.currentConfiguration.uri.toString();
    showFeedbackSheet(context, pageUrl: pageUrl);
  }
}
