import TimeClock from '@/components/common/time-clock'
import { cn } from '@/lib/utils'

export default function Page() {
  return (
    <>
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
              'inline-flex w-fit items-center gap-1',
              'relative after:absolute after:bottom-0.5 after:left-0 after:h-px after:w-full',
              'after:bg-text-primary/50 hover:after:bg-text-primary after:transition-colors after:duration-200'
            )}
          >
            <span className={'i-mingcute-github-line'} />
            Github
          </a>
        </p>

        <p>
          Mail me:{' '}
          <a
            href={'mailto:hi@lesenelir.me'}
            className={cn(
              'inline-flex w-fit items-center gap-1',
              'relative after:absolute after:bottom-0.5 after:left-0 after:h-px after:w-full',
              'after:bg-text-primary/50 hover:after:bg-text-primary after:transition-colors after:duration-200'
            )}
          >
            hi@lesenelir.me
          </a>
        </p>
      </div>
    </>
  )
}
