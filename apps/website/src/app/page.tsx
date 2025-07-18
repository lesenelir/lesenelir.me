import TimeClock from '@/components/common/time-clock'
import { cn } from '@/lib/utils'

export default function Page() {
  return (
    <div className={'max-w-2xl'}>
      <h3 className={'font-comic text-text-primary mb-11'}>Lesenelir Zhou</h3>

      <div className={'text-text-primary/85 space-y-8'}>
        <p>Hello, I'm lesenelir, a software engineer who loves building anything.</p>

        <p>
          Beijing is where I currently work and live. Local time here: <TimeClock />. Prior to
          Beijing, My life journey has taken me to live in Quzhou, Wuhan, and Guangzhou.
        </p>

        <p>Besides coding, I also enjoy sleeping, movies, photography, and rock music.</p>

        <p className={'flex flex-col gap-0.5'}>
          Find me on:
          <a
            target={'_blank'}
            href={'https://github.com/lesenelir'}
            rel={'noopener noreferrer'}
            className={cn(
              'text-link inline-flex w-fit items-center gap-1',
              'decoration-link/85 hover:decoration-link underline decoration-1 underline-offset-2',
              'transition-colors duration-200'
            )}
          >
            <span className={'i-mingcute-github-line'} />
            Github
            <span
              className={
                'i-mingcute-right-line relative -top-1.5 -ml-1.5 inline-block size-3.5 -rotate-45'
              }
            />
          </a>
        </p>

        <p>Or mail me: hi@lesenelir.me</p>
      </div>
    </div>
  )
}
