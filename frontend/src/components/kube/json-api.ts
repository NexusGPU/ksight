import type { PartialDeep, ValueOf } from 'type-fest'
import type { Defaulted, QueryParam, QueryParams } from './utils/index'
import { merge } from 'lodash'
import { getRequestOptionsWithEnv } from '~/stores/env'
import { EventEmitter } from '../common/event-emitter'
import { json, stringify } from './utils/index'

export interface JsonApiData {}

export interface JsonApiError {
  code?: number
  message?: string
  errors?: { id: string, title: string, status?: number }[]
}

export interface JsonApiParams<D> {
  data?: PartialDeep<D> // request body
}

export interface JsonApiLog {
  method: string
  reqUrl: string
  reqInit: RequestInit
  data?: any
  error?: any
}

export type GetRequestOptions = () => Promise<RequestInit>

export interface JsonApiConfig {
  apiBase?: string
  serverAddress?: string
  debug?: boolean
  getRequestOptions?: GetRequestOptions
}

export type ParamsAndQuery<Params, Query> = (
  ValueOf<Query> extends QueryParam
    ? Params & { query?: Query }
    : Params & { query?: undefined }
)

export class JsonApi<Data = JsonApiData, Params extends JsonApiParams<Data> = JsonApiParams<Data>> {
  static readonly reqInitDefault = {
    headers: {
      'content-type': 'application/json',
    },
  }

  protected readonly reqInit: Defaulted<RequestInit, keyof typeof JsonApi['reqInitDefault']>

  static readonly configDefault: Partial<JsonApiConfig> = {
    serverAddress: location.origin,
    apiBase: '/proxy',
    debug: false,
    getRequestOptions: getRequestOptionsWithEnv,
  }

  constructor(public readonly config: JsonApiConfig, reqInit?: RequestInit) {
    this.config = Object.assign({}, JsonApi.configDefault, config)
    this.reqInit = merge({}, JsonApi.reqInitDefault, reqInit)
    this.parseResponse = this.parseResponse.bind(this)
    this.getRequestOptions = this.config.getRequestOptions ?? (() => Promise.resolve({}))
  }

  public readonly onData = new EventEmitter<[Data, Response]>()
  public readonly onError = new EventEmitter<[JsonApiErrorParsed, Response]>()
  private readonly getRequestOptions: GetRequestOptions

  async getResponse<Query>(
    path: string,
    params?: ParamsAndQuery<Params, Query>,
    init: RequestInit = {},
  ): Promise<Response> {
    let reqUrl = `${this.config.serverAddress}${this.config.apiBase}${path}`
    const reqInit = merge(
      {
        method: 'get',
      },
      this.reqInit,
      await this.getRequestOptions(),
      init,
    )
    const { query } = params ?? {}

    if (query) {
      const queryString = stringify(query as unknown as QueryParams)

      reqUrl += (reqUrl.includes('?') ? '&' : '?') + queryString
    }

    return fetch(reqUrl, reqInit)
  }

  get<OutData = Data, Query = QueryParams>(
    path: string,
    params?: ParamsAndQuery<Params, Query>,
    reqInit: RequestInit = {},
  ) {
    return this.request<OutData, Query>(path, params, { ...reqInit, method: 'GET' })
  }

  post<OutData = Data, Query = QueryParams>(
    path: string,
    params?: ParamsAndQuery<Params, Query>,
    reqInit: RequestInit = {},
  ) {
    return this.request<OutData, Query>(path, params, { ...reqInit, method: 'POST' })
  }

  put<OutData = Data, Query = QueryParams>(
    path: string,
    params?: ParamsAndQuery<Params, Query>,
    reqInit: RequestInit = {},
  ) {
    return this.request<OutData, Query>(path, params, { ...reqInit, method: 'PUT' })
  }

  patch<OutData = Data, Query = QueryParams>(
    path: string,
    params?: (ParamsAndQuery<Omit<Params, 'data'>, Query> & { data?: PartialDeep<Data> }),
    reqInit: RequestInit = {},
  ) {
    return this.request<OutData, Query>(path, params, { ...reqInit, method: 'PATCH' })
  }

  del<OutData = Data, Query = QueryParams>(
    path: string,
    params?: ParamsAndQuery<Params, Query>,
    reqInit: RequestInit = {},
  ) {
    return this.request<OutData, Query>(path, params, { ...reqInit, method: 'DELETE' })
  }

  protected async request<OutData, Query = QueryParams>(
    path: string,
    params: (ParamsAndQuery<Omit<Params, 'data'>, Query> & { data?: unknown }) | undefined,
    init: Defaulted<RequestInit, 'method'>,
  ) {
    let reqUrl = `${this.config.serverAddress}${this.config.apiBase}${path}`
    const reqInit = merge(
      {},
      this.reqInit,
      await this.getRequestOptions(),
      init,
    )
    const { data, query } = params || {}

    if (data && !reqInit.body) {
      reqInit.body = JSON.stringify(data)
    }

    if (query) {
      const queryString = stringify(query as unknown as QueryParams)

      reqUrl += (reqUrl.includes('?') ? '&' : '?') + queryString
    }
    const infoLog: JsonApiLog = {
      method: reqInit.method.toUpperCase(),
      reqUrl,
      reqInit,
    }

    const res = await fetch(reqUrl, reqInit)

    return this.parseResponse<OutData>(res, infoLog)
  }

  protected async parseResponse<OutData>(res: Response, log: JsonApiLog): Promise<OutData> {
    const { status } = res

    const text = await res.text()
    let data: any

    try {
      data = text ? json.parse(text) : '' // DELETE-requests might not have response-body
    }
    catch {
      data = text
    }

    if (status >= 200 && status < 300) {
      this.onData.emit(data, res)
      this.writeLog({ ...log, data })

      // eliminate managedFields
      if ('items' in data) {
        data.items.forEach((item: any) => delete item.metadata.managedFields)
      }
      else {
        delete data.metadata.managedFields
      }
      return data
    }

    if (log.method === 'GET' && res.status === 403) {
      this.writeLog({ ...log, error: data })
      throw data
    }

    const error = new JsonApiErrorParsed(data, this.parseError(data, res), status)

    this.onError.emit(error, res)
    this.writeLog({ ...log, error })

    throw error
  }

  protected parseError(error: JsonApiError | string, res: Response): string[] {
    if (typeof error === 'string') {
      return [error]
    }

    if (Array.isArray(error.errors)) {
      return error.errors.map(error => error.title)
    }

    if (error.message) {
      return [error.message]
    }

    return [res.statusText || 'Error!']
  }

  protected writeLog(log: JsonApiLog) {
    const { method, reqUrl, ...params } = log

    console.debug(`[JSON-API] request ${method} ${reqUrl}`, params)
  }
}

export class JsonApiErrorParsed {
  isUsedForNotification = false
  statusCode: number

  constructor(private error: JsonApiError | DOMException, private messages: string[], statusCode: number) {
    this.statusCode = statusCode
  }

  get isAborted() {
    return this.error.code === DOMException.ABORT_ERR
  }

  toString() {
    return this.messages.join('\n')
  }
}
