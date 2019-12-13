import React, {PureComponent} from 'react'
import {ROUTER_TABLE_SIZING} from 'src/addon/128t/constants'
import {Router} from 'src/addon/128t/types'
import {fixedDecimalPercentage} from 'src/shared/utils/decimalPlaces'
import {transBps} from 'src/shared/utils/units'

interface Props {
  router: Router
}

interface State {
  showModal: boolean
}

class RouterTableRow extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props)

    this.state = {
      showModal: false,
    }
  }

  public focusedClasses = (): string => {
    return 'hosts-table--tr'
  }

  private TableItem = ({width, title}) => {
    return (
      <div
        className="hosts-table--td"
        style={{width: width, alignItems: 'center'}}
      >
        {title}
      </div>
    )
  }

  render() {
    const {
      assetId,
      locationCoordinates,
      managementConnected,
      bandwidth_avg,
      session_arrivals,
      enabled,
      role,
      startTime,
      softwareVersion,
      memoryUsage,
      cpuUsage,
      diskUsage,
      // topSources,
      // topSessions,
    } = this.props.router

    const {
      ASSETID,
      LOCATIONCOORDINATES,
      MANAGEMENTCONNECTED,
      BANDWIDTH_AVG,
      SESSION_CNT_AVG,
      ENABLED,
      ROLE,
      STARTTIME,
      SOFTWAREVERSION,
      MEMORYUSAGE,
      CPUUSAGE,
      DISKUSAGE,
    } = ROUTER_TABLE_SIZING

    return (
      <div className={this.focusedClasses()}>
        <this.TableItem title={assetId} width={ASSETID} />
        <this.TableItem title={role} width={ROLE} />
        <this.TableItem
          title={(() => {
            return enabled ? 'True' : 'False'
          })()}
          width={ENABLED}
        />
        <this.TableItem
          title={locationCoordinates}
          width={LOCATIONCOORDINATES}
        />
        <this.TableItem
          title={managementConnected}
          width={MANAGEMENTCONNECTED}
        />
        <this.TableItem title={startTime} width={STARTTIME} />
        <this.TableItem title={softwareVersion} width={SOFTWAREVERSION} />
        <this.TableItem
          title={fixedDecimalPercentage(cpuUsage, 2)}
          width={CPUUSAGE}
        />
        <this.TableItem
          title={fixedDecimalPercentage(memoryUsage, 2)}
          width={MEMORYUSAGE}
        />
        <this.TableItem
          title={fixedDecimalPercentage(diskUsage, 2)}
          width={DISKUSAGE}
        />
        <this.TableItem
          title={transBps(bandwidth_avg * 8, 2)}
          width={BANDWIDTH_AVG}
        />
        <this.TableItem title={session_arrivals} width={SESSION_CNT_AVG} />
      </div>
    )
  }
}

export default RouterTableRow
