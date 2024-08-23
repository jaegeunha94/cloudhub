// Libraries
import React, {Component} from 'react'
import {connect} from 'react-redux'

// Components
import LayoutRenderer from 'src/shared/components/LayoutRenderer'
import {Page} from 'src/reusable_ui'
import TimeZoneToggle from 'src/shared/components/time_zones/TimeZoneToggle'

// Actions
import * as appActions from 'src/shared/actions/app'

// Constants
import {STATUS_PAGE_TIME_RANGE} from 'src/shared/data/timeRanges'
import {fixtureStatusPageCells} from 'src/status/fixtures'
import {
  TEMP_VAR_DASHBOARD_TIME,
  TEMP_VAR_UPPER_DASHBOARD_TIME,
} from 'src/shared/constants'

// Types
import {
  Cell,
  Source,
  Template,
  TemplateType,
  TemplateValueType,
  TimeZones,
} from 'src/types'

import {ErrorHandling} from 'src/shared/decorators/errors'

interface OwnProps {
  source: Source
  shellModalVisible: boolean
}

interface StateProps {
  timeZone: TimeZones
  onSetTimeZone: typeof appActions.setTimeZone
}

const timeRange = STATUS_PAGE_TIME_RANGE

type Props = StateProps & OwnProps

@ErrorHandling
class StatusPage extends Component<Props> {
  constructor(props: Props) {
    super(props)

    this.state = {
      shellModalVisible: false,
    }
  }

  public render() {
    const {source, onSetTimeZone, timeZone} = this.props
    const cells = fixtureStatusPageCells(source)

    return (
      <Page>
        <Page.Header fullWidth={true}>
          <Page.Header.Left>
            <Page.Title title="Status" />
          </Page.Header.Left>
          <Page.Header.Right showSourceIndicator={true}>
            <TimeZoneToggle onSetTimeZone={onSetTimeZone} timeZone={timeZone} />
          </Page.Header.Right>
        </Page.Header>
        <Page.Contents fullWidth={true}>
          <div className="dashboard container-fluid full-width">
            {cells.length ? (
              <LayoutRenderer
                host=""
                sources={[]}
                cells={this.reBuildCell(cells)}
                source={source}
                manualRefresh={0}
                isEditable={false}
                isStatusPage={true}
                isStaticPage={false}
                timeRange={timeRange}
                templates={this.templates}
              />
            ) : (
              <span>Loading Status Page...</span>
            )}
          </div>
        </Page.Contents>
      </Page>
    )
  }

  private reBuildCell(cells: Cell[]) {
    const {timeZone} = this.props

    return cells.map(cell => {
      if (cell.i === 'alerts-bar-graph') {
        return {
          ...cell,
          queries: cell.queries.map(i => {
            return {
              ...i,
              groupbys: ['time(1d)'],
              wheres: [],
              tz:
                timeZone === TimeZones.UTC
                  ? 'UTC'
                  : `${Intl.DateTimeFormat().resolvedOptions().timeZone}`,
            }
          }),
        }
      } else {
        return cell
      }
    })
  }

  private get templates(): Template[] {
    const dashboardTime = {
      id: 'dashtime',
      tempVar: TEMP_VAR_DASHBOARD_TIME,
      type: TemplateType.Constant,
      label: '',
      values: [
        {
          value: timeRange.lower,
          type: TemplateValueType.Constant,
          selected: true,
          localSelected: true,
        },
      ],
    }

    const upperDashboardTime = {
      id: 'upperdashtime',
      tempVar: TEMP_VAR_UPPER_DASHBOARD_TIME,
      type: TemplateType.Constant,
      label: '',
      values: [
        {
          value: 'now()',
          type: TemplateValueType.Constant,
          selected: true,
          localSelected: true,
        },
      ],
    }

    return [dashboardTime, upperDashboardTime]
  }
}

const mstp = ({app}) => ({
  timeZone: app.persisted.timeZone,
})

const mdtp = {
  onSetTimeZone: appActions.setTimeZone,
}

export default connect(mstp, mdtp)(StatusPage)
