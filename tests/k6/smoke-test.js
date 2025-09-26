import http from 'k6/http'
import { check, sleep } from 'k6'
import { BASE_URL, pickExistingPatientId, rfc3339 } from './helpers.js'
export { handleSummary } from './summary.js'

export const options = {
  vus: 1,
  duration: '30s',
}

export default function () {
  const health = http.get(`${BASE_URL}/healthz`)
  check(health, { 'health 200': (r) => r.status === 200 })

  const bh = http.get(`${BASE_URL}/business-hours`)
  check(bh, { 'business-hours 200': (r) => r.status === 200 })

  const today = http.get(`${BASE_URL}/appointments/today`)
  check(today, { 'today 200': (r) => r.status === 200 })

  const start = new Date(); start.setDate(start.getDate()-3)
  const end = new Date(); end.setDate(end.getDate()+3)
  const range = http.get(`${BASE_URL}/appointments?start_time=${encodeURIComponent(rfc3339(start))}&end_time=${encodeURIComponent(rfc3339(end))}`)
  check(range, { 'range 200': (r) => r.status === 200 })

  const pid = pickExistingPatientId()
  if (pid) {
    const p = http.get(`${BASE_URL}/patients/${pid}`)
    check(p, { 'patient 200|404': (r) => r.status === 200 || r.status === 404 })
  }

  sleep(1)
}
