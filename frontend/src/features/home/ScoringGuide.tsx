export function ScoringGuide() {
  return (
    <section className="scoring-guide" id="scoring-guide" tabIndex={-1} aria-labelledby="scoring-guide-heading">
      <div className="home-section-header">
        <p className="eyebrow">Understand the evidence</p>
        <h2 id="scoring-guide-heading">How scoring works</h2>
        <p>Higher difficulty means more recorded struggle, not a measurement of mastery.</p>
      </div>
      <div className="scoring-guide-grid">
        <article>
          <h3>Vocabulary difficulty</h3>
          <p>Each saved answer contributes a weight:</p>
          <dl className="scoring-weights">
            <div><dt>Again</dt><dd>1</dd></div>
            <div><dt>Hard</dt><dd>0.5</dd></div>
            <div><dt>Good</dt><dd>0</dd></div>
            <div><dt>Easy</dt><dd>0</dd></div>
          </dl>
          <p>Add the weights, divide by the answer count, multiply by 100, and round.</p>
          <p><strong>Suggested practice:</strong> at least 3 answers and a score of 50 or more.</p>
          <p>Fewer than 3 answers means limited evidence, even when the score is high.</p>
        </article>
        <article>
          <h3>Kanji difficulty</h3>
          <p>A word qualifies once it has at least 3 answers. Each qualifying written word contributes its vocabulary score equally to the kanji it contains.</p>
          <p>The kanji score is the rounded average. Less-reviewed words remain visible as early evidence but do not affect that average.</p>
          <p><strong>Suggested practice:</strong> at least 2 qualifying words and a score of 50 or more.</p>
          <p>One qualifying word means limited cross-word evidence. With none, the kanji is not scored yet.</p>
        </article>
      </div>
      <p className="scoring-guide-note">
        These are experimental suggestions. All recorded learning, review, and relearning answers currently have equal influence regardless of age.
        A difficult word is evidence to investigate, not proof that a specific kanji caused the difficulty.
      </p>
    </section>
  )
}
