import 'package:flutter/material.dart';

import '../../../../core/theme/app_theme.dart';
import '../../../../l10n/app_localizations.dart';
import '../../domain/entities/feedback_attachment.dart';
import '../../domain/entities/feedback_attachment_kind.dart';

/// Attachments block of the feedback sheet.
///
/// Anonymous reporters see a disabled state with a hint ("Sign in to
/// attach…") — text-only submissions remain possible. Logged-in
/// reporters get the image + video pick buttons and a chip list of the
/// uploaded files. Attachment gating is decided by the parent via
/// [canAttach]; this widget only renders accordingly.
class FeedbackAttachmentsSection extends StatelessWidget {
  const FeedbackAttachmentsSection({
    super.key,
    required this.canAttach,
    required this.attachments,
    required this.isUploading,
    required this.onPickImage,
    required this.onPickVideo,
    required this.onRemove,
  });

  final bool canAttach;
  final List<FeedbackAttachment> attachments;
  final bool isUploading;
  final VoidCallback onPickImage;
  final VoidCallback onPickVideo;
  final ValueChanged<FeedbackAttachment> onRemove;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final cs = Theme.of(context).colorScheme;
    final soleil = Theme.of(context).extension<AppColors>()!;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.feedbackAttachmentsLabel,
          style: SoleilTextStyles.bodyEmphasis.copyWith(
            fontSize: 13,
            color: cs.onSurfaceVariant,
          ),
        ),
        const SizedBox(height: 8),
        if (!canAttach)
          _LoginHint(text: l10n.feedbackAttachmentsLoginHint)
        else ...[
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: isUploading ? null : onPickImage,
                  icon: const Icon(Icons.image_outlined, size: 18),
                  label: Text(l10n.feedbackAddImage),
                  style: OutlinedButton.styleFrom(
                    minimumSize: const Size.fromHeight(44),
                    shape: const StadiumBorder(),
                    side: BorderSide(color: soleil.borderStrong),
                    foregroundColor: cs.onSurface,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: isUploading ? null : onPickVideo,
                  icon: const Icon(Icons.videocam_outlined, size: 18),
                  label: Text(l10n.feedbackAddVideo),
                  style: OutlinedButton.styleFrom(
                    minimumSize: const Size.fromHeight(44),
                    shape: const StadiumBorder(),
                    side: BorderSide(color: soleil.borderStrong),
                    foregroundColor: cs.onSurface,
                  ),
                ),
              ),
            ],
          ),
          if (isUploading)
            Padding(
              padding: const EdgeInsets.only(top: 12),
              child: Row(
                children: [
                  SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: cs.primary,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Text(
                    l10n.feedbackUploading,
                    style: SoleilTextStyles.body.copyWith(
                      color: cs.onSurfaceVariant,
                    ),
                  ),
                ],
              ),
            ),
          if (attachments.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 12),
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  for (final attachment in attachments)
                    _AttachmentChip(
                      attachment: attachment,
                      onRemove: () => onRemove(attachment),
                      removeLabel: l10n.feedbackRemoveAttachment,
                    ),
                ],
              ),
            ),
        ],
      ],
    );
  }
}

class _LoginHint extends StatelessWidget {
  const _LoginHint({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final soleil = Theme.of(context).extension<AppColors>()!;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: soleil.muted,
        borderRadius: BorderRadius.circular(AppTheme.radiusMd),
        border: Border.all(color: soleil.border),
      ),
      child: Row(
        children: [
          Icon(Icons.lock_outline, size: 18, color: cs.onSurfaceVariant),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              text,
              style: SoleilTextStyles.body.copyWith(
                fontSize: 13,
                color: cs.onSurfaceVariant,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AttachmentChip extends StatelessWidget {
  const _AttachmentChip({
    required this.attachment,
    required this.onRemove,
    required this.removeLabel,
  });

  final FeedbackAttachment attachment;
  final VoidCallback onRemove;
  final String removeLabel;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final soleil = Theme.of(context).extension<AppColors>()!;
    final isImage = attachment.kind == FeedbackAttachmentKind.image;
    return Container(
      padding: const EdgeInsets.fromLTRB(10, 6, 6, 6),
      decoration: BoxDecoration(
        color: soleil.accentSoft,
        borderRadius: BorderRadius.circular(AppTheme.radiusFull),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            isImage ? Icons.image_outlined : Icons.videocam_outlined,
            size: 16,
            color: cs.primary,
          ),
          const SizedBox(width: 6),
          Text(
            _formatSize(attachment.sizeBytes),
            style: SoleilTextStyles.mono.copyWith(
              fontSize: 11,
              color: soleil.primaryDeep,
            ),
          ),
          const SizedBox(width: 2),
          Semantics(
            button: true,
            label: removeLabel,
            child: InkWell(
              onTap: onRemove,
              borderRadius: BorderRadius.circular(AppTheme.radiusFull),
              child: Padding(
                padding: const EdgeInsets.all(2),
                child: Icon(Icons.close, size: 16, color: cs.primary),
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Compact human-readable size for the chip (KB / MB).
  String _formatSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).round()} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(1)} MB';
  }
}
