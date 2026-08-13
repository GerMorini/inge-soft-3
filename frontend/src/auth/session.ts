export const accessTokenKey = 'accessToken'

export function storeAccessToken(token: string) {
  sessionStorage.setItem(accessTokenKey, token)
}

export function readAccessToken(): string | null {
  return sessionStorage.getItem(accessTokenKey)
}

export function clearAccessToken() {
  sessionStorage.removeItem(accessTokenKey)
}

export function isAccessTokenExpired(token: string, now = Date.now()): boolean {
  try {
    const encodedPayload = token.split('.')[1]
    if (!encodedPayload) return true
    const base64 = encodedPayload.replaceAll('-', '+').replaceAll('_', '/')
    const normalized = base64.padEnd(Math.ceil(base64.length / 4) * 4, '=')
    const payload = JSON.parse(atob(normalized)) as { exp?: number }
    return typeof payload.exp !== 'number' || now >= payload.exp * 1000
  } catch {
    return true
  }
}
