import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:marketplace_mobile/features/feedback/data/repositories/feedback_repository_impl.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_attachment_kind.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/feedback_type.dart';
import 'package:marketplace_mobile/features/feedback/domain/entities/submit_feedback_input.dart';

import '../../helpers/fake_api_client.dart';

/// A Dio [HttpClientAdapter] that records the single PUT it receives and
/// returns 200 — stands in for the R2 presigned-URL storage endpoint.
class _RecordingPutAdapter implements HttpClientAdapter {
  String? putUrl;
  String? putContentType;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    putUrl = options.uri.toString();
    putContentType = options.contentType;
    // Drain the stream so the request "completes".
    if (requestStream != null) {
      await requestStream.drain<void>();
    }
    return ResponseBody.fromString('', 200);
  }

  @override
  void close({bool force = false}) {}
}

Response<Map<String, dynamic>> _ok(Map<String, dynamic> data, String path) {
  return Response<Map<String, dynamic>>(
    requestOptions: RequestOptions(path: path),
    statusCode: 200,
    data: {'data': data, 'meta': const {}},
  );
}

void main() {
  late File tempFile;

  setUp(() async {
    tempFile = File(
      '${Directory.systemTemp.path}/feedback_test_${DateTime.now().microsecondsSinceEpoch}.png',
    );
    await tempFile.writeAsBytes(List<int>.filled(64, 7));
  });

  tearDown(() async {
    if (tempFile.existsSync()) await tempFile.delete();
  });

  test('submit posts to /api/v1/feedback and unwraps the data envelope',
      () async {
    final api = FakeApiClient();
    Map<String, dynamic>? sentBody;
    api.postHandlers['/api/v1/feedback'] = (data) async {
      sentBody = data as Map<String, dynamic>;
      return _ok({
        'id': 'fb_1',
        'type': 'bug',
        'status': 'new',
        'created_at': '2026-05-25T09:00:00Z',
      }, '/api/v1/feedback',);
    };

    final repo = FeedbackRepositoryImpl(api, uploadDio: Dio());
    final result = await repo.submit(
      const SubmitFeedbackInput(
        type: FeedbackType.bug,
        title: 'T',
        description: 'D',
        pageUrl: '/dashboard',
      ),
    );

    expect(result.id, 'fb_1');
    expect(result.type, FeedbackType.bug);
    expect(result.status, 'new');
    // Payload reached the endpoint with the contract keys.
    expect(sentBody!['title'], 'T');
    expect(sentBody!['page_url'], '/dashboard');
    expect(sentBody!['hp'], '');
  });

  test('uploadAttachment presigns then PUTs the bytes to the storage URL',
      () async {
    final api = FakeApiClient();
    Map<String, dynamic>? presignBody;
    api.postHandlers['/api/v1/feedback/attachments/presign'] = (data) async {
      presignBody = data as Map<String, dynamic>;
      return _ok({
        'url': 'https://r2.example/put?sig=abc',
        'object_key': 'feedback/u1/xyz.png',
        'kind': 'image',
      }, '/api/v1/feedback/attachments/presign',);
    };

    final putAdapter = _RecordingPutAdapter();
    final uploadDio = Dio()..httpClientAdapter = putAdapter;
    final repo = FeedbackRepositoryImpl(api, uploadDio: uploadDio);

    final attachment = await repo.uploadAttachment(
      file: tempFile,
      kind: FeedbackAttachmentKind.image,
      contentType: 'image/png',
      filename: 'shot.png',
    );

    // Presign request carried the right metadata.
    expect(presignBody!['kind'], 'image');
    expect(presignBody!['content_type'], 'image/png');
    expect(presignBody!['size_bytes'], 64);
    expect(presignBody!['filename'], 'shot.png');

    // Bytes PUT to the exact presigned URL with the right content type.
    expect(putAdapter.putUrl, 'https://r2.example/put?sig=abc');
    expect(putAdapter.putContentType, 'image/png');

    // The returned reference echoes the server-minted object key.
    expect(attachment.objectKey, 'feedback/u1/xyz.png');
    expect(attachment.kind, FeedbackAttachmentKind.image);
    expect(attachment.sizeBytes, 64);
    expect(attachment.contentType, 'image/png');
  });

  test('uploadAttachment surfaces a 401 from the presign endpoint', () async {
    final api = FakeApiClient();
    api.postHandlers['/api/v1/feedback/attachments/presign'] = (_) async {
      throw DioException(
        requestOptions:
            RequestOptions(path: '/api/v1/feedback/attachments/presign'),
        response: Response<dynamic>(
          requestOptions:
              RequestOptions(path: '/api/v1/feedback/attachments/presign'),
          statusCode: 401,
        ),
        type: DioExceptionType.badResponse,
      );
    };

    final repo = FeedbackRepositoryImpl(api, uploadDio: Dio());
    await expectLater(
      repo.uploadAttachment(
        file: tempFile,
        kind: FeedbackAttachmentKind.image,
        contentType: 'image/png',
        filename: 'shot.png',
      ),
      throwsA(isA<DioException>()),
    );
  });
}
