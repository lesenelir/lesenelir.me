export default function Footer() {
  return (
    <div className={'flex flex-row justify-between'}>
      <p className={'text-sm text-navbarLight opacity-80'}>
        <a
          href="https://creativecommons.org/licenses/by-nc-sa/4.0/"
          rel='noreferrer'
          target='_blank'
          className={'footer-left'}>
          CC BY-NC-SA 4.0
        </a>
        {' '} 2023-PRESENT © Lesenelir Zhou
      </p>

      <div className={'flex flex-row gap-4 text-sm text-navbarLight cursor-pointer'}>
        <p className={'footer-right'}>
          Theme
        </p>
        <p className={'footer-right'}>
          RSS
        </p>
      </div>
    </div>
  )
}
