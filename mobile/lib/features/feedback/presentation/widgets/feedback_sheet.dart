import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';

import '../../../../core/theme/app_theme.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../auth/presentation/providers/auth_provider.dart';
import '../../domain/entities/feedback_attachment.dart';
import '../../domain/entities/feedback_attachment_kind.dart';
import '../../domain/entities/feedback_type.dart';
import '../../domain/entities/submit_feedback_input.dart';
import '../providers/feedback_providers.dart';
import 'feedback_attachments_section.dart';
import 'feedback_type_toggle.dart';

/// Opens the feedback ("Signaler") bottom sheet. [pageUrl] is the
/// current route captured by the caller (the global FAB) before opening
/// — the sheet sits above the router so it cannot read the location
/// itself.
Future<void> showFeedbackSheet(
  BuildContext context, {
  required String pageUrl,
}) {
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    backgroundColor: Theme.of(context).colorScheme.surfaceContainerLowest,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    builder: (_) => FeedbackSheet(pageUrl: pageUrl),
  );
}

/// Bug & vulnerability reporting form. Type toggle (Bug / Security),
/// title, description, optional email, and — for logged-in reporters —
/// image/video attachments uploaded via presign + PUT before submit.
class FeedbackSheet extends ConsumerStatefulWidget {
  const FeedbackSheet({super.key, required this.pageUrl});

  final String pageUrl;

  @override
  ConsumerState<FeedbackSheet> createState() => _FeedbackSheetState();
}

class _FeedbackSheetState extends ConsumerState<FeedbackSheet> {
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _emailController = TextEditingController();
  final _picker = ImagePicker();

  FeedbackType _type = FeedbackType.bug;
  final List<FeedbackAttachment> _attachments = [];
  bool _isUploading = false;
  bool _showErrors = false;

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    _emailController.dispose();
    super.dispose();
  }

  bool get _isLoggedIn =>
      ref.read(authProvider).status == AuthStatus.authenticated;

  Future<void> _pickImage() => _pickAndUpload(FeedbackAttachmentKind.image);

  Future<void> _pickVideo() => _pickAndUpload(FeedbackAttachmentKind.video);

  /// Picks one media file then presigns + PUTs it, appending the
  /// resulting reference to [_attachments]. Failures surface a snackbar
  /// and never crash the sheet.
  Future<void> _pickAndUpload(FeedbackAttachmentKind kind) async {
    final XFile? picked = kind == FeedbackAttachmentKind.image
        ? await _picker.pickImage(source: ImageSource.gallery)
        : await _picker.pickVideo(source: ImageSource.gallery);
    if (picked == null) return;

    setState(() => _isUploading = true);
    try {
      final attachment =
          await ref.read(feedbackRepositoryProvider).uploadAttachment(
                file: File(picked.path),
                kind: kind,
                contentType: _contentTypeFor(kind, picked),
                filename: picked.name,
              );
      if (!mounted) return;
      setState(() => _attachments.add(attachment));
    } catch (_) {
      if (!mounted) return;
      _snack(AppLocalizations.of(context)!.feedbackUploadError, isError: true);
    } finally {
      if (mounted) setState(() => _isUploading = false);
    }
  }

  Future<void> _submit() async {
    final title = _titleController.text.trim();
    final description = _descriptionController.text.trim();
    if (title.isEmpty || description.isEmpty) {
      setState(() => _showErrors = true);
      return;
    }

    final l10n = AppLocalizations.of(context)!;
    final ok = await ref.read(submitFeedbackProvider.notifier).submit(
          _buildInput(title, description),
        );
    if (!mounted) return;
    if (ok) {
      Navigator.of(context).pop();
      _snack(l10n.feedbackSuccess);
    } else {
      _snack(l10n.feedbackError, isError: true);
    }
  }

  SubmitFeedbackInput _buildInput(String title, String description) {
    final locale = Localizations.localeOf(context).languageCode;
    final role = ref.read(authProvider).user?['role'] as String? ?? '';
    return SubmitFeedbackInput(
      type: _type,
      title: title,
      description: description,
      pageUrl: widget.pageUrl,
      reporterEmail: _emailController.text.trim(),
      context: FeedbackContextInput(
        role: role,
        locale: locale,
        platform: 'android',
      ),
      // Attachments only ever populated for a logged-in reporter (the
      // attach controls are disabled otherwise).
      attachments: _isLoggedIn ? List.unmodifiable(_attachments) : const [],
      honeypot: '',
    );
  }

  void _snack(String message, {bool isError = false}) {
    final cs = Theme.of(context).colorScheme;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: isError ? cs.error : cs.primary,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final isSubmitting = ref.watch(submitFeedbackProvider).isLoading;
    return Padding(
      padding: EdgeInsets.only(
        left: 20,
        right: 20,
        top: 16,
        bottom: MediaQuery.of(context).viewInsets.bottom + 20,
      ),
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const _Grabber(),
            const SizedBox(height: 16),
            _Header(),
            const SizedBox(height: 16),
            FeedbackTypeToggle(
              value: _type,
              onChanged: (next) => setState(() => _type = next),
            ),
            const SizedBox(height: 16),
            _TitleField(controller: _titleController, showErrors: _showErrors),
            const SizedBox(height: 12),
            _DescriptionField(
              controller: _descriptionController,
              showErrors: _showErrors,
            ),
            const SizedBox(height: 12),
            _EmailField(controller: _emailController),
            const SizedBox(height: 16),
            FeedbackAttachmentsSection(
              canAttach: _isLoggedIn,
              attachments: _attachments,
              isUploading: _isUploading,
              onPickImage: _pickImage,
              onPickVideo: _pickVideo,
              onRemove: (a) => setState(() => _attachments.remove(a)),
            ),
            const SizedBox(height: 20),
            _SubmitButton(
              isBusy: isSubmitting || _isUploading,
              onPressed: _submit,
            ),
          ],
        ),
      ),
    );
  }

  /// Best-effort MIME from the picker. image_picker returns a `mimeType`
  /// on some platforms; otherwise we derive a sensible default from the
  /// extension so the presign content-type allowlist passes.
  String _contentTypeFor(FeedbackAttachmentKind kind, XFile file) {
    final mime = file.mimeType;
    if (mime != null && mime.isNotEmpty) return mime;
    final ext = file.name.split('.').last.toLowerCase();
    if (kind == FeedbackAttachmentKind.image) {
      return switch (ext) {
        'png' => 'image/png',
        'webp' => 'image/webp',
        'gif' => 'image/gif',
        _ => 'image/jpeg',
      };
    }
    return switch (ext) {
      'mov' => 'video/quicktime',
      'webm' => 'video/webm',
      _ => 'video/mp4',
    };
  }
}

