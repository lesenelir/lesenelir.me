import { useEffect, useState } from 'react'

import type { LayoutMode } from '@/types'

import { getPhotosLayout, setPhotosLayout } from '@/lib/photos-layout'
import { cn } from '@/lib/utils'

// The active layout is styled by CSS from <html data-photos-layout>, so it is right from the first paint.
// Each button only takes the hover color in the other two layouts.
const layouts: { mode: LayoutMode; title: string; icon: string; className: string }[] = [
  {
    mode: 'single',
    title: 'Single column layout',
    icon: 'i-mingcute-layout-bottom-fill',
    className:
      'photos-single:text-text-primary photos-row:hover:text-text-foreground photos-grid:hover:text-text-foreground',
  },
  {
    mode: 'row',
    title: 'Single row layout',
    icon: 'i-mingcute-layout-right-fill',
    className:
      'photos-row:text-text-primary photos-single:hover:text-text-foreground photos-grid:hover:text-text-foreground',
  },
  {
    mode: 'grid',
    title: 'Grid layout',
    icon: 'i-mingcute-layout-grid-fill',
    className:
      'photos-grid:text-text-primary photos-single:hover:text-text-foreground photos-row:hover:text-text-foreground',
  },
]

export function LayoutSelector() {
  // Only known in the browser, so aria-pressed is left out of the server HTML and added once mounted
  const [activeMode, setActiveMode] = useState<LayoutMode | null>(null)

  useEffect(() => {
    setActiveMode(getPhotosLayout())
  }, [])

  return (
    // /80 is the lightest these icons can be and still stand out 3:1 from the background, in both themes
    <fieldset
      aria-label={'Photo layout'}
      className={'mb-8 inline-flex flex-wrap items-center justify-start gap-4 text-text-foreground/80'}
    >
      {layouts.map(layout => (
        <button
          key={layout.mode}
          type='button'
          onClick={() => {
            setPhotosLayout(layout.mode)
            setActiveMode(layout.mode)
          }}
          className={cn(
            'flex size-6 cursor-pointer items-center justify-center transition-colors duration-200',
            layout.className
          )}
          title={layout.title}
          aria-label={layout.title}
          aria-pressed={activeMode === null ? undefined : activeMode === layout.mode}
        >
          <span className={cn(layout.icon, 'size-4.5')} />
        </button>
      ))}
    </fieldset>
  )
}
