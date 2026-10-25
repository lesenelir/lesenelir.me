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
        <p>Hi, I'm Lesenelir. I'm a software engineer, and I enjoy building things.</p>

        <p>
          I started writing code because it felt like building with LEGO, and I could get completely lost in it for
          hours. Now that AI writes a lot of the code, that feeling doesn't come around as often (only half joking).
          These days, I'm spending most of my time on <ExternalLink href={'https://hyper.ai'}>HyperAI</ExternalLink>, a
          platform for people working in AI, with open datasets, the latest papers, and GPU compute.
        </p>

        <p>
          This is my personal site, where I write about programming and whatever else I've been thinking about. I also
          share some of my photos and the songs I've been listening to.
        </p>

        <p>When I'm not coding, I'm usually taking photos, watching a movie, or playing Overwatch.</p>

        <p>
          My local time is <TimeClock />.
        </p>

        <p className={'flex flex-col gap-0.5'}>
          Find me on:
          <ExternalLink href={'https://github.com/lesenelir'} icon={<span className={'i-mingcute-github-line'} />}>
            GitHub
          </ExternalLink>
        </p>

        <p>
          Or email me:{' '}
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
