import type { JsonApiData, JsonApiError } from './json-api'
import type { KubeJsonApiObjectMetadata } from './kube-object'
import { JsonApi } from './json-api'

export interface KubeJsonApiListMetadata {
  resourceVersion: string
  selfLink?: string
}

export interface KubeJsonApiDataList<T = KubeJsonApiData> {
  kind: string
  apiVersion: string
  items: T[]
  metadata: KubeJsonApiListMetadata
}

export interface KubeJsonApiData<
  Metadata extends KubeJsonApiObjectMetadata = KubeJsonApiObjectMetadata,
  Status = unknown,
  Spec = unknown,
> extends JsonApiData {
  kind: string
  apiVersion: string
  metadata: Metadata
  status?: Status
  spec?: Spec
  [otherKeys: string]: unknown
}

export interface KubeJsonApiError extends JsonApiError {
  code: number
  status: string
  message?: string
  reason: string
  details: {
    name: string
    kind: string
  }
}

export class KubeJsonApi extends JsonApi<KubeJsonApiData> {
  muteNotFoundError = false

  protected override parseError(error: KubeJsonApiError | string, res: Response): string[] {
    if (typeof error === 'string') {
      return [error]
    }

    const { status, reason, message } = error

    if (status && reason) {
      return [message || `${status}: ${reason}`]
    }

    return super.parseError(error, res)
  }
}
