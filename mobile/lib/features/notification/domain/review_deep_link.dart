import 'entities/app_notification.dart';

/// Review deep-link helper for mobile in-app notification taps.
///
/// Mirrors the web `buildReviewDeepLink` contract: review-related
/// notification kinds (proposal_completed, review_received) ship a
/// proposal-flow payload (`proposal_id` + `conversation_id` +
/// `proposal_title`) and tapping them should open the conversation
/// route with the openReview flag so the chat screen auto-opens the
/// review sheet for the right proposal.
///
/// Returns `null` when the notification kind is not review-related OR
/// when the payload is missing/malformed — the caller (the
/// NotificationScreen tap handler) then falls back to plain
/// mark-as-read with no navigation.
const _reviewKinds = {
  'proposal_completed',
  'review_received',
};

class ReviewDeepLink {
  final String conversationId;
  final String proposalId;

  const ReviewDeepLink({
    required this.conversationId,
    required this.proposalId,
  });

  /// Builds the go_router-friendly path for the chat screen:
  /// `/chat/<convId>?openReview=1&reviewProposalId=<proposalId>`.
  String toRoute(String chatBasePath) {
    return '$chatBasePath/$conversationId'
        '?openReview=1&reviewProposalId=$proposalId';
  }
}

/// Reads the navigation hints from the notification payload. Mirrors
/// the web shape exactly so a notification produced by the backend
/// dispute / proposal flow yields a usable deep-link on both surfaces.
ReviewDeepLink? buildReviewDeepLink(AppNotification notification) {
  if (!_reviewKinds.contains(notification.type)) return null;

  final conversationId = _readString(notification.data, 'conversation_id');
  final proposalId = _readString(notification.data, 'proposal_id');
  if (conversationId == null || proposalId == null) return null;

  return ReviewDeepLink(
    conversationId: conversationId,
    proposalId: proposalId,
  );
}

String? _readString(Map<String, dynamic> data, String key) {
  final value = data[key];
  if (value is String && value.isNotEmpty) return value;
  return null;
}
