import http from 'k6/http'
import { check, sleep } from 'k6'
import { Trend } from 'k6/metrics'
import { BASE_URL, pickExistingPatientId, rfc3339 } from './helpers.js'
export { handleSummary } from './summary.js'

const readDuration = new Trend('read_duration')

export const options = {
  scenarios: {
    steady_read: {
      executor: 'ramping-vus',
      stages: [
        { duration: '30s', target: 10 },
        { duration: '1m', target: 25 },
        { duration: '2m', target: 50 },
        { duration: '30s', target: 0 },
      ],
      gracefulRampDown: '15s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    'http_req_duration{scenario:steady_read}': ['p(95)<400'],
    checks: ['rate>0.95'],
  },
}

export default function () {
  const start = new Date(); start.setDate(start.getDate()-1)
  const end = new Date(); end.setDate(end.getDate()+1)

  const res = http.batch([
    ['GET', `${BASE_URL}/healthz`, null, {}],
    ['GET', `${BASE_URL}/business-hours`, null, {}],
    ['GET', `${BASE_URL}/appointments/today`, null, {}],
    ['GET', `${BASE_URL}/appointments?start_time=${encodeURIComponent(rfc3339(start))}&end_time=${encodeURIComponent(rfc3339(end))}`, null, {}],
  ])

  const ok = res.every(r => r.status === 200)
  check({ ok }, { 'batch 200s': (v) => v.ok })

  res.forEach(r => readDuration.add(r.timings.duration))

  const pid = pickExistingPatientId()
  if (pid) {
    const p = http.get(`${BASE_URL}/patients/${pid}`)
    check(p, { 'patient 200|404': (r) => r.status === 200 || r.status === 404 })
  }

  sleep(0.5)
}
