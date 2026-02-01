export interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  role: string
  permissions: string[]
  tenant_id: string
  avatar_url?: string
  created_at: string
  updated_at: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_at: string
  user: User
}

export interface RegisterData {
  email: string
  password: string
  first_name: string
  last_name: string
  role: string
}

export interface ChangePasswordData {
  current_password: string
  new_password: string
  confirm_password: string
}

export interface ForgotPasswordData {
  email: string
}

export interface ResetPasswordData {
  token: string
  password: string
  confirm_password: string
}
