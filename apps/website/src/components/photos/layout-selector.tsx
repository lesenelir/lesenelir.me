import type { LayoutMode } from '@/types'

import { setPhotosLayout } from '@/lib/photos-layout'
import { cn } from '@/lib/utils'

// The active layout is styled by CSS from <html data-photos-layout>, so it is right from the first paint.
// Each button only takes the hover color in the other two layouts.
const layouts: { mode: LayoutMode; title: string; icon: string; className: string }[] = [
  {
    mode: 'single',
    title: 'Single column layout',
    icon: 'i-mingcute-layout-bottom-fill',
    className:
      'photos-single:text-text-primary photos-row:hover:text-text-foreground/80 photos-grid:hover:text-text-foreground/80',
  },
  {
    mode: 'row',
    title: 'Single row layout',
    icon: 'i-mingcute-layout-right-fill',
    className:
      'photos-row:text-text-primary photos-single:hover:text-text-foreground/80 photos-grid:hover:text-text-foreground/80',
  },
  {
    mode: 'grid',
    title: 'Grid layout',
    icon: 'i-mingcute-layout-grid-fill',
    className:
      'photos-grid:text-text-primary photos-single:hover:text-text-foreground/80 photos-row:hover:text-text-foreground/80',
  },
]

export function LayoutSelector() {
  return (
    <div className={'mb-8 inline-flex flex-wrap items-center justify-start gap-4 text-text-foreground/55'}>
      {layouts.map(layout => (
        <button
          key={layout.mode}
          type='button'
          onClick={() => setPhotosLayout(layout.mode)}
          className={cn(
            'flex size-6 cursor-pointer items-center justify-center transition-colors duration-200',
            layout.className
          )}
          title={layout.title}
        >
          <span className={cn(layout.icon, 'size-4.5')} />
        </button>
      ))}
    </div>
  )
}
