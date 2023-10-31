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
        otherData.map(item => (
          <a
            key={item.name}
            href={item.href}
            target='_blank'
            className={'flex md:flex-row max-md:flex-col justify-between py-2 border-b border-gray-300 antialiased cursor-pointer no-underline dark:border-gray-500'}
          >
            <span>{item.name}</span>
            <span className={'text-sm text-navbarLight md:leading-7 dark:text-navbarDark'}>{item.description}</span>
          </a>
        ))
      }
    </>
  )
}
