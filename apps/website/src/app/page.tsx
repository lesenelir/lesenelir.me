import type { Metadata } from 'next'

import { cn } from '@/lib/utils'

import ExternalLink from '@/components/common/external-link'
import TimeClock from '@/components/common/time-clock'

export const metadata: Metadata = {
  alternates: { canonical: '/' },
}

export default function Page() {
  return (
    <div className={'max-w-2xl'}>
      <h1 className={'mb-11 font-comic text-text-primary'}>Lesenelir Zhou</h1>

      <div className={'space-y-8 text-text-primary/85'}>
        <p>Hello, I'm lesenelir, a software engineer who loves building anything.</p>

        <p>
          Beijing is where I currently work and live. Local time here: <TimeClock />. Prior to Beijing, my life journey
          has taken me to live in Quzhou, Wuhan, and Guangzhou.
        </p>

        <p>Besides coding, I also enjoy sleeping, movies, photography, and rock music.</p>

        <p className={'flex flex-col gap-0.5'}>
          Find me on:
          <ExternalLink href={'https://github.com/lesenelir'} icon={<span className={'i-mingcute-github-line'} />}>
            GitHub
          </ExternalLink>
        </p>

        <p>
          Or mail me:{' '}
          <a
            href={'mailto:hi@lesenelir.me'}
            className={cn(
              'text-link',
              'underline decoration-link/85 decoration-1 underline-offset-2 hover:decoration-link',
              'transition-colors duration-200'
            )}
          >
            hi@lesenelir.me
          </a>
        </p>

        <hr className={'my-12 w-1/3 border-t border-dividing'} />

        <ExternalLink href={'https://github.com/lesenelir/lesenelir.me/tree/master/apps/website'}>
          View source
        </ExternalLink>
      </div>
    </div>
  )
}
