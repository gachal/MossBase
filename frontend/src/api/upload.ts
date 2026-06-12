import client from './client'

export async function uploadFile(file: File, type: 'avatar' | 'space-cover'): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)
  const { data } = await client.post(`/upload/${type}`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return data.data.url
}
