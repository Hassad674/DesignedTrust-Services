import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/feedback/data/dto/presign_feedback_attachment_request.dart';
import 'package:marketplace_mobile/features/feedback/data/dto/presign_feedback_attachment_response.dart';
import 'package:marketplace_mobile/features/feedback/data/dto/submit_feedback_request.dart';
import 'package:marketplace_mobile/features/feedback/data/dto/submit_feedback_response.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_attachment.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_attachment_kind.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_type.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/submit_feedback_input.dart';

void main() {
  group('SubmitFeedbackRequest wire shape', () {
    test('serializes every required key with snake_case names', () {
      final json = SubmitFeedbackRequest.fromDomain(
        const SubmitFeedbackInput(
          type: FeedbackType.security,
          title: 'Title',
          description: 'Desc',
          pageUrl: '/missions',
          reporterEmail: 'me@example.com',
          context: FeedbackContextInput(
            role: 'provider',
            locale: 'fr',
            platform: 'android',
          ),
          attachments: [
            FeedbackAttachment(
              kind: FeedbackAttachmentKind.image,
              objectKey: 'feedback/abc.png',
              contentType: 'image/png',
              sizeBytes: 4096,
            ),
          ],
          honeypot: '',
        ),
      ).toJson();

      expect(json['type'], 'security');
      expect(json['title'], 'Title');
      expect(json['description'], 'Desc');
      expect(json['page_url'], '/missions');
      expect(json['reporter_email'], 'me@example.com');
      expect(json['hp'], '');

      final ctx = json['context'] as Map<String, dynamic>;
      expect(ctx['role'], 'provider');
      expect(ctx['locale'], 'fr');
      expect(ctx['platform'], 'android');
      // Native client cannot supply these — present but empty.
      expect(ctx['app_version'], '');
      expect(ctx['viewport'], '');
      expect(ctx['user_agent'], '');

      final keys = json['attachment_keys'] as List<dynamic>;
      expect(keys, hasLength(1));
      final first = keys.first as Map<String, dynamic>;
      expect(first['kind'], 'image');
      expect(first['object_key'], 'feedback/abc.png');
      expect(first['content_type'], 'image/png');
      expect(first['size_bytes'], 4096);
    });

    test('context serializes to null when absent', () {
      final json = SubmitFeedbackRequest.fromDomain(
        const SubmitFeedbackInput(
          type: FeedbackType.bug,
          title: 'T',
          description: 'D',
        ),
      ).toJson();
      expect(json.containsKey('context'), isTrue);
      expect(json['context'], isNull);
      expect(json['attachment_keys'], isEmpty);
      expect(json['type'], 'bug');
    });
  });

  group('PresignFeedbackAttachmentRequest wire shape', () {
    test('serializes snake_case keys', () {
      final json = const PresignFeedbackAttachmentRequest(
        kind: 'video',
        contentType: 'video/mp4',
        sizeBytes: 999,
        filename: 'clip.mp4',
      ).toJson();
      expect(json['kind'], 'video');
      expect(json['content_type'], 'video/mp4');
      expect(json['size_bytes'], 999);
      expect(json['filename'], 'clip.mp4');
    });
  });

  group('response decoding', () {
    test('SubmitFeedbackResponse maps to domain with parsed date + type', () {
      final domain = SubmitFeedbackResponse.fromJson(<String, dynamic>{
        'id': 'fb_42',
        'type': 'security',
        'status': 'triaged',
        'created_at': '2026-05-25T10:30:00Z',
      }).toDomain();
      expect(domain.id, 'fb_42');
      expect(domain.type, FeedbackType.security);
      expect(domain.status, 'triaged');
      expect(domain.createdAt.toUtc().year, 2026);
    });

    test('unknown type falls back to bug (forward-compatible)', () {
      final domain = SubmitFeedbackResponse.fromJson(<String, dynamic>{
        'id': 'x',
        'type': 'something_new',
        'status': 'new',
        'created_at': '2026-05-25T10:30:00Z',
      }).toDomain();
      expect(domain.type, FeedbackType.bug);
    });

    test('PresignFeedbackAttachmentResponse maps url + object_key + kind', () {
      final domain = PresignFeedbackAttachmentResponse.fromJson(
        <String, dynamic>{
          'url': 'https://r2.example/put?sig=1',
          'object_key': 'feedback/u1/abc',
          'kind': 'video',
        },
      ).toDomain();
      expect(domain.uploadUrl, 'https://r2.example/put?sig=1');
      expect(domain.objectKey, 'feedback/u1/abc');
      expect(domain.kind, FeedbackAttachmentKind.video);
    });
  });
}
