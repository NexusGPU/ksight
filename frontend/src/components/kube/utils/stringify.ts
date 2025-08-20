export type QueryParam = string | number | boolean | null | undefined | readonly string[] | readonly number[] | readonly boolean[]
export type QueryParams = Partial<Record<string, QueryParam | undefined>>

export function stringify(params: QueryParams): string {
  const searchParams = new URLSearchParams()

  for (const [key, value] of Object.entries(params)) {
    if (value === null || value === undefined) {
      continue
    }

    if (Array.isArray(value)) {
      for (const item of value) {
        searchParams.append(key, String(item))
      }
    }
    else {
      searchParams.set(key, String(value))
    }
  }

  return searchParams.toString()
}
