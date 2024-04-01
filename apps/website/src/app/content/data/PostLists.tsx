import Link from 'next/link'

interface IPosts {
  name: string
  href: string
  file: string
  date: string
}

export const postData: IPosts[] = [
  {
    name: 'Jotai 理念与实践',
    href: '/posts/jotai',
    file: '2024-04-01-Jotai',
    date: 'Apr 01, 2024 ⋅ 20 min'
  },
  {
    name: '地球村',
    href: '/posts/earth-village',
    file: '2023-10-09-earth-village',
    date: 'Oct 09, 2023 ⋅ 08 min'
  },
  {
    name: '强制类型语言',
    href: '/posts/typescript',
    file: '2023-08-22-typescript',
    date: 'Aug 22, 2023 ⋅ 15 min'
  },
  {
    name: '我所理解的递归',
    href: '/posts/recursion',
    file: '2022-12-22-recursion',
    date: 'Dec 22, 2022 ⋅ 22 min'
  },
  {
    name: '时间已经停止，剩下就是慢慢变老',
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
            <span className={'posts-name'}>{item.name}</span>
            <span className={'posts-date text-sm text-navbarLight md:leading-7 dark:text-navbarDark'}>{item.date}</span>
          </Link>
        ))
      }
    </>
  )
}
