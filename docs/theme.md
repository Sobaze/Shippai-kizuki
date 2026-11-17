# Lantern-alley interface theme

The interface takes its cues from a late-night izakaya lane: charcoal/indigo
surroundings, warm lantern light, vermilion details, wood-toned borders, and
cream paper surfaces. Japanese display typography and segmented navigation give
it a storefront/menu feel while the learning controls remain straightforward.

## Where to adjust it

- `frontend/src/index.css`: semantic colour roles, font stacks, shared radii,
  background texture, focus styles, and the `.paper-surface` palette.
- `frontend/src/components/app-navigation.css`: lantern brand mark and bilingual
  navigation; mobile uses two rows with English and Japanese labels.
- `frontend/src/features/home/home.css`: paper-and-illustration Home hero.
- `frontend/src/features/quiz/quiz.css`: wide study cards, paper question sheets,
  targets, choices, and result feedback.
- Other feature styles use the same roles for cards, tables, tags, and dialogs.

The primary semantic roles are `--text`, `--muted`, `--accent`, `--secondary`,
`--primary`, `--positive`, `--negative`, and `--warning`. Corresponding backgrounds
and borders keep controls consistent. The paper palette overrides the roles
locally so highlighted targets, form controls, and focus rings use ink-safe
colours rather than light-on-dark accent colours.

## Local artwork and typography

`frontend/public/izakaya-alley.svg` is original local artwork for this project.
`lantern-mark.svg` is the matching favicon. The Home image is decorative and is
hidden from assistive technology. Fonts use installed Japanese/system fallbacks:
display headings and isolated kanji glyphs prefer Mincho/serif faces; controls,
sentences, and answer text use the body font stack. The theme needs no remote
images, font requests, or libraries.

## Study readability and responsive layout

- Quiz cards retain the full content width and larger padding.
- Sentence text and Japanese choices retain `1.5–1.875rem` sizing (24–30px at the
  default root size). Dictionary reading/furigana sizes are preserved.
- Long-form comprehension passages use 22–26px text and 20–24px choices, with
  paragraph breaks and a full-width reading/results layout.
- Practice-type controls share one row on larger screens and stack on mobile.
- The Home hero stacks on tablet/mobile; the alley illustration scales locally.
- Status labels retain text as well as colour. Keyboard focus rings work on
  both paper and dark surfaces, and reduced-motion preferences are respected.

Browser checks cover Home, populated Insights, dictionary dialogs, vocabulary,
quiz setup, all twelve questions of Exercise 4, and results at 1440px, 768px, and
320px. Checks include document overflow, preserved Japanese font sizes, loaded
artwork, and at least 4.5:1 contrast for the sampled body, target, and primary
action text/background pairs on paper and dark surfaces.
