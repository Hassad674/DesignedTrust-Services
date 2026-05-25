// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'submit_feedback_response.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

SubmitFeedbackResponse _$SubmitFeedbackResponseFromJson(
        Map<String, dynamic> json) =>
    SubmitFeedbackResponse(
      id: json['id'] as String,
      type: json['type'] as String,
      status: json['status'] as String,
      createdAt: json['created_at'] as String,
    );

Map<String, dynamic> _$SubmitFeedbackResponseToJson(
        SubmitFeedbackResponse instance) =>
    <String, dynamic>{
      'id': instance.id,
      'type': instance.type,
      'status': instance.status,
      'created_at': instance.createdAt,
    };
