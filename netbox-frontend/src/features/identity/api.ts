import { request } from '@/api/http'

export interface AuthUserDTO {
  id: number
  username: string
  email: string
  first_name: string
  last_name: string
  is_staff: boolean
  is_superuser: boolean
}

export interface SessionDTO {
  user: AuthUserDTO
  permissions?: string[]
}

export interface LoginDTO {
  username: string
  password: string
}

export async function establishCSRF(): Promise<void> {
  await request({ method: 'GET', url: '/auth/csrf/' })
}

export async function getSession(): Promise<SessionDTO> {
  const response = await request<SessionDTO>({ method: 'GET', url: '/auth/session/' })
  return response.data
}

export async function createSession(credentials: LoginDTO): Promise<SessionDTO> {
  await establishCSRF()
  const response = await request<SessionDTO>({
    method: 'POST',
    url: '/auth/login/',
    data: credentials,
  })
  return response.data
}

export async function deleteSession(): Promise<void> {
  await request({ method: 'POST', url: '/auth/logout/' })
}
