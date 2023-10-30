import ThemeChange from '@/components/theme/ThemeChange'

export default function Footer() {
  return (
    <div className={'flex flex-row justify-between text-sm max-md:text-xs'}>
      <p className={'text-navbarLight dark:text-navbarDark'}>
        <a
          href="https://creativecommons.org/licenses/by-nc-sa/4.0/"
          rel='noreferrer'
          target='_blank'
          className={'footer-left'}>
          CC BY-NC-SA 4.0
        </a>
        <span className={'opacity-80'}>{' '} 2023-PRESENT © Lesenelir</span>
      </p>

      <div className={'flex flex-row gap-4 text-navbarLight dark:text-navbarDark'}>
        <div className={'footer-right'}>
          <ThemeChange/>
        </div>
        <div className={'footer-right'}>
          RSS
        </div>
      </div>
    </div>
  )
}
