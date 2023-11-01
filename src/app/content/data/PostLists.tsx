import Link from 'next/link'

interface IPosts {
  name: string
  href: string
  file: string
  date: string
}

export const postData: IPosts[] = [
  {
    name: '地球村',
    href: '/posts/earth-village',
    file: '2023-10-09-earth-village',
    date: 'Oct 9, 2023 ⋅ 06 min'
  },
  {
    name: '我所理解的递归',
    href: '/posts/recursion',
    file: '2022-12-22-recursion',
    date: 'Dec 22, 2022 ⋅ 22 min'
  },
  {
    name: '时间已经停止',
    href: '/posts/time-stop',
    file: '2022-11-18-time-stop',
    date: 'Nov 18, 2022 ⋅ 10 min'
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
