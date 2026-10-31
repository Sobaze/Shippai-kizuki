import { useEffect, useRef } from 'react'
import { KanjiStudyContent } from './KanjiStudyContent'
import type { KanjiStats } from './types'
import './kanji-detail.css'

type KanjiDetailProps = {
  entry: KanjiStats
  position: number
  total: number
  onPrevious: () => void
  onNext: () => void
  onClose: () => void
}

export function KanjiDetail({ entry, position, total, onPrevious, onNext, onClose }: KanjiDetailProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const headingRef = useRef<HTMLHeadingElement>(null)
  const hasPrevious = position > 1
  const hasNext = position < total

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return

    const previousOverflow = document.body.style.overflow
    dialog.showModal()
    headingRef.current?.focus()
    document.body.style.overflow = 'hidden'

    return () => {
      dialog.close()
      document.body.style.overflow = previousOverflow
    }
  }, [])

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return
    dialog.scrollTo(0, 0)

    // Keep keyboard navigation available if the focused arrow became disabled.
    const active = document.activeElement
    if (!dialog.contains(active) || (active instanceof HTMLButtonElement && active.disabled)) {
      headingRef.current?.focus({ preventScroll: true })
    }
  }, [entry.kanji])

  return (
    <dialog
      className="kanji-detail"
      ref={dialogRef}
      aria-labelledby="kanji-detail-heading"
      onCancel={(event) => {
        event.preventDefault()
        onClose()
      }}
      onKeyDown={(event) => {
        if (event.defaultPrevented || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return
        const target = event.target
        if (target instanceof HTMLElement && (target.isContentEditable || target.closest('input, textarea, select'))) return
        if (event.key === 'ArrowLeft') {
          event.preventDefault()
          if (hasPrevious) onPrevious()
        } else if (event.key === 'ArrowRight') {
          event.preventDefault()
          if (hasNext) onNext()
        }
      }}
      onClick={(event) => {
        if (event.target !== event.currentTarget) return
        const bounds = event.currentTarget.getBoundingClientRect()
        if (event.clientX < bounds.left || event.clientX > bounds.right || event.clientY < bounds.top || event.clientY > bounds.bottom) {
          onClose()
        }
      }}
    >
      <header className="kanji-modal-toolbar">
        <div className="kanji-modal-title">
          <h2 id="kanji-detail-heading" ref={headingRef} tabIndex={-1}>Study <span lang="ja">{entry.kanji}</span></h2>
          <p className="kanji-detail-position" role="status">{position} of {total} on this page</p>
        </div>
        <div className="kanji-modal-actions">
          <div className="kanji-detail-navigation" role="group" aria-label="Browse kanji on this page">
            <button type="button" className="kanji-detail-arrow" onClick={onPrevious} disabled={!hasPrevious} aria-label="Previous kanji" aria-keyshortcuts="ArrowLeft" title="Previous kanji (Left arrow)">
              ← <span className="kanji-direction-label">Previous</span>
            </button>
            <button type="button" className="kanji-detail-arrow" onClick={onNext} disabled={!hasNext} aria-label="Next kanji" aria-keyshortcuts="ArrowRight" title="Next kanji (Right arrow)">
              <span className="kanji-direction-label">Next</span> →
            </button>
          </div>
          <button type="button" className="kanji-close" onClick={onClose} aria-label="Close kanji details">Close ×</button>
        </div>
      </header>
      {/* Keep the dialog open while resetting lookup and practice state per kanji. */}
      <KanjiStudyContent key={entry.kanji} entry={entry} />
    </dialog>
  )
}
