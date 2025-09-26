import http from 'k6/http'
import { check } from 'k6'
import { BASE_URL, getHeaders, pickExistingPatientId } from './helpers.js'
export { handleSummary } from './summary.js'

export const options = { vus: 1, iterations: 1 }

export default function () {
  // 1) CORS probe: preflight OPTIONS
  const cors = http.request('OPTIONS', `${BASE_URL}/patients/search?q=a`, null, {
    headers: {
      'Origin': 'https://malicious.example',
      'Access-Control-Request-Method': 'GET',
      'Access-Control-Request-Headers': 'Authorization, Content-Type',
    },
  })
  check(cors, {
    'cors status 204|200|4xx': (r) => [200,204,400,403].includes(r.status),
  })

  // 2) SQL Injection attempt on search
  const sqli = http.get(`${BASE_URL}/patients/search?q=${encodeURIComponent("a' OR '1'='1")}`)
  check(sqli, {
    'sqli not 500': (r) => r.status < 500,
  })

  // 3) XSS attempt on search
  const xss = http.get(`${BASE_URL}/patients/search?q=${encodeURIComponent('<script>alert(1)</script>')}`)
  check(xss, {
    'xss not 500': (r) => r.status < 500,
  })

  // 4) Auth-required endpoints without token
  const pid = pickExistingPatientId()
  if (pid) {
    const p = http.get(`${BASE_URL}/patients/${pid}`)
    check(p, { 'patients/:id 2xx|404|401': (r) => [200,401,403,404].includes(r.status) })
  }

  // 5) Invalid token
  const resInvalid = http.get(`${BASE_URL}/appointments/today`, { headers: getHeaders('invalid.token.here') })
  check(resInvalid, { 'invalid token not 2xx': (r) => r.status >= 400 })
}