// ---------------------------------------------------------------------------
// Private sub-widgets — keep the build() method lean.
// ---------------------------------------------------------------------------

class _Grabber extends StatelessWidget {
  const _Grabber();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 40,
        height: 4,
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.outline,
          borderRadius: BorderRadius.circular(2),
        ),
      ),
    );
  }
}

class _Header extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final cs = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.feedbackSheetTitle,
          style: SoleilTextStyles.headlineMedium.copyWith(
            color: cs.onSurface,
            fontSize: 22,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          l10n.feedbackSheetSubtitle,
          style: SoleilTextStyles.body.copyWith(color: cs.onSurfaceVariant),
        ),
      ],
    );
  }
}

class _TitleField extends StatelessWidget {
  const _TitleField({required this.controller, required this.showErrors});

  final TextEditingController controller;
  final bool showErrors;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return TextField(
      controller: controller,
      maxLength: 200,
      buildCounter: _noCounter,
      decoration: InputDecoration(
        labelText: l10n.feedbackTitleLabel,
        hintText: l10n.feedbackTitleHint,
        errorText: showErrors && controller.text.trim().isEmpty
            ? l10n.feedbackTitleRequired
            : null,
      ),
    );
  }
}

class _DescriptionField extends StatelessWidget {
  const _DescriptionField({
    required this.controller,
    required this.showErrors,
  });

  final TextEditingController controller;
  final bool showErrors;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return TextField(
      controller: controller,
      maxLines: 4,
      maxLength: 5000,
      buildCounter: _noCounter,
      decoration: InputDecoration(
        labelText: l10n.feedbackDescriptionLabel,
        hintText: l10n.feedbackDescriptionHint,
        alignLabelWithHint: true,
        errorText: showErrors && controller.text.trim().isEmpty
            ? l10n.feedbackDescriptionRequired
            : null,
      ),
    );
  }
}

class _EmailField extends StatelessWidget {
  const _EmailField({required this.controller});

  final TextEditingController controller;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return TextField(
      controller: controller,
      keyboardType: TextInputType.emailAddress,
      decoration: InputDecoration(
        labelText: l10n.feedbackEmailLabel,
        hintText: l10n.feedbackEmailHint,
      ),
    );
  }
}

class _SubmitButton extends StatelessWidget {
  const _SubmitButton({required this.isBusy, required this.onPressed});

  final bool isBusy;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final cs = Theme.of(context).colorScheme;
    final soleil = Theme.of(context).extension<AppColors>()!;
    return SizedBox(
      width: double.infinity,
      child: FilledButton(
        onPressed: isBusy ? null : onPressed,
        style: FilledButton.styleFrom(
          backgroundColor: cs.primary,
          foregroundColor: cs.onPrimary,
          disabledBackgroundColor: soleil.borderStrong,
          disabledForegroundColor: cs.onSurfaceVariant,
          minimumSize: const Size.fromHeight(48),
          shape: const StadiumBorder(),
          textStyle: SoleilTextStyles.button,
        ),
        child: isBusy
            ? SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: cs.onPrimary,
                ),
              )
            : Text(l10n.feedbackSubmit),
      ),
    );
  }
}

/// Shared "no character counter" builder used by the text fields.
Widget? _noCounter(
  BuildContext context, {
  required int currentLength,
  required bool isFocused,
  required int? maxLength,
}) =>
    null;
