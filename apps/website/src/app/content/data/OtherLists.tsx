import { clsx } from 'clsx'

export interface IOther {
  name: string
  href: string
  description: string
}

export const otherData: IOther[] = [
  {
    name: 'Leetcode',
    href: 'https://github.com/lesenelir/meta-leetcode-js',
    description: 'Some JavaScript Solutions for LeetCode Problems'
  },
]

export default function OtherLists() {

  return (
    <>
      {
        otherData.map((item, index) => (
          <a
            key={item.name}
            href={item.href}
            target='_blank'
            className={clsx(
              'projects-li',
              'font-normal text-[#374151] hover:text-black',
              'dark:text-[#D1D5DB] dark:hover:text-white',
              index === otherData.length - 1 ? '' : 'border-b border-gray-300 dark:border-gray-500'
            )}
          >
            <span>{item.name}</span>
            <span className={'text-sm text-navbarLight md:leading-7 dark:text-navbarDark'}>{item.description}</span>
          </a>
        ))
      }
    </>
  )
}
