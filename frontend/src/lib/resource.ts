import { V1Node, V1Pod } from "@kubernetes/client-node"

export type Quantity = string

export interface ResourcePlugin {
    name(): string[]
    
    description(): Record<string, string>
    
    resourceQuantityOfPod(pod: V1Pod): Record<string, Quantity>
    
    resourceQuantityOfNode(node: V1Node): Record<string, Quantity>
    
    formatQuantity(resName: string, quantity: Quantity): string

    nodeMetadata<T>(node: V1Node): T
}


/**
 * Humanizes Kubernetes resource.Quantity values to a readable format
 * Supports conversion of raw values (like 49962863820) or formatted values (like 15Gi)
 * to appropriate units (B, Ki, Mi, Gi, Ti, Pi, Ei)
 *
 * @param value - The resource.Quantity value to humanize
 * @param targetUnit - Optional target unit (B, Ki, Mi, Gi, Ti, Pi, Ei)
 * @param decimals - Number of decimal places to show (default: 2)
 * @returns Humanized string representation
 */
export function humanizeK8sQuantityBytes(value: string | number, targetUnit?: string, decimals = 1): string {
  // If value is already a string with a unit (like 15Gi), parse it
  if (typeof value === 'string' && /^\d+[KMGTPE]i|[KMGTPE]?$/.test(value)) {
    const numericPart = Number.parseInt(value.replace(/\D/g, ''), 10)
    const unitPart = value.replace(/\d/g, '')

    // If target unit is specified, convert to that unit
    if (targetUnit) {
      return convertToUnit(numericPart, unitPart, targetUnit, decimals)
    }

    // Otherwise, return the original formatted value
    return value
  }

  // Convert numeric value to string with appropriate unit
  const numericValue = typeof value === 'string' ? Number.parseInt(value, 10) : value

  // If target unit is specified, convert directly to that unit
  if (targetUnit) {
    return convertToUnit(numericValue, 'B', targetUnit, decimals)
  }

  // Auto-select the most appropriate unit
  return autoFormatBytes(numericValue, decimals)
}

/**
 * Converts a value from one unit to another
 */
function convertToUnit(value: number, fromUnit: string, toUnit: string, decimals: number): string {
  const units = ['B', 'Ki', 'Mi', 'Gi', 'Ti', 'Pi', 'Ei']

  // Normalize units (handle both IEC and SI units)
  const normalizedFromUnit = fromUnit.endsWith('i') ? fromUnit : ((fromUnit === 'B' || fromUnit === '') ? 'B' : `${fromUnit}i`)
  const normalizedToUnit = toUnit.endsWith('i') ? toUnit : (toUnit === 'B' ? 'B' : `${toUnit}i`)

  // Get unit indices
  const fromIndex = units.indexOf(normalizedFromUnit)
  const toIndex = units.indexOf(normalizedToUnit)

  if (fromIndex === -1 || toIndex === -1) {
    return `${value} ${fromUnit}` // Return original if units not recognized
  }

  // Convert to bytes first
  const bytes = value * 1024 ** fromIndex

  // Then convert to target unit
  const converted = bytes / 1024 ** toIndex

  // Format with specified decimal places
  return `${converted.toFixed(decimals)} ${toUnit}`
}

/**
 * Automatically formats a byte value to the most appropriate unit
 */
export function autoFormatBytes(bytes: number, decimals: number): string {
  if (bytes === 0)
    return '0 B'

  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB']
  const k = 1024

  // Determine appropriate unit
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  // Format with specified decimal places
  return `${Number.parseFloat((bytes / k ** i).toFixed(decimals))} ${units[i]}`
}

/**
 * Humanizes TFlops values to a readable format
 * Automatically converts between TF and PF based on value magnitude
 *
 * @param value - The TFlops value to humanize (numeric)
 * @returns Humanized string representation (e.g., "60 TF", "300 PF")
 */
export function humanizeK8sQuantityTFlops(value: string | number): string {
  // Convert string to number if needed
  let numericValue = typeof value === 'string' ? Number.parseInt(value, 10) : value

  if (typeof value === 'string' && value.endsWith('m')) {
    numericValue = numericValue / 1000
  }

  // Handle zero case
  if (numericValue === 0)
    return '0 TF'

  // Convert to PF if value is 1000 TF or greater
  if (numericValue >= 1000) {
    // Convert to PF (1 PF = 1000 TF)
    const pfValue = numericValue / 1000

    // If the PF value is a whole number, don't show decimal places
    if (pfValue % 1 === 0)
      return `${pfValue} PF`
    else
      return `${pfValue.toFixed(2)} PF`
  }

  // For values less than 1000 TF, display as TF
  // If the value is a whole number, don't show decimal places
  if (numericValue % 1 === 0)
    return `${numericValue} TF`
  else
    return `${numericValue.toFixed(2)} TF`
}

/**
 * Converts Kubernetes resource quantity strings to numbers
 * Handles formats like:
 * - "100m" (millicores or milliunits)
 * - "1" (cores/units)
 * - "100Ki", "100Mi", "100Gi", "100Ti", "100Pi", "100Ei" (binary suffixes)
 * - "100K", "100M", "100G", "100T", "100P", "100E" (decimal suffixes)
 *
 * @param quantity The k8s resource quantity string or number
 * @param inBytes Whether to return memory values in bytes (true) or in the most appropriate unit (false)
 * @returns The numeric value as a number
 */
export function k8sQuantityToNumber(quantity: string | number, inBytes: boolean = true): number {
    // If quantity is already a number, return it
    if (typeof quantity === 'number') {
      return quantity
    }
  
    // If empty string, return 0
    if (!quantity) {
      return 0
    }
  
    // Handle milliunits/millicores (e.g., "100m")
    if (quantity.endsWith('m')) {
      return Number.parseFloat(quantity.slice(0, -1)) / 1000
    }
  
    // Binary suffixes (powers of 2)
    const binarySuffixes: Record<string, number> = {
      Ki: 1024,
      Mi: 1024 ** 2,
      Gi: 1024 ** 3,
      Ti: 1024 ** 4,
      Pi: 1024 ** 5,
      Ei: 1024 ** 6,
    }
  
    // Decimal suffixes (powers of 10)
    const decimalSuffixes: Record<string, number> = {
      K: 1000,
      M: 1000 ** 2,
      G: 1000 ** 3,
      T: 1000 ** 4,
      P: 1000 ** 5,
      E: 1000 ** 6,
    }
  
    // Check for binary suffixes
    for (const [suffix, multiplier] of Object.entries(binarySuffixes)) {
      if (quantity.endsWith(suffix)) {
        const value = Number.parseFloat(quantity.slice(0, -suffix.length))
        return inBytes ? value * multiplier : value
      }
    }
  
    // Check for decimal suffixes
    for (const [suffix, multiplier] of Object.entries(decimalSuffixes)) {
      if (quantity.endsWith(suffix)) {
        const value = Number.parseFloat(quantity.slice(0, -suffix.length))
        return inBytes ? value * multiplier : value
      }
    }
  
    // No suffix, just a plain number
    return Number.parseFloat(quantity)
  }
  