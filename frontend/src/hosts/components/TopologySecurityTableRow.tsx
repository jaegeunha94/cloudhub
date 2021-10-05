import React, {PureComponent} from 'react'
import {CLOUD_HOST_SECURITY_TABLE_SIZING} from '../constants/tableSizing'

interface Props {
  rowData: any
}

class TopologyDetailsSectionTableRow extends PureComponent<Props> {
  constructor(props: Props) {
    super(props)
  }
  public render() {
    const {
      rowData: {port, protocol, source, destination, security_groups},
    } = this.props
    const {
      PortWidth,
      ProtocolWidth,
      SourceWidth,
      SecurityGroupsWidth,
    } = CLOUD_HOST_SECURITY_TABLE_SIZING
    console.log('props: ', this.props)
    return (
      <>
        <div className={`hosts-table--tr`}>
          <div className="hosts-table--td" style={{width: PortWidth}}>
            {port}
          </div>
          <div className="hosts-table--td" style={{width: ProtocolWidth}}>
            {protocol}
          </div>
          <div
            className="monotype hosts-table--td"
            style={{width: SourceWidth}}
          >
            {source || destination}
          </div>
          <div
            className="monotype hosts-table--td"
            style={{width: SecurityGroupsWidth}}
          >
            {security_groups}
          </div>
        </div>
      </>
    )
  }
}

export default TopologyDetailsSectionTableRow
