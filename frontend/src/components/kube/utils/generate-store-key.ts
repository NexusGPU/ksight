import type { KubeObjectStoreOptions } from '../kube-api/kube-object-store'

export function generateStoreKey(storeName: string, opts?: KubeObjectStoreOptions): string {
  if (!opts || Object.keys(opts).length === 0) {
    return storeName
  }

  // create a stable, remove undefined value configuration object
  const cleanOpts: Record<string, any> = {}

  Object.keys(opts).sort().forEach((key) => {
    const value = (opts as any)[key]
    if (value !== undefined && value !== null) {
      cleanOpts[key] = value
    }
  })

  if (Object.keys(cleanOpts).length === 0) {
    return storeName
  }

  // generate stable hash
  const optsString = JSON.stringify(cleanOpts)

  // use simple hash
  let hash = 0
  for (let i = 0; i < optsString.length; i++) {
    const char = optsString.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash |= 0 // convert to 32-bit signed integer
  }

  const shortHash = Math.abs(hash).toString(36).substring(0, 6)
  return `${storeName}-${shortHash}`
}
