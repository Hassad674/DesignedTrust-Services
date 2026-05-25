import 'dart:io';
import 'dart:typed_data';

import 'package:dio/dio.dart';

import '../../../../core/network/api_client.dart';
import '../../domain/entities/feedback_attachment.dart';
import '../../domain/entities/feedback_attachment_kind.dart';
import '../../domain/entities/feedback_submission.dart';
import '../../domain/entities/presigned_feedback_upload.dart';
import '../../domain/entities/submit_feedback_input.dart';
import '../../domain/repositories/feedback_repository.dart';
import '../dto/presign_feedback_attachment_request.dart';
import '../dto/presign_feedback_attachment_response.dart';
import '../dto/submit_feedback_request.dart';
import '../dto/submit_feedback_response.dart';

/// Concrete [FeedbackRepository] backed by the Go API.
///
/// The submit + presign calls go through the authenticated [ApiClient]
/// (the submit endpoint is anonymous-friendly — the interceptor simply
/// omits the bearer when no token is stored). The raw-bytes PUT to the
/// presigned storage URL uses a SEPARATE Dio with no JWT interceptor,
/// mirroring `chat_file_uploader.dart` / `dispute_uploader.dart`: the
/// presigned URL is signature-authenticated, so forwarding the bearer
/// cross-origin would be both pointless and a token leak.
class FeedbackRepositoryImpl implements FeedbackRepository {
  FeedbackRepositoryImpl(this._api, {Dio? uploadDio})
      : _uploadDio = uploadDio ??
            Dio(
              BaseOptions(
                connectTimeout: const Duration(seconds: 30),
                sendTimeout: const Duration(seconds: 120),
                receiveTimeout: const Duration(seconds: 30),
              ),
            );

  final ApiClient _api;
  final Dio _uploadDio;

  @override
  Future<FeedbackSubmission> submit(SubmitFeedbackInput input) async {
    final body = SubmitFeedbackRequest.fromDomain(input).toJson();
    final response = await _api.post<Map<String, dynamic>>(
      '/api/v1/feedback',
      data: body,
    );
    final data = _unwrap(response.data);
    return SubmitFeedbackResponse.fromJson(data).toDomain();
  }

  @override
  Future<FeedbackAttachment> uploadAttachment({
    required File file,
    required FeedbackAttachmentKind kind,
    required String contentType,
    required String filename,
  }) async {
    final bytes = await file.readAsBytes();

    // Step 1 — presign (auth required; throws DioException 401 if the
    // caller is somehow anonymous, which the UI prevents upstream).
    final presigned = await _presign(
      kind: kind,
      contentType: contentType,
      sizeBytes: bytes.length,
      filename: filename,
    );

    // Step 2 — PUT the bytes to the short-lived storage URL.
    await _putBytes(presigned.uploadUrl, bytes, contentType);

    // Step 3 — hand back the reference to echo in the submit payload.
    return FeedbackAttachment(
      kind: presigned.kind,
      objectKey: presigned.objectKey,
      contentType: contentType,
      sizeBytes: bytes.length,
    );
  }

  Future<PresignedFeedbackUpload> _presign({
    required FeedbackAttachmentKind kind,
    required String contentType,
    required int sizeBytes,
    required String filename,
  }) async {
    final body = PresignFeedbackAttachmentRequest(
      kind: kind.wireValue,
      contentType: contentType,
      sizeBytes: sizeBytes,
      filename: filename,
    ).toJson();
    final response = await _api.post<Map<String, dynamic>>(
      '/api/v1/feedback/attachments/presign',
      data: body,
    );
    final data = _unwrap(response.data);
    return PresignFeedbackAttachmentResponse.fromJson(data).toDomain();
  }

  Future<void> _putBytes(
    String uploadUrl,
    Uint8List bytes,
    String contentType,
  ) async {
    await _uploadDio.put<void>(
      uploadUrl,
      data: Stream.fromIterable([bytes]),
      options: Options(
        contentType: contentType,
        headers: {Headers.contentLengthHeader: bytes.length},
      ),
    );
  }

  /// Unwraps the `{ data, meta }` envelope. Accepts a bare object too so
  /// the repository stays robust if a handler ever returns the payload
  /// flat (some legacy endpoints do).
  Map<String, dynamic> _unwrap(Map<String, dynamic>? body) {
    if (body == null) {
      throw StateError('feedback response body is empty');
    }
    final data = body['data'];
    if (data is Map<String, dynamic>) {
      return data;
    }
    return body;
  }
}
