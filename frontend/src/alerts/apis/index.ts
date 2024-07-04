import {proxy} from 'src/utils/queryUrlGenerator'
import {TimeRange} from 'src/types'

export const getAlerts = (
  source: string,
  timeRange: TimeRange,
  limit: number,
  db: string
) => {
  const query = `SELECT host, value, level, alertName, triggerType FROM cloudhub_alerts WHERE time >= '${
    timeRange.lower
  }' AND time <= '${timeRange.upper}' ORDER BY time desc ${
    limit ? `LIMIT ${limit}` : ''
  }`

  return proxy({
    source,
    query,
    db,
  })
}
