export interface IProject {
  name: string
  href: string
  description: string
}

export const projectData: IProject[] = [
  {
    name: 'AILab',
    href: 'https://ailab.uz/',
    description: 'Production Practices of GPT'
  },
  {
    name: 'NextGPT',
    href: 'https://github.com/lesenelir/next-gpt',
    description: 'Implementing ChatGPT with Next.js'
  },
  {
    name: 'Type',
    href: 'https://github.com/lesenelir/type',
    description: 'Replicating OpenAI\'s Animation'
  },
  {
    name: 'Lesenelir.me',
    href: 'https://github.com/lesenelir/lesenelir.me',
    description: 'Recreating AntFu Blog with Next.js'
  },
  {
    name: 'React Admin',
    href: 'https://github.com/lesenelir/meta-react-admin',
    description: 'A full stack project with React and Koa2'
  },
  {
    name: 'LearnWeb3',
    href: 'https://github.com/lesenelir/meta-learnweb3-dao-projects',
    description: 'Some example practices from LearnWeb3 DAO'
  },
  {
    name: 'GmWeb3',
    href: 'https://github.com/lesenelir/gmweb3',
    description: 'A cli tool for generating DApp front-end templates'
  },
  {
    name: 'Conf',
    href: 'https://github.com/lesenelir/conf',
    description: 'A monorepo for configuration options'
  },
  {
    name: 'Mini React',
    href: 'https://github.com/lesenelir/mini-react',
    description: 'Build your own react tutorial'
  },
  {
    name: 'Suspense',
    href: 'https://github.com/lesenelir/suspense',
    description: 'Data fetching with React Suspense'
  },
  {
    name: 'Flex Baidu',
    href: 'https://github.com/lesenelir/meta-baidu',
    description: 'Implementing Baidu\'s Page with CSS Flex'
  },
  {
    name: 'Cats Classifier',
    href: 'https://github.com/lesenelir/Django-Cats-Classifier',
    description: 'Classification of four cats-kinds with CNN and Django'
  }
]

export default function ProjectLists() {

  return (
    <>
      {
        projectData.map(item => (
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
