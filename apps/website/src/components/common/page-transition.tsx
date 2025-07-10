'use client'

import { AnimatePresence, motion } from 'motion/react'
import { usePathname } from 'next/navigation'
import type { ReactNode } from 'react'

interface PageTransitionProps {
  children: ReactNode
}

const pageVariants = {
  initial: {
    filter: 'blur(1.5px)'
  },
  in: {
    filter: 'blur(0px)'
  },
  out: {
    filter: 'blur(1.5px)'
  }
}

const pageTransition = {
  type: 'tween' as const,
  ease: 'easeIn' as const,
  duration: 0.3
}

export function PageTransition({ children }: PageTransitionProps) {
  const pathname = usePathname()

  return (
    <AnimatePresence mode={'wait'} initial={false}>
      <motion.div
        key={pathname}
        initial={'initial'}
        animate={'in'}
        exit={'out'}
        variants={pageVariants}
        transition={pageTransition}
      >
        {children}
      </motion.div>
    </AnimatePresence>
  )
}
