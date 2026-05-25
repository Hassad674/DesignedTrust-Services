import 'package:flutter/material.dart';

import 'feedback_fab.dart';

/// Wraps the whole app (via `MaterialApp.builder`) and paints the
/// always-visible [FeedbackFab] above [child].
///
/// Pure and feature-isolated: it knows nothing about other features.
/// The composition root (`main.dart`) decides when to hide it — e.g.
/// while a full-screen call is active — by toggling [hidden], so this
/// widget never imports the call feature.
///
/// The FAB is anchored bottom-right and lifted above the bottom
/// navigation bar / home indicator so it does not collide with either.
class FeedbackOverlay extends StatelessWidget {
  const FeedbackOverlay({
    super.key,
    required this.child,
    this.hidden = false,
  });

  final Widget child;

  /// When true the FAB is not painted (the rest of the tree is
  /// untouched). Used to step aside for full-screen surfaces such as an
  /// active call.
  final bool hidden;

  /// Distance from the bottom of the screen — clears a standard
  /// Material bottom nav (~80) plus the safe-area home indicator.
  static const double _bottomInset = 92;
  static const double _rightInset = 16;

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        child,
        if (!hidden)
          Positioned(
            right: _rightInset,
            bottom: _bottomInset + MediaQuery.viewPaddingOf(context).bottom,
            child: const FeedbackFab(),
          ),
      ],
    );
  }
}
