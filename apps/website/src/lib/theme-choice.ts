// <html data-theme> only holds the theme in use, which can't tell a dark system from a dark choice. The choice goes on
// <html data-theme-choice> (absent for system), so the theme toggle shows it from the first paint.

// Runs in <head> before the first paint. next-themes keeps the choice under 'theme', as a plain string.
export const themeChoiceScript = `try{var t=localStorage.getItem('theme');if(t==='light'||t==='dark')document.documentElement.dataset.themeChoice=t}catch(e){}`

export function setThemeChoice(theme: string | undefined) {
  if (theme === 'light' || theme === 'dark') {
    document.documentElement.dataset.themeChoice = theme
  } else {
    delete document.documentElement.dataset.themeChoice
  }
}
