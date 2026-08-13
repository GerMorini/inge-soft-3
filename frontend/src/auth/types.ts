export interface AddressInput {
  street: string
  number: string
  apartment: string
  city: string
  province: string
}

export interface RegisterInput {
  firstName: string
  lastName: string
  phone: string
  address: AddressInput
  username: string
  email: string
  password: string
}

export interface RegisteredUser {
  id: number
  username: string
  email: string
}

export interface LoginInput {
  username: string
  password: string
}

export interface LoginResult {
  accessToken: string
  tokenType: 'Bearer'
  expiresIn: 1800
}

export interface CurrentUser {
  id: number
  username: string
}

export interface ErrorBody {
  code: string
  message: string
  fields?: Record<string, string[]>
}

export class ApiError extends Error {
  readonly status: number
  readonly body: ErrorBody

  constructor(status: number, body: ErrorBody) {
    super(body.message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }
}
