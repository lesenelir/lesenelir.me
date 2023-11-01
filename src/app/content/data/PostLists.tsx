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
    date: 'Oct 9, 2023 ⋅ 06min'
  },
  {
    name: '我所理解的递归',
    href: '/posts/recursion',
    date: 'Dec 22, 2022 ⋅ 22min'
  },
  {
    name: '时间已经停止',
    href: '/posts/time-stop',
    date: 'Nov 18, 2022 ⋅ 10min'
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
