// Libraries
import React, {FunctionComponent} from 'react'
import {Link} from 'react-router'
import classnames from 'classnames'

//components
import {HostsPageHostTab} from 'src/hosts/containers/HostsPageHostTab'

//types
import {HOSTS_TABLE_SIZING} from 'src/hosts/constants/tableSizing'
import {Host} from 'src/types'

//constants
import {NOT_AVAILABLE_STATUS} from 'src/hosts/constants/topology'

interface Props {
  sourceID: string
  host: Host
  focusedHost: string
  onClickTableRow: HostsPageHostTab['handleClickTableRow']
}

const HostRow: FunctionComponent<Props> = ({
  host,
  sourceID,
  focusedHost,
  onClickTableRow,
}) => {
  const {name, cpu, load, apps = []} = host
  const {NameWidth, StatusWidth, CPUWidth, LoadWidth} = HOSTS_TABLE_SIZING

  const CPUValue = isNaN(cpu) ? NOT_AVAILABLE_STATUS : `${cpu.toFixed(2)}%`
  const loadValue = isNaN(load) ? NOT_AVAILABLE_STATUS : `${load.toFixed(2)}`
  const dotClassName = classnames(
    'table-dot',
    Math.max(host.deltaUptime || 0, host.winDeltaUptime || 0) > 0
      ? 'dot-success'
      : 'dot-critical'
  )

  const focusedClasses = (): string => {
    if (name === focusedHost) return 'hosts-table--tr focused'
    return 'hosts-table--tr'
  }

  return (
    <div className={focusedClasses()} onClick={onClickTableRow(name)}>
      <div className="hosts-table--td" style={{width: NameWidth}}>
        <Link to={`/sources/${sourceID}/infrastructure/details/${name}`}>
          {name}
        </Link>
      </div>
      <div className="hosts-table--td" style={{width: StatusWidth}}>
        <div className={dotClassName} />
      </div>
      <div style={{width: CPUWidth}} className="monotype hosts-table--td">
        {CPUValue}
      </div>
      <div style={{width: LoadWidth}} className="monotype hosts-table--td">
        {loadValue}
      </div>
      <div className="hosts-table--td list-type">
        {apps.map((app, index) => {
          return (
            <span key={app}>
              <Link
                style={{marginLeft: '2px'}}
                to={{
                  pathname: `/sources/${sourceID}/infrastructure/details/${name}`,
                  query: {app},
                }}
              >
                {app}
              </Link>
              {index === apps.length - 1 ? null : ', '}
            </span>
          )
        })}
      </div>
    </div>
  )
}

export default HostRow
