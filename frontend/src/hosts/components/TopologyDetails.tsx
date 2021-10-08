import _ from 'lodash'
import React, {PureComponent} from 'react'
import TopologyDetailsSection from 'src/hosts/components/TopologyDetailsSection'

// component
import {AWSInstanceData} from 'src/hosts/types/cloud'

interface Props {
  selectInstanceData: AWSInstanceData['instanceID']
  instanceTypeModal?: () => void
}

class TopologyDetails extends PureComponent<Props> {
  constructor(props: Props) {
    super(props)
  }

  public render() {
    const {selectInstanceData, instanceTypeModal} = this.props

    return (
      <>
        {_.isNull(selectInstanceData) || _.isUndefined(selectInstanceData) ? (
          <div
            className={'tab-pannel'}
            style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              height: '100%',
              fontSize: '18px',
            }}
          >
            <div>No Data</div>
          </div>
        ) : (
          <div className={'tab-pannel'}>
            {_.map(_.keys(selectInstanceData), d => {
              return (
                <TopologyDetailsSection
                  key={d}
                  title={d.replaceAll('_', ' ')}
                  selectInstanceData={selectInstanceData[d]}
                  instanceTypeModal={instanceTypeModal}
                />
              )
            })}
          </div>
        )}
      </>
    )
  }
}

export default TopologyDetails
