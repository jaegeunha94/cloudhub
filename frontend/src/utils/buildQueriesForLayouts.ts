import _ from 'lodash'

import {buildQuery} from 'src/utils/influxql'
import {TYPE_SHIFTED, TYPE_QUERY_CONFIG} from 'src/dashboards/constants'
import {
  TEMP_VAR_DASHBOARD_TIME,
  TEMP_VAR_UPPER_DASHBOARD_TIME,
} from 'src/shared/constants'
import {timeRanges} from 'src/shared/data/timeRanges'

import {Cell, CellQuery, LayoutQuery, TimeRange} from 'src/types'

const buildCannedDashboardQuery = (
  query: LayoutQuery | CellQuery,
  {lower, upper}: TimeRange,
  host: string,
  instance?: object,
  measurement?: string
): string => {
  const {defaultGroupBy} = timeRanges.find(range => range.lower === lower) || {
    defaultGroupBy: '5m',
  }

  let text = query.query
  const wheres = _.get(query, 'wheres')
  const groupbys = _.get(query, 'groupbys')

  if (upper) {
    text += ` where time > '${lower}' AND time < '${upper}'`
  } else {
    text += ` where time > ${lower}`
  }

  if (host) {
    if (measurement === 'cloudwatch_aws_application_elb') {
      text += ` and \"load_balancer\" = '${host}' or \"host\" = '${host}'`
    } else if (measurement === 'ipmi_sensor') {
      text += ` and \"hostname" \= '${host}'`
    } else if (measurement === 'snmp_nx') {
      text += ` and \"agent_host" \= '${host}'`
    } else {
      text += ` and \"host\" = '${host}'`
    }
  }

  if (instance) {
    if (measurement === 'cloudwatch_aws_ec2') {
      text += ` and \"region\" = '${instance['namespace']}' and \"instance_id\" = '${instance['instanceid']}'`
    } else if (measurement === 'compute.googleapis.com/instance') {
      text += ` and \"project_id\" = '${instance['namespace']}' and \"instance_id\" = '${instance['instanceid']}'`
    } else if (measurement === 'openstack_server_diagnostics') {
      text += ` and \"tenant\" = '${instance['namespace']}' and \"server_id\" = '${instance['instanceid']}'`
    } else {
      text += ` and \"host\" = '${instance['instancename']}'`
    }
  }

  if (wheres && wheres.length > 0) {
    text += ` and ${wheres.join(' and ')}`
  }

  if (groupbys) {
    if (groupbys.find(g => g.includes('time'))) {
      text += ` group by ${groupbys.join(',')}`
    } else if (groupbys.length > 0) {
      text += ` group by time(${defaultGroupBy}),${groupbys.join(',')}`
    } else {
      text += ` group by time(${defaultGroupBy})`
    }
  } else {
    text += ` group by time(${defaultGroupBy})`
  }

  return text
}

const addTimeBoundsToRawText = (rawText: string): string => {
  if (!rawText) {
    return
  }

  const dashboardTimeRegex = new RegExp(
    `time( )?>( )?${TEMP_VAR_DASHBOARD_TIME}`,
    'g'
  )
  const dashboardTimeText: string = `time > ${TEMP_VAR_DASHBOARD_TIME}`
  const isUsingTimeSelectorBounds: boolean = !_.isEmpty(
    rawText.match(dashboardTimeRegex)
  )

  if (isUsingTimeSelectorBounds) {
    const upperTimeBoundRegex = new RegExp('time( )?<', 'g')
    const hasUpperTimeBound = !_.isEmpty(rawText.match(upperTimeBoundRegex))
    if (
      rawText.indexOf(TEMP_VAR_UPPER_DASHBOARD_TIME) === -1 &&
      !hasUpperTimeBound
    ) {
      const upperDashboardTimeText = `time < ${TEMP_VAR_UPPER_DASHBOARD_TIME}`
      const fullTimeText = `${dashboardTimeText} AND ${upperDashboardTimeText}`
      const boundedQueryText = rawText.replace(dashboardTimeRegex, fullTimeText)
      return boundedQueryText
    }
  }
  return rawText
}

export const buildQueriesForLayouts = (
  cell: Cell,
  timeRange: TimeRange,
  host: string,
  instance?: object
): CellQuery[] => {
  return cell.queries.map(query => {
    let queryText: string
    // Canned dashboards use an different a schema different from queryConfig.
    if (query.queryConfig) {
      const {
        queryConfig: {database, measurement, fields, shifts, rawText, range},
      } = query
      const tR: TimeRange = range || {
        upper: TEMP_VAR_UPPER_DASHBOARD_TIME,
        lower: TEMP_VAR_DASHBOARD_TIME,
      }

      queryText =
        addTimeBoundsToRawText(rawText) ||
        buildQuery(TYPE_QUERY_CONFIG, tR, query.queryConfig)
      const isParsable: boolean =
        !_.isEmpty(database) && !_.isEmpty(measurement) && fields.length > 0

      if (shifts && shifts.length && isParsable) {
        const shiftedQueries: string[] = shifts
          .filter(s => s.unit)
          .map(s => buildQuery(TYPE_SHIFTED, timeRange, query.queryConfig, s))

        queryText = `${queryText};${shiftedQueries.join(';')}`
      }
    } else {
      queryText = buildCannedDashboardQuery(
        query,
        timeRange,
        host,
        instance,
        _.get(cell, 'measurement')
      )
    }

    return {...query, text: queryText}
  })
}
