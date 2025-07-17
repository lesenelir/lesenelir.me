import { useAtomValue, useSetAtom } from 'jotai'
import { AnimatePresence, motion } from 'motion/react'
import Image from 'next/image'
import { createPortal } from 'react-dom'
import { useHotkeys } from 'react-hotkeys-hook'

import { closePhotoModalAtom, isPhotoModalOpenAtom, selectedPhotoAtom } from '@/atoms/photos'

const backdropVariants = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 }
}

const imageContainerVariants = {
  initial: { scale: 0.8, opacity: 0 },
  animate: { scale: 1, opacity: 1 },
  exit: { scale: 0.8, opacity: 0 }
}

const transition = {
  type: 'tween' as const,
  ease: 'easeOut' as const,
  duration: 0.3
}

export function PhotoModal() {
  const isOpen = useAtomValue(isPhotoModalOpenAtom)
  const selectedPhoto = useAtomValue(selectedPhotoAtom)
  const closeModal = useSetAtom(closePhotoModalAtom)

  // Keyboard navigation
  useHotkeys('esc', closeModal, { enabled: isOpen })

  return createPortal(
    <AnimatePresence mode="wait">
      {isOpen && selectedPhoto && (
        <motion.div
          onClick={closeModal}
          className={
            'fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm'
          }
          variants={backdropVariants}
          initial="initial"
          animate="animate"
          exit="exit"
          transition={transition}
        >
          {/* Photo content */}
          <motion.div
            className={'relative flex size-full items-center justify-center'}
            variants={imageContainerVariants}
            initial="initial"
            animate="animate"
            exit="exit"
            transition={transition}
          >
            <Image
              src={selectedPhoto.src}
              alt={selectedPhoto.alt}
              width={1200}
              height={800}
              className={'h-auto max-h-full w-full max-w-full object-contain lg:h-full lg:w-auto'}
              sizes={'(max-width: 1024px) 100vw, 100vh'}
              priority
            />
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>,
    document.body
  )
}
