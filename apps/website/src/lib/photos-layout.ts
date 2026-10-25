import type { LayoutMode } from '@/types'

// The photo gallery's layout lives on <html data-photos-layout>, so CSS lays the gallery out from the first paint
// instead of React switching to the stored layout after hydration
const STORAGE_KEY = 'photos-layout-mode'

// Runs in <head> before the first paint. The value is stored as JSON, the way jotai's atomWithStorage kept it.
export const photosLayoutScript = `try{var m=JSON.parse(localStorage.getItem('${STORAGE_KEY}'));if(m==='single'||m==='row'||m==='grid')document.documentElement.dataset.photosLayout=m}catch(e){}`

export function setPhotosLayout(mode: LayoutMode) {
  document.documentElement.dataset.photosLayout = mode
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(mode))
  } catch {
    // Storage can be unavailable (e.g. blocked site data); the layout still applies until the page reloads
  }
}
