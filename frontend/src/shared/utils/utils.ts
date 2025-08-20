import type { Updater } from '@tanstack/vue-table'
import type { ClassValue } from 'clsx'
import { clsx } from 'clsx'
import yaml from 'js-yaml'
import ms from 'ms'
import { twMerge } from 'tailwind-merge'

// Temporary K8s object type - will be replaced by actual K8s SDK types
interface KubeObject {
  apiVersion: string
  kind: string
  metadata?: any
  spec?: any
  status?: any
}

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function valueUpdater<T extends Updater<any>>(updaterOrValue: T, ref: Ref) {
  ref.value = typeof updaterOrValue === 'function'
    ? updaterOrValue(ref.value)
    : updaterOrValue
}

export function dumpKubeYaml<T extends Partial<KubeObject>>(obj: T, withStatus?: boolean) {
  return yaml.dump({
    apiVersion: obj.apiVersion,
    kind: obj.kind,
    metadata: obj.metadata,
    spec: obj.spec,
    status: withStatus ? obj.status : undefined,
  })
}

export function loadKubeYaml<T extends KubeObject>(data: string) {
  return yaml.load(data) as T
}

export function getItemAge(createdAt?: string, short = false) {
  if (!createdAt)
    return 'unknown'
  const age = Date.now() - new Date(createdAt).valueOf()
  return short ? ms(age, { long: false }) : ms(age, { long: true })
}

export const convertAutoInterval = (interval: ms.StringValue | 'auto', limit: number, timeGroup: [number, number]) => {
  if (interval === 'auto') {
    if (!timeGroup?.length || !limit)
      throw new Error('Invalid time range')

    const [start, end] = timeGroup
    const timeDiffMs = Math.abs(start - end)
    const intervalSeconds = Math.max(1, Math.round(timeDiffMs / (limit * 1000)))

    if (intervalSeconds < 60)
      return `${intervalSeconds}s`
    if (intervalSeconds < 3600)
      return `${Math.round(intervalSeconds / 60)}m`
    if (intervalSeconds < 86400)
      return `${Math.round(intervalSeconds / 3600)}h`
    return `${Math.round(intervalSeconds / 86400)}d`
  }
  // avoid too small interval
  if (ms(interval) < 5000)
    return '5s'
  return interval
}

export function moveToFirst<T>(arr: T[], value: T): T[] {
  const index = arr.indexOf(value)
  if (index > -1) {
    arr.splice(index, 1)
    arr.unshift(value)
  }
  return arr
}
