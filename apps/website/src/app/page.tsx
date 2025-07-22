import ExternalLink from '@/components/common/external-link'
import TimeClock from '@/components/common/time-clock'

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
          <ExternalLink
            href={'https://github.com/lesenelir'}
            icon={<span className={'i-mingcute-github-line'} />}
          >
            Github
          </ExternalLink>
        </p>

        <p>Or mail me: hi@lesenelir.me</p>

        <hr className={'border-dividing my-12 w-1/3 border-t'} />

        <ExternalLink href={'https://github.com/lesenelir/lesenelir.me/tree/master/apps/website'}>
          View source
        </ExternalLink>
      </div>
    </div>
  )
}
