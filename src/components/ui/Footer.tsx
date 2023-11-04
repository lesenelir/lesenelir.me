import ThemeChange from '@/components/theme/ThemeChange'

export default function Footer() {
  return (
    <>
      <div className={'flex flex-row gap-2 mb-20 mt-2'}>
        <a href={'https://github.com/lesenelir'} target={'_blank'} className={'footer-a'}>Github</a>
        <a href={'https://twitter.com/lesenelir'} target={'_blank'} className={'footer-a'}>Twitter</a>
        <a href={'https://mirror.xyz/lesenelir.eth'} target={'_blank'} className={'footer-a'}>Mirror</a>
        <a href={'mailto:hi@lesenelir.me'} target={'_blank'} className={'footer-a'}>hi@lesenelir.me</a>
      </div>

      <div className={'md:flex md:flex-row justify-between text-sm'}>
        {/* Theme and RSS */}
        <div className={'text-navbarLight flex flex-row gap-4 dark:text-navbarDark max-md:mb-4'}>
          <div className={'footer-a'}>
            <ThemeChange/>
          </div>
          <div className={'footer-a'}>
            <a href={'/feed.xml'} target={'_blank'}>
              RSS
            </a>
          </div>
        </div>

        {/* License */}
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
      </div>
    </>
  )
}
