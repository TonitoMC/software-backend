import http from 'k6/http'
import { check, sleep } from 'k6'
import { BASE_URL, rfc3339 } from './helpers.js'
export { handleSummary } from './summary.js'

export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 50,
      maxVUs: 200,
      stages: [
        { target: 50, duration: '30s' },
        { target: 100, duration: '1m' },
        { target: 150, duration: '1m' },
        { target: 200, duration: '1m' },
        { target: 0, duration: '30s' },
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1000'],
    checks: ['rate>0.90'],
  },
}

export default function () {
  const start = new Date(); start.setDate(start.getDate()-1)
  const end = new Date(); end.setDate(end.getDate()+1)

  // hit a mix of fast endpoints
  const r1 = http.get(`${BASE_URL}/healthz`)
  const r2 = http.get(`${BASE_URL}/business-hours`)
  const r3 = http.get(`${BASE_URL}/appointments?start_time=${encodeURIComponent(rfc3339(start))}&end_time=${encodeURIComponent(rfc3339(end))}`)

  check(r1, { 'health 200': (r) => r.status === 200 })
  check(r2, { 'bh 200': (r) => r.status === 200 })
  check(r3, { 'appts 200': (r) => r.status === 200 })

  sleep(0.2)
}
