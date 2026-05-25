// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'feedback_attachment.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

/// @nodoc
mixin _$FeedbackAttachment {
  FeedbackAttachmentKind get kind => throw _privateConstructorUsedError;
  String get objectKey => throw _privateConstructorUsedError;
  String get contentType => throw _privateConstructorUsedError;
  int get sizeBytes => throw _privateConstructorUsedError;

  /// Create a copy of FeedbackAttachment
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  $FeedbackAttachmentCopyWith<FeedbackAttachment> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $FeedbackAttachmentCopyWith<$Res> {
  factory $FeedbackAttachmentCopyWith(
          FeedbackAttachment value, $Res Function(FeedbackAttachment) then) =
      _$FeedbackAttachmentCopyWithImpl<$Res, FeedbackAttachment>;
  @useResult
  $Res call(
      {FeedbackAttachmentKind kind,
      String objectKey,
      String contentType,
      int sizeBytes});
}

/// @nodoc
class _$FeedbackAttachmentCopyWithImpl<$Res, $Val extends FeedbackAttachment>
    implements $FeedbackAttachmentCopyWith<$Res> {
  _$FeedbackAttachmentCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  /// Create a copy of FeedbackAttachment
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? kind = null,
    Object? objectKey = null,
    Object? contentType = null,
    Object? sizeBytes = null,
  }) {
    return _then(_value.copyWith(
      kind: null == kind
          ? _value.kind
          : kind // ignore: cast_nullable_to_non_nullable
              as FeedbackAttachmentKind,
      objectKey: null == objectKey
          ? _value.objectKey
          : objectKey // ignore: cast_nullable_to_non_nullable
              as String,
      contentType: null == contentType
          ? _value.contentType
          : contentType // ignore: cast_nullable_to_non_nullable
              as String,
      sizeBytes: null == sizeBytes
          ? _value.sizeBytes
          : sizeBytes // ignore: cast_nullable_to_non_nullable
              as int,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$FeedbackAttachmentImplCopyWith<$Res>
    implements $FeedbackAttachmentCopyWith<$Res> {
  factory _$$FeedbackAttachmentImplCopyWith(_$FeedbackAttachmentImpl value,
          $Res Function(_$FeedbackAttachmentImpl) then) =
      __$$FeedbackAttachmentImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {FeedbackAttachmentKind kind,
      String objectKey,
      String contentType,
      int sizeBytes});
}

/// @nodoc
class __$$FeedbackAttachmentImplCopyWithImpl<$Res>
    extends _$FeedbackAttachmentCopyWithImpl<$Res, _$FeedbackAttachmentImpl>
    implements _$$FeedbackAttachmentImplCopyWith<$Res> {
  __$$FeedbackAttachmentImplCopyWithImpl(_$FeedbackAttachmentImpl _value,
      $Res Function(_$FeedbackAttachmentImpl) _then)
      : super(_value, _then);

  /// Create a copy of FeedbackAttachment
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? kind = null,
    Object? objectKey = null,
    Object? contentType = null,
    Object? sizeBytes = null,
  }) {
    return _then(_$FeedbackAttachmentImpl(
      kind: null == kind
          ? _value.kind
          : kind // ignore: cast_nullable_to_non_nullable
              as FeedbackAttachmentKind,
      objectKey: null == objectKey
          ? _value.objectKey
          : objectKey // ignore: cast_nullable_to_non_nullable
              as String,
      contentType: null == contentType
          ? _value.contentType
          : contentType // ignore: cast_nullable_to_non_nullable
              as String,
      sizeBytes: null == sizeBytes
          ? _value.sizeBytes
          : sizeBytes // ignore: cast_nullable_to_non_nullable
              as int,
    ));
  }
}

/// @nodoc

class _$FeedbackAttachmentImpl implements _FeedbackAttachment {
  const _$FeedbackAttachmentImpl(
      {required this.kind,
      required this.objectKey,
      required this.contentType,
      required this.sizeBytes});

  @override
  final FeedbackAttachmentKind kind;
  @override
  final String objectKey;
  @override
  final String contentType;
  @override
  final int sizeBytes;

  @override
  String toString() {
    return 'FeedbackAttachment(kind: $kind, objectKey: $objectKey, contentType: $contentType, sizeBytes: $sizeBytes)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$FeedbackAttachmentImpl &&
            (identical(other.kind, kind) || other.kind == kind) &&
            (identical(other.objectKey, objectKey) ||
                other.objectKey == objectKey) &&
            (identical(other.contentType, contentType) ||
                other.contentType == contentType) &&
            (identical(other.sizeBytes, sizeBytes) ||
                other.sizeBytes == sizeBytes));
  }

  @override
  int get hashCode =>
      Object.hash(runtimeType, kind, objectKey, contentType, sizeBytes);

  /// Create a copy of FeedbackAttachment
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  @pragma('vm:prefer-inline')
  _$$FeedbackAttachmentImplCopyWith<_$FeedbackAttachmentImpl> get copyWith =>
      __$$FeedbackAttachmentImplCopyWithImpl<_$FeedbackAttachmentImpl>(
          this, _$identity);
}

abstract class _FeedbackAttachment implements FeedbackAttachment {
  const factory _FeedbackAttachment(
      {required final FeedbackAttachmentKind kind,
      required final String objectKey,
      required final String contentType,
      required final int sizeBytes}) = _$FeedbackAttachmentImpl;

  @override
  FeedbackAttachmentKind get kind;
  @override
  String get objectKey;
  @override
  String get contentType;
  @override
  int get sizeBytes;

  /// Create a copy of FeedbackAttachment
  /// with the given fields replaced by the non-null parameter values.
  @override
  @JsonKey(includeFromJson: false, includeToJson: false)
  _$$FeedbackAttachmentImplCopyWith<_$FeedbackAttachmentImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
