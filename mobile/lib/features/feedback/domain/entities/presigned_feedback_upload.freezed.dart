// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'presigned_feedback_upload.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

/// @nodoc
mixin _$PresignedFeedbackUpload {
  String get uploadUrl => throw _privateConstructorUsedError;
  String get objectKey => throw _privateConstructorUsedError;
  FeedbackAttachmentKind get kind => throw _privateConstructorUsedError;

  /// Create a copy of PresignedFeedbackUpload
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  $PresignedFeedbackUploadCopyWith<PresignedFeedbackUpload> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $PresignedFeedbackUploadCopyWith<$Res> {
  factory $PresignedFeedbackUploadCopyWith(PresignedFeedbackUpload value,
          $Res Function(PresignedFeedbackUpload) then) =
      _$PresignedFeedbackUploadCopyWithImpl<$Res, PresignedFeedbackUpload>;
  @useResult
  $Res call({String uploadUrl, String objectKey, FeedbackAttachmentKind kind});
}

/// @nodoc
class _$PresignedFeedbackUploadCopyWithImpl<$Res,
        $Val extends PresignedFeedbackUpload>
    implements $PresignedFeedbackUploadCopyWith<$Res> {
  _$PresignedFeedbackUploadCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  /// Create a copy of PresignedFeedbackUpload
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? uploadUrl = null,
    Object? objectKey = null,
    Object? kind = null,
  }) {
    return _then(_value.copyWith(
      uploadUrl: null == uploadUrl
          ? _value.uploadUrl
          : uploadUrl // ignore: cast_nullable_to_non_nullable
              as String,
      objectKey: null == objectKey
          ? _value.objectKey
          : objectKey // ignore: cast_nullable_to_non_nullable
              as String,
      kind: null == kind
          ? _value.kind
          : kind // ignore: cast_nullable_to_non_nullable
              as FeedbackAttachmentKind,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$PresignedFeedbackUploadImplCopyWith<$Res>
    implements $PresignedFeedbackUploadCopyWith<$Res> {
  factory _$$PresignedFeedbackUploadImplCopyWith(
          _$PresignedFeedbackUploadImpl value,
          $Res Function(_$PresignedFeedbackUploadImpl) then) =
      __$$PresignedFeedbackUploadImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({String uploadUrl, String objectKey, FeedbackAttachmentKind kind});
}

/// @nodoc
class __$$PresignedFeedbackUploadImplCopyWithImpl<$Res>
    extends _$PresignedFeedbackUploadCopyWithImpl<$Res,
        _$PresignedFeedbackUploadImpl>
    implements _$$PresignedFeedbackUploadImplCopyWith<$Res> {
  __$$PresignedFeedbackUploadImplCopyWithImpl(
      _$PresignedFeedbackUploadImpl _value,
      $Res Function(_$PresignedFeedbackUploadImpl) _then)
      : super(_value, _then);

  /// Create a copy of PresignedFeedbackUpload
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? uploadUrl = null,
    Object? objectKey = null,
    Object? kind = null,
  }) {
    return _then(_$PresignedFeedbackUploadImpl(
      uploadUrl: null == uploadUrl
          ? _value.uploadUrl
          : uploadUrl // ignore: cast_nullable_to_non_nullable
              as String,
      objectKey: null == objectKey
          ? _value.objectKey
          : objectKey // ignore: cast_nullable_to_non_nullable
              as String,
      kind: null == kind
          ? _value.kind
          : kind // ignore: cast_nullable_to_non_nullable
              as FeedbackAttachmentKind,
    ));
  }
}

/// @nodoc

class _$PresignedFeedbackUploadImpl implements _PresignedFeedbackUpload {
  const _$PresignedFeedbackUploadImpl(
      {required this.uploadUrl, required this.objectKey, required this.kind});

  @override
  final String uploadUrl;
  @override
  final String objectKey;
  @override
  final FeedbackAttachmentKind kind;

  @override
  String toString() {
    return 'PresignedFeedbackUpload(uploadUrl: $uploadUrl, objectKey: $objectKey, kind: $kind)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$PresignedFeedbackUploadImpl &&
            (identical(other.uploadUrl, uploadUrl) ||
                other.uploadUrl == uploadUrl) &&
            (identical(other.objectKey, objectKey) ||
                other.objectKey == objectKey) &&
            (identical(other.kind, kind) || other.kind == kind));
  }

  @override
  int get hashCode => Object.hash(runtimeType, uploadUrl, objectKey, kind);

  /// Create a copy of PresignedFeedbackUpload
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  @pragma('vm:prefer-inline')
  _$$PresignedFeedbackUploadImplCopyWith<_$PresignedFeedbackUploadImpl>
      get copyWith => __$$PresignedFeedbackUploadImplCopyWithImpl<
          _$PresignedFeedbackUploadImpl>(this, _$identity);
}

abstract class _PresignedFeedbackUpload implements PresignedFeedbackUpload {
  const factory _PresignedFeedbackUpload(
          {required final String uploadUrl,
          required final String objectKey,
          required final FeedbackAttachmentKind kind}) =
      _$PresignedFeedbackUploadImpl;

  @override
  String get uploadUrl;
  @override
  String get objectKey;
  @override
  FeedbackAttachmentKind get kind;

  /// Create a copy of PresignedFeedbackUpload
  /// with the given fields replaced by the non-null parameter values.
  @override
  @JsonKey(includeFromJson: false, includeToJson: false)
  _$$PresignedFeedbackUploadImplCopyWith<_$PresignedFeedbackUploadImpl>
      get copyWith => throw _privateConstructorUsedError;
}
