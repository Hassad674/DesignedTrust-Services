// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'presign_feedback_attachment_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

PresignFeedbackAttachmentResponse _$PresignFeedbackAttachmentResponseFromJson(
        Map<String, dynamic> json) =>
    PresignFeedbackAttachmentResponse(
      uploadUrl: json['url'] as String,
      objectKey: json['object_key'] as String,
      kind: json['kind'] as String,
    );

Map<String, dynamic> _$PresignFeedbackAttachmentResponseToJson(
        PresignFeedbackAttachmentResponse instance) =>
    <String, dynamic>{
      'url': instance.uploadUrl,
      'object_key': instance.objectKey,
      'kind': instance.kind,
    };
