import http from 'k6/http'
import { check, sleep } from 'k6'

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:4000'

export function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

export function getHeaders(token) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  return headers
}

export function loginOrRegister() {
  const username = __ENV.USERNAME || `testuser_${Math.floor(Math.random()*1e6)}`
  const password = __ENV.PASSWORD || 'TestPass123!'
  const email = `${username}@example.com`

  // Try register (may conflict), then login
  http.post(`${BASE_URL}/register`, JSON.stringify({ username, email, password }), { headers: getHeaders() })
  const res = http.post(`${BASE_URL}/login`, JSON.stringify({ username, password }), { headers: getHeaders() })
  check(res, { 'login 200': (r) => r.status === 200 })
  const body = res.json()
  const token = body?.token
  return { token, user: body?.user }
}

export function pickExistingPatientId() {
  const q = 'a' // broad query to increase hit chance
  const res = http.get(`${BASE_URL}/patients/search?q=${encodeURIComponent(q)}`)
  if (res.status === 200 && Array.isArray(res.json()) && res.json().length > 0) {
    return res.json()[0].id
  }
  return null
}

export function rfc3339(date) {
  return date.toISOString()
}
