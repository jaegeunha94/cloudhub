import React, {useEffect, useMemo, useState} from 'react'
import PredictionHexbin from './PredictionHexbin'
import {
  HexagonInputData,
  Links,
  NotificationAction,
  PredictionTooltipNode,
  Source,
} from 'src/types'
import {getLiveDeviceInfo} from '../apis'
import {generateForHosts} from 'src/utils/tempVars'
import {Auth} from 'src/types/reducers/auth'

import {bindActionCreators} from 'redux'
import {notify as notifyAction} from 'src/shared/actions/notifications'
import {connect} from 'react-redux'
import _ from 'lodash'
import {PredictionModal} from './PredictionModal'
import PredictionDashboardHeader from './PredictionDashboardHeader'
import {
  DEFAULT_CELL_BG_COLOR,
  DEFAULT_CELL_TEXT_COLOR,
} from 'src/dashboards/constants'
import LoadingDots from 'src/shared/components/LoadingDots'
import {CloudAutoRefresh} from 'src/clouds/types/type'
import {GlobalAutoRefresher} from 'src/utils/AutoRefresher'
import {setFilteredHexbin} from '../actions'
import {notifyPredictionHexbinGetFailed} from 'src/shared/copy/notifications'
import PredictionTooltipView from './PredictionTooltipView'
import {statusCal, statusHexColor} from '../utils'
import PredictionHexbinToggle from './PredictionHexbinToggle'

interface Props {
  source: Source
  links?: Links
  auth?: Auth
  notify?: NotificationAction
  cloudAutoRefresh?: CloudAutoRefresh
  setFilteredHexbin?: (value: string) => void
}

