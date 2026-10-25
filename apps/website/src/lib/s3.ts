// The SDK must never reach the browser: importing this module from a client component fails the build
import 'server-only'

import { GetObjectCommand, ListObjectsV2Command, S3Client } from '@aws-sdk/client-s3'

const s3Client = new S3Client({
  region: process.env.S3_REGION,
  endpoint: process.env.S3_ENDPOINT,
  credentials: {
    accessKeyId: process.env.S3_ACCESS_KEY_ID ?? '',
    secretAccessKey: process.env.S3_SECRET_ACCESS_KEY ?? '',
  },
  forcePathStyle: process.env.S3_FORCE_PATH_STYLE === 'true',
})

// Errors are thrown, not swallowed: a failed build stops instead of shipping an empty list, and a failed
// revalidation keeps serving the last good page
export const listSongs = async () => {
  const command = new ListObjectsV2Command({
    Bucket: process.env.S3_BUCKET_NAME,
    Prefix: 'songs/',
  })

  const response = await s3Client.send(command)
  return response.Contents || []
}

export const listPhotos = async () => {
  const command = new ListObjectsV2Command({
    Bucket: process.env.S3_BUCKET_NAME,
    Prefix: 'photos/',
  })

  const response = await s3Client.send(command)
  return response.Contents || []
}

// Downloads only the first bytes of an object, e.g. the header of a photo
export const getObjectStart = async (key: string, bytes: number) => {
  const command = new GetObjectCommand({
    Bucket: process.env.S3_BUCKET_NAME,
    Key: key,
    Range: `bytes=0-${bytes - 1}`,
  })

  const response = await s3Client.send(command)
  return response.Body ? response.Body.transformToByteArray() : new Uint8Array()
}

export const getPublicUrl = (key: string) => {
  return `${process.env.S3_PUBLIC_URL}/${key}`
}
