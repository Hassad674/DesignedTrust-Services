// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'feedback_submission.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

/// @nodoc
mixin _$FeedbackSubmission {
  String get id => throw _privateConstructorUsedError;
  FeedbackType get type => throw _privateConstructorUsedError;
  String get status => throw _privateConstructorUsedError;
  DateTime get createdAt => throw _privateConstructorUsedError;

  /// Create a copy of FeedbackSubmission
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  $FeedbackSubmissionCopyWith<FeedbackSubmission> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $FeedbackSubmissionCopyWith<$Res> {
  factory $FeedbackSubmissionCopyWith(
          FeedbackSubmission value, $Res Function(FeedbackSubmission) then) =
      _$FeedbackSubmissionCopyWithImpl<$Res, FeedbackSubmission>;
  @useResult
  $Res call({String id, FeedbackType type, String status, DateTime createdAt});
}

/// @nodoc
class _$FeedbackSubmissionCopyWithImpl<$Res, $Val extends FeedbackSubmission>
    implements $FeedbackSubmissionCopyWith<$Res> {
  _$FeedbackSubmissionCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  /// Create a copy of FeedbackSubmission
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? type = null,
    Object? status = null,
    Object? createdAt = null,
  }) {
    return _then(_value.copyWith(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      type: null == type
          ? _value.type
          : type // ignore: cast_nullable_to_non_nullable
              as FeedbackType,
      status: null == status
          ? _value.status
          : status // ignore: cast_nullable_to_non_nullable
              as String,
      createdAt: null == createdAt
          ? _value.createdAt
          : createdAt // ignore: cast_nullable_to_non_nullable
              as DateTime,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$FeedbackSubmissionImplCopyWith<$Res>
    implements $FeedbackSubmissionCopyWith<$Res> {
  factory _$$FeedbackSubmissionImplCopyWith(_$FeedbackSubmissionImpl value,
          $Res Function(_$FeedbackSubmissionImpl) then) =
      __$$FeedbackSubmissionImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({String id, FeedbackType type, String status, DateTime createdAt});
}

/// @nodoc
class __$$FeedbackSubmissionImplCopyWithImpl<$Res>
    extends _$FeedbackSubmissionCopyWithImpl<$Res, _$FeedbackSubmissionImpl>
    implements _$$FeedbackSubmissionImplCopyWith<$Res> {
  __$$FeedbackSubmissionImplCopyWithImpl(_$FeedbackSubmissionImpl _value,
      $Res Function(_$FeedbackSubmissionImpl) _then)
      : super(_value, _then);

  /// Create a copy of FeedbackSubmission
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? type = null,
    Object? status = null,
    Object? createdAt = null,
  }) {
    return _then(_$FeedbackSubmissionImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      type: null == type
          ? _value.type
          : type // ignore: cast_nullable_to_non_nullable
              as FeedbackType,
      status: null == status
          ? _value.status
          : status // ignore: cast_nullable_to_non_nullable
              as String,
      createdAt: null == createdAt
          ? _value.createdAt
          : createdAt // ignore: cast_nullable_to_non_nullable
              as DateTime,
    ));
  }
}

/// @nodoc

class _$FeedbackSubmissionImpl implements _FeedbackSubmission {
  const _$FeedbackSubmissionImpl(
      {required this.id,
      required this.type,
      required this.status,
      required this.createdAt});

  @override
  final String id;
  @override
  final FeedbackType type;
  @override
  final String status;
  @override
  final DateTime createdAt;

  @override
  String toString() {
    return 'FeedbackSubmission(id: $id, type: $type, status: $status, createdAt: $createdAt)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$FeedbackSubmissionImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.type, type) || other.type == type) &&
            (identical(other.status, status) || other.status == status) &&
            (identical(other.createdAt, createdAt) ||
                other.createdAt == createdAt));
  }

  @override
  int get hashCode => Object.hash(runtimeType, id, type, status, createdAt);

  /// Create a copy of FeedbackSubmission
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  @pragma('vm:prefer-inline')
  _$$FeedbackSubmissionImplCopyWith<_$FeedbackSubmissionImpl> get copyWith =>
      __$$FeedbackSubmissionImplCopyWithImpl<_$FeedbackSubmissionImpl>(
          this, _$identity);
}

abstract class _FeedbackSubmission implements FeedbackSubmission {
  const factory _FeedbackSubmission(
      {required final String id,
      required final FeedbackType type,
      required final String status,
      required final DateTime createdAt}) = _$FeedbackSubmissionImpl;

  @override
  String get id;
  @override
  FeedbackType get type;
  @override
  String get status;
  @override
  DateTime get createdAt;

  /// Create a copy of FeedbackSubmission
  /// with the given fields replaced by the non-null parameter values.
  @override
  @JsonKey(includeFromJson: false, includeToJson: false)
  _$$FeedbackSubmissionImplCopyWith<_$FeedbackSubmissionImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