function PredictionHexbinWrapper({
  source,
  auth,
  cloudAutoRefresh,
  notify,
  setFilteredHexbin,
}: Props) {
  const getMlDlTagInit = () => {
    if (!!localStorage.getItem('hexbinTag')) {
      return JSON.parse(localStorage.getItem('hexbinTag'))
    } else {
      return {
        ml: false,
        dl: false,
        hexagon: true,
      }
    }
  }

  const [hostList, setHostList] = useState<PredictionTooltipNode[]>(null)

  const [error, setError] = useState<string>()

  const [loading, setLoading] = useState<boolean>(true)

  const [isMouseInComponent, setIsMouseInComponent] = useState(false)

  const [isPredictionModalOpen, setIsPredictionModalOpen] = useState(false)

  const [openNum, setOpenNum] = useState<string>(null)

  const [isHexbinDisplay, setIsHexbinDisplay] = useState(
    getMlDlTagInit().hexagon
  )

  const [isMlChartDisplay, setIsMlChartDisplay] = useState(getMlDlTagInit().ml)

  const [isDlChartDisplay, setIsDlChartDisplay] = useState(getMlDlTagInit().dl)

  let intervalID

  useEffect(() => {
    fetchDeviceInfo()
    return () => setFilteredHexbin('')
  }, [])

  useEffect(() => {
    GlobalAutoRefresher.poll(cloudAutoRefresh.prediction)
    const controller = new AbortController()

    if (!!cloudAutoRefresh.prediction) {
      clearInterval(intervalID)
      intervalID = window.setInterval(() => {
        fetchDeviceInfo()
      }, cloudAutoRefresh.prediction)
    }

    GlobalAutoRefresher.poll(cloudAutoRefresh.prediction)

    return () => {
      controller.abort()
      clearInterval(intervalID)
      intervalID = null
      GlobalAutoRefresher.stopPolling()
    }
  }, [cloudAutoRefresh])

  const inputData = useMemo<HexagonInputData[]>(() => {
    if (hostList === null) {
      return []
    }

    return hostList?.map(hex => {
      if (typeof hex.cpu === 'number' && typeof hex.memory === 'number') {
        return {
          statusColor: statusHexColor(statusCal((hex.cpu + hex.memory) / 2)),
          name: hex.name,
          cpu: Number(hex.cpu.toFixed()),
          memory: Number(hex.memory.toFixed()),
          traffic: hex.traffic,
          status: statusCal((hex.cpu + hex.memory) / 2),
        }
      } else {
        return {
          statusColor: statusHexColor('invalid'),
          name: hex.name,
          cpu: -1,
          memory: -1,
          traffic: hex.traffic,
          status: 'invalid',
        }
      }
    })
  }, [hostList])

  const onHexbinClick = (host: string, filteredHexbinHost?: string) => {
    // the way to close modal is hexbin double click

    if (filteredHexbinHost === host) {
      setIsPredictionModalOpen(false)
      setFilteredHexbin('')
    } else {
      setOpenNum(host)
      setIsPredictionModalOpen(true)
      setFilteredHexbin(host)
    }
  }

  //TODO: timerange var change to redux data not props
  const fetchDeviceInfo = async () => {
    const tempVars = generateForHosts(source)
    const meRole = _.get(auth, 'me.role', '')
    try {
      getLiveDeviceInfo(source.links.proxy, source.telegraf, tempVars, meRole)
        .then(resp => {
          setHostList(resp)
          setError(null)
          setLoading(false)
        })
        .catch(e => {
          setError('Hexbin Chart Error')
          setLoading(false)
          console.log(e)
        })
    } catch (e) {
      setError('Hexbin Chart Error')
      notify(notifyPredictionHexbinGetFailed(e))
      setLoading(false)
      console.log(e)
    }
  }

  const hexbinViewHandler = () => {
    setIsHexbinDisplay(!isHexbinDisplay)
    setTagLocalStorage(!isHexbinDisplay, 'hexagon')
  }

  const setTagLocalStorage = (value: boolean, key: 'ml' | 'dl' | 'hexagon') => {
    const result = {
      ml: isMlChartDisplay,
      dl: isDlChartDisplay,
      hexagon: isHexbinDisplay,
    }

    result[key] = value
    localStorage.setItem('hexbinTag', JSON.stringify(result))
  }

  return (
    <>
      <div
        onMouseOver={() => setIsMouseInComponent(true)}
        onMouseLeave={() => setIsMouseInComponent(false)}
        style={{height: '100%', backgroundColor: '#292933'}}
      >
        <PredictionDashboardHeader
          cellName={`Nodes Health Status`}
          cellBackgroundColor={DEFAULT_CELL_BG_COLOR}
          cellTextColor={DEFAULT_CELL_TEXT_COLOR}
          setModalOpen={setIsPredictionModalOpen}
        >
          {loading && (
            <LoadingDots
              className={'graph-panel__refreshing openstack-dots--loading'}
            />
          )}
          <PredictionHexbinToggle
            label={isHexbinDisplay ? 'Hexagon' : 'Card'}
            onChange={hexbinViewHandler}
            isActive={isHexbinDisplay}
            isLeft={true}
          />
          <PredictionHexbinToggle
            isActive={isDlChartDisplay}
            onChange={() => {
              setIsDlChartDisplay(!isDlChartDisplay)
              setTagLocalStorage(!isDlChartDisplay, 'dl')
            }}
            label="DL"
            // isHide={true}
          />
          <PredictionHexbinToggle
            isActive={isMlChartDisplay}
            onChange={() => {
              setIsMlChartDisplay(!isMlChartDisplay)
              setTagLocalStorage(!isMlChartDisplay, 'ml')
            }}
            label="ML"
          />
        </PredictionDashboardHeader>

        {!hostList || error ? (
          <div>{error}</div>
        ) : isHexbinDisplay ? (
          <PredictionHexbin
            onHexbinClick={onHexbinClick}
            inputData={inputData}
            isMouseInComponent={isMouseInComponent}
          />
        ) : (
          <PredictionTooltipView
            onHexbinClick={onHexbinClick}
            inputData={inputData}
          />
        )}
        {(isMlChartDisplay || isDlChartDisplay) && isPredictionModalOpen && (
          <PredictionModal
            isOpen={isPredictionModalOpen}
            setIsOpen={setIsPredictionModalOpen}
            isMl={isMlChartDisplay}
            isDl={isDlChartDisplay}
            host={openNum}
          />
        )}
      </div>
    </>
  )
}

const mstp = state => {
  const {
    app: {
      persisted: {cloudAutoRefresh},
      ephemeral: {inPresentationMode},
    },
    links,
    auth,
  } = state
  return {
    links,
    cloudAutoRefresh,
    inPresentationMode,
    auth,
  }
}

const mdtp = dispatch => ({
  notify: bindActionCreators(notifyAction, dispatch),
  setFilteredHexbin: bindActionCreators(setFilteredHexbin, dispatch),
})

export default connect(mstp, mdtp, null)(PredictionHexbinWrapper)
