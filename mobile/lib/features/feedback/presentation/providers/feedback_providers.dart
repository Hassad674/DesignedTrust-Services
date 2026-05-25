import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/network/api_client.dart';
import '../../data/repositories/feedback_repository_impl.dart';
import '../../domain/entities/submit_feedback_input.dart';
import '../../domain/repositories/feedback_repository.dart';

/// Provides the concrete [FeedbackRepository] wired with the Dio
/// [ApiClient]. Scoped to the app lifecycle, like every other
/// repository provider in this codebase.
final feedbackRepositoryProvider = Provider<FeedbackRepository>((ref) {
  final api = ref.watch(apiClientProvider);
  return FeedbackRepositoryImpl(api);
});

/// Drives the submit phase of the feedback sheet.
///
/// The sheet owns its own form fields (text controllers, type toggle,
/// picked attachments) as local widget state — only the network submit
/// outcome flows through Riverpod so the button can render
/// loading / error / success without the widget tracking it manually.
final submitFeedbackProvider =
    AsyncNotifierProvider.autoDispose<SubmitFeedbackNotifier, void>(
  SubmitFeedbackNotifier.new,
);

/// AsyncNotifier whose `state` is the in-flight / done / failed status
/// of the most recent submit. `void` data because a successful submit
/// carries no screen-level payload — the sheet closes and a snackbar
/// confirms. Errors surface via [AsyncValue.error] for inline display.
class SubmitFeedbackNotifier extends AutoDisposeAsyncNotifier<void> {
  @override
  Future<void> build() async {
    // Idle until the user submits — no work on construction.
  }

  /// Files the report. On success leaves the state as `AsyncData(null)`;
  /// on failure as `AsyncError` so the sheet can show an inline message
  /// and re-enable the submit button. Returns true on success so the
  /// caller can close the sheet without re-reading the state.
  Future<bool> submit(SubmitFeedbackInput input) async {
    state = const AsyncLoading();
    final repo = ref.read(feedbackRepositoryProvider);
    state = await AsyncValue.guard(() => repo.submit(input));
    return !state.hasError;
  }
}
