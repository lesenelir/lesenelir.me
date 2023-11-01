import Link from 'next/link'

interface IPosts {
  name: string
  href: string
  date: string
}

export const postData: IPosts[] = [
  {
    name: '地球村',
    href: '/posts/earth-village',
    date: 'Oct 9, 2023 ⋅ 6min'
  },
  {
    name: '地球村3',
    href: '',
    date: 'Jan 22, 2023 ⋅ 6min'
  },
  {
    name: '时间已经停止',
    href: '/posts/time-stop',
    date: 'Nov18, 2022 ⋅ 10min'
  }
]

export default function PostLists() {
  return (
    <>
      {
        postData.map((item) => (
          <Link
            key={item.name}
            href={item.href}
            className={`posts-li`}
          >
            <span>{item.name}</span>
            <span className={'text-sm text-navbarLight md:leading-7 dark:text-navbarDark'}>{item.date}</span>
          </Link>
        ))
      }
    </>
  )
}
