import { ListObjectsV2Command, S3Client } from '@aws-sdk/client-s3'

const s3Client = new S3Client({
  region: process.env.S3_REGION,
  endpoint: process.env.S3_ENDPOINT,
  credentials: {
    accessKeyId: process.env.S3_ACCESS_KEY_ID ?? '',
    secretAccessKey: process.env.S3_SECRET_ACCESS_KEY ?? ''
  },
  forcePathStyle: process.env.S3_FORCE_PATH_STYLE === 'true'
})

export const listSongs = async () => {
  try {
    const command = new ListObjectsV2Command({
      Bucket: process.env.S3_BUCKET_NAME,
      Prefix: 'songs/'
    })

    const response = await s3Client.send(command)
    return response.Contents || []
  } catch (error) {
    console.error('Error listing songs:', error)
    return []
  }
}

export const listPhotos = async () => {
  try {
    const command = new ListObjectsV2Command({
      Bucket: process.env.S3_BUCKET_NAME,
      Prefix: 'photos/'
    })

    const response = await s3Client.send(command)
    return response.Contents || []
  } catch (error) {
    console.error('Error listing photos:', error)
    return []
  }
}

export const getPublicUrl = (key: string) => {
  return `${process.env.S3_PUBLIC_URL}/${key}`
}
