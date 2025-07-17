import { useAtom } from 'jotai'

import { layoutModeAtom } from '@/atoms/photos'
import { cn } from '@/lib/utils'

export function LayoutSelector() {
  const [layoutMode, setLayoutMode] = useAtom(layoutModeAtom)

  return (
    <div
      className={
        'text-text-foreground/55 mb-8 inline-flex flex-wrap items-center justify-start gap-4'
      }
    >
      <button
        onClick={() => setLayoutMode('single')}
        className={cn(
          'flex size-6 cursor-pointer items-center justify-center transition-colors duration-200',
          layoutMode === 'single' ? 'text-text-primary' : 'hover:text-text-foreground/80'
        )}
        title={'Single column layout'}
      >
        <span className={'i-mingcute-layout-bottom-fill size-4.5'} />
      </button>

      <button
        onClick={() => setLayoutMode('row')}
        className={cn(
          'flex size-6 cursor-pointer items-center justify-center transition-colors duration-200',
          layoutMode === 'row' ? 'text-text-primary' : 'hover:text-text-foreground/80'
        )}
        title={'Single row layout'}
      >
        <span className={'i-mingcute-layout-right-fill size-4.5'} />
      </button>

      <button
        onClick={() => setLayoutMode('grid')}
        className={cn(
          'flex size-6 cursor-pointer items-center justify-center transition-colors duration-200',
          layoutMode === 'grid' ? 'text-text-primary' : 'hover:text-text-foreground/80'
        )}
        title={'Grid layout'}
      >
        <span className={'i-mingcute-layout-grid-fill size4.5'} />
      </button>
    </div>
  )
}
