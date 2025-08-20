import type { ShellParams } from '../type'

export const getPlatform = () => {
  // only return mac or win
  if ('platform' in window.navigator) {
    if (window.navigator.platform.startsWith('Mac'))
      return 'Mac'
    return 'Win'
  }
  return 'Win'
}

export const buildBasicUrl = (params: ShellParams) => {
  const { namespace, podName, type, container } = params
  const res = {
    baseUrl: `/shell-ws`,
    params: {
      url: `/api/v1/namespaces/${namespace}/pods/${podName}/${type}?container=${container}&command=/bin/sh&command=-c&command=(bash||ash||sh)&stdin=true&stdout=true&tty=true`,
      env: getEnvRef()?.name,
    },
  }
  return res
}

export const buildUrlWithParams = (baseUrl: string, params: Record<string, any>) => {
  const url = new URL(baseUrl)
  const searchParams = new URLSearchParams(params)
  url.search = searchParams.toString()
  return url.href
}
