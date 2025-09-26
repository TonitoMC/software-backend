import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js'
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.1/index.js'

export function handleSummary(data) {
  const outDir = __ENV.OUT_DIR || '.'
  return {
    [`${outDir}/summary.html`]: htmlReport(data),
    [`${outDir}/summary.txt`]: textSummary(data, { indent: ' ', enableColors: false }),
  }
}
