/// Domain entity for the profile-completion endpoint payload. Plain
/// Dart (no freezed) matching the rest of the split-profile mobile
/// modules — keeps build_runner cycles short.
class ProfileCompletionSection {
  const ProfileCompletionSection({
    required this.key,
    required this.filled,
    required this.labelKey,
    required this.completionPath,
  });

  /// Stable machine identifier — backend never translates.
  final String key;

  /// Whether the section is filled in the user's current org state.
  final bool filled;

  /// Fully qualified i18n bucket the frontend resolves on render.
  /// e.g. `profile.completion.section.title`.
  final String labelKey;

  /// In-app URL to open when the user wants to fill the section.
  final String completionPath;

  factory ProfileCompletionSection.fromJson(Map<String, dynamic> json) {
    return ProfileCompletionSection(
      key: (json['key'] ?? '') as String,
      filled: (json['filled'] ?? false) as bool,
      labelKey: (json['label_key'] ?? '') as String,
      completionPath: (json['completion_path'] ?? '') as String,
    );
  }
}

class ProfileCompletionReport {
  const ProfileCompletionReport({
    required this.role,
    required this.persona,
    required this.percent,
    required this.totalSections,
    required this.filledSections,
    required this.sections,
  });

  /// Empty placeholder used while the first fetch is in flight.
  static const ProfileCompletionReport empty = ProfileCompletionReport(
    role: '',
    persona: '',
    percent: 0,
    totalSections: 0,
    filledSections: 0,
    sections: <ProfileCompletionSection>[],
  );

  final String role;
  final String persona;
  final int percent;
  final int totalSections;
  final int filledSections;
  final List<ProfileCompletionSection> sections;

  /// Convenience getter that surfaces the count of empty sections —
  /// drives the "X sections à compléter" pill on the bar.
  int get missingCount => totalSections - filledSections;

  /// True when every section is filled. Used to hide the bar on
  /// surfaces that opt out of celebrating the milestone.
  bool get isComplete => percent >= 100;

  factory ProfileCompletionReport.fromJson(Map<String, dynamic> json) {
    final raw = (json['sections'] ?? const <dynamic>[]) as List<dynamic>;
    final sections = raw
        .whereType<Map<String, dynamic>>()
        .map(ProfileCompletionSection.fromJson)
        .toList(growable: false);
    return ProfileCompletionReport(
      role: (json['role'] ?? '') as String,
      persona: (json['persona'] ?? '') as String,
      percent: (json['percent'] ?? 0) as int,
      totalSections: (json['total_sections'] ?? 0) as int,
      filledSections: (json['filled_sections'] ?? 0) as int,
      sections: sections,
    );
  }
}
