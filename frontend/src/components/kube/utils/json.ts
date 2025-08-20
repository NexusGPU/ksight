import type { JsonValue } from 'type-fest'

export function parse(input: string): JsonValue {
  return JSON.parse(input)
}
