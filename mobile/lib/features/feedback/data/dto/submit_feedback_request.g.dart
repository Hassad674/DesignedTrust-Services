// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'submit_feedback_request.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Map<String, dynamic> _$SubmitFeedbackRequestToJson(
        SubmitFeedbackRequest instance) =>
    <String, dynamic>{
      'type': instance.type,
      'title': instance.title,
      'description': instance.description,
      'page_url': instance.pageUrl,
      'context': instance.context?.toJson(),
      'reporter_email': instance.reporterEmail,
      'attachment_keys':
          instance.attachmentKeys.map((e) => e.toJson()).toList(),
      'hp': instance.hp,
    };

Map<String, dynamic> _$SubmitFeedbackAttachmentRequestToJson(
        SubmitFeedbackAttachmentRequest instance) =>
    <String, dynamic>{
      'kind': instance.kind,
      'object_key': instance.objectKey,
      'content_type': instance.contentType,
      'size_bytes': instance.sizeBytes,
    };

Map<String, dynamic> _$FeedbackContextRequestToJson(
        FeedbackContextRequest instance) =>
    <String, dynamic>{
      'role': instance.role,
      'locale': instance.locale,
      'platform': instance.platform,
      'app_version': instance.appVersion,
      'viewport': instance.viewport,
      'user_agent': instance.userAgent,
    };
