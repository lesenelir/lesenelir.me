// <html data-theme> only holds the theme in use, which can't tell a dark system from a dark choice. The choice goes on
// <html data-theme-choice> (absent for system), so the theme toggle shows it from the first paint.

// The browser toolbar's color in each theme, the page's background (--lz-background in src/styles/colors.css). The
// layout's viewport gives one <meta name="theme-color"> per system scheme; a chosen theme gives both its own color,
// so the toolbar follows the choice whichever scheme the system is in.
export const themeColors = { light: '#fcfcfc', dark: '#131413' }

// Runs in <head> before the first paint, after the theme-color tags. next-themes keeps the choice under 'theme', as a
// plain string.
export const themeChoiceScript = `try{var t=localStorage.getItem('theme');if(t==='light'||t==='dark'){document.documentElement.dataset.themeChoice=t;var c=${JSON.stringify(themeColors)}[t];document.querySelectorAll('meta[name="theme-color"]').forEach(function(m){m.content=c})}}catch(e){}`

export function setThemeChoice(theme: string | undefined) {
  const choice = theme === 'light' || theme === 'dark' ? theme : undefined
  if (choice) {
    document.documentElement.dataset.themeChoice = choice
  } else {
    delete document.documentElement.dataset.themeChoice
  }

  // For system, each tag gets back the color of its own scheme
  for (const meta of document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]')) {
    meta.content = themeColors[choice ?? (meta.media.includes('dark') ? 'dark' : 'light')]
  }
}
