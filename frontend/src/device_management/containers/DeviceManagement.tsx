// Library
import React, {PureComponent} from 'react'
import {connect} from 'react-redux'
import {bindActionCreators} from 'redux'
import _ from 'lodash'

// Components
import ImportDevicePage from 'src/device_management/components/ImportDevicePage'
import TableComponent from 'src/device_management/components/TableComponent'
import LoadingSpinner from 'src/reusable_ui/components/spinners/LoadingSpinner'
import DeviceConnection from 'src/device_management/components/DeviceConnection'

// Actions
import {notify as notifyAction} from 'src/shared/actions/notifications'
import {openShell} from 'src/shared/actions/shell'

// Constants
import Authorized, {EDITOR_ROLE} from 'src/auth/Authorized'
import {
  DEFAULT_NETWORK_DEVICE_DATA,
  columns,
} from 'src/device_management/constants'

// Type
import {
  Me,
  Notification,
  Organization,
  DeviceData,
  DeviceConnectionStatus,
  ShellInfo,
  AiModal,
  Source,
  Links,
  DeviceMonitoringStatus,
} from 'src/types'

// API
import {
  deleteDevice,
  fetchDeviceMonitoringStatus,
  getDeviceList,
} from 'src/device_management/apis'
import {getEnv} from 'src/shared/apis/env'

// Utils
import {convertDeviceDataOrganizationIDToName} from 'src/device_management/utils'
import {getDeep} from 'src/utils/wrappers'
import {generateForHosts} from 'src/utils/tempVars'

import {ErrorHandling} from 'src/shared/decorators/errors'
import {DELETE_MODAL_INFO, MONITORING_MODAL_INFO} from '../constants/deviceData'
import {closeModal, openModal} from 'src/shared/actions/aiModal'
import {ComponentColor} from 'src/reusable_ui'
import {DEVICE_INFO_SELECTED_MONITORING} from '../constants/deviceManagementColumn'
import {notifyFetchDeviceMonitoringStatusFailed} from 'src/shared/copy/notifications'

interface Auth {
  me: Me
}

interface Props {
  auth: Auth
  source: Source
  links: Links
  isUsingAuth: boolean
  me: Me
  organizations: Organization[]
  notify: (n: Notification) => void
  openShell: (shell: ShellInfo) => void
  openModal: (aiModal: AiModal) => void
  closeModal: () => void
}

interface State {
  isLoading: boolean
  data: DeviceData[]
  deviceConnectionVisibility: boolean
  deviceConnectionStatus: DeviceConnectionStatus
  deviceMonitoringStatus: DeviceMonitoringStatus
  importDeviceWizardVisibility: boolean
  deviceData: DeviceData[]
  selectedDeviceData: DeviceData
  checkedArray: string[]
}

@ErrorHandling
class DeviceManagement extends PureComponent<Props, State> {
  private isComponentMounted: boolean = true

  constructor(props: Props) {
    super(props)
    this.state = {
      data: [],
      isLoading: false,
      deviceConnectionVisibility: false,
      deviceConnectionStatus: 'None',
      deviceMonitoringStatus: {},
      importDeviceWizardVisibility: false,
      deviceData: [DEFAULT_NETWORK_DEVICE_DATA as DeviceData],
      selectedDeviceData: DEFAULT_NETWORK_DEVICE_DATA,
      checkedArray: [],
    }

    this.setState = (args, callback) => {
      if (!this.isComponentMounted) return
      PureComponent.prototype.setState.bind(this)(args, callback)
    }

    this.connectDevice = this.connectDevice.bind(this)
    this.handleRowClick = this.handleRowClick.bind(this)
  }

  public componentDidMount(): void {
    try {
      this.getDeviceAJAX()
      this.fetchDeviceMonitoringStatus()
    } catch (error) {
      console.error(error)
      throw error
    }
  }

  public componentWillUnmount() {
    this.isComponentMounted = false
  }

  public render() {
    const {me, organizations, isUsingAuth} = this.props
    const {
      data,
      deviceMonitoringStatus,
      deviceConnectionVisibility,
      deviceConnectionStatus,
      importDeviceWizardVisibility,
      selectedDeviceData,
    } = this.state
    const updatedDeviceData = this.getDeviceMonitoringStatus(
      data,
      deviceMonitoringStatus
    )

    return (
      <>
        <TableComponent
          tableTitle={`${
            this.state.data.length
              ? this.state.data.length === 1
                ? '1 Device'
                : this.state.data.length + ' ' + 'Devices'
              : '0 Device'
          } list`}
          data={updatedDeviceData}
          columns={this.column}
          checkedArray={this.state.checkedArray}
          setCheckedArray={(value: string[]) =>
            this.setState({checkedArray: value})
          }
          // options={this.options}
          topLeftRender={
            <div className="device-management-top left">
              <div className="space-x">
                <Authorized requiredRole={EDITOR_ROLE}>
                  <button
                    onClick={() => {
                      this.onClickDelete(this.state.checkedArray)
                    }}
                    className="btn button btn-sm btn-primary"
                    disabled={this.state.checkedArray.length === 0}
                  >
                    <span className="icon trash" /> Delete Device
                  </button>
                </Authorized>
                {/* TODO Consder requiredRole */}
                <Authorized requiredRole={EDITOR_ROLE}>
                  <button
                    onClick={() => {
                      this.onClickMonitoring(this.state.checkedArray)
                    }}
                    className="btn button btn-sm btn-primary"
                    disabled={this.state.checkedArray.length === 0}
                  >
                    <span className="icon import" /> Apply Monitoring
                  </button>
                </Authorized>
              </div>
              <div className="space-x">
                <Authorized requiredRole={EDITOR_ROLE}>
                  <div className="btn button btn-sm btn-primary">
                    <span className="icon computer-desktop" /> Learning Setting
                  </div>
                </Authorized>

                <Authorized requiredRole={EDITOR_ROLE}>
                  <div
                    onClick={this.connectDevice('Creating')}
                    className="btn button btn-sm btn-primary"
                  >
                    <span className="icon plus" /> Add Device
                  </div>
                </Authorized>
                {/* TODO Consder requiredRole */}
                <Authorized requiredRole={EDITOR_ROLE}>
                  <div
                    onClick={this.importDevice}
                    className="btn button btn-sm btn-primary"
                  >
                    <span className="icon import" /> Import Device
                  </div>
                </Authorized>
              </div>
            </div>
          }
        />

        <DeviceConnection
          deviceConnectionStatus={deviceConnectionStatus}
          isVisible={deviceConnectionVisibility}
          notify={this.props.notify}
          me={me}
          organizations={organizations}
          selectedDeviceData={selectedDeviceData}
          isUsingAuth={isUsingAuth}
          toggleVisibility={this.handleToggleDeviceConnectionModal}
        />
        <ImportDevicePage
          isVisible={importDeviceWizardVisibility}
          organizations={organizations}
          onDismissOverlay={this.handleDismissImportDeviceModalOverlay}
          notify={this.props.notify}
        />

        {this.state.isLoading && (
          <div className="loading-box">
            <LoadingSpinner />
          </div>
        )}
      </>
    )
  }

  private getDeviceAJAX = async () => {
    const {organizations} = this.props
    const {data} = await getDeviceList()
    const convertedDeviceData = convertDeviceDataOrganizationIDToName(
      data.Devices,
      organizations
    ) as DeviceData[]

    this.setState({data: convertedDeviceData})
  }

  private fetchDeviceMonitoringStatus = async () => {
    try {
      const {source, links} = this.props
      const envVars = await getEnv(links.environment)
      const telegrafSystemInterval = getDeep<string>(
        envVars,
        'telegrafSystemInterval',
        ''
      )
      const tempVars = generateForHosts(source)
      const deviceMonitoringStatus = await fetchDeviceMonitoringStatus(
        source.links.proxy,
        source.telegraf,
        telegrafSystemInterval,
        tempVars
      )

      this.setState({
        deviceMonitoringStatus,
      })
    } catch (error) {
      this.props.notify(notifyFetchDeviceMonitoringStatusFailed(error.message))
    }
  }

  private getDeviceMonitoringStatus(
    devicesData: DeviceData[],
    deviceMonitoringStatus: DeviceMonitoringStatus
  ) {
    return devicesData.map(device => {
      const {device_ip} = device
      const uptime = deviceMonitoringStatus?.[device_ip]?.uptime || 0
      const isMonitoring = uptime !== 0
      return {
        ...device,
        isMonitoring,
      }
    })
  }

  private onClickShellModalOpen = (shell: ShellInfo) => {
    this.props.openShell(shell)
  }

  private handleRowClick = selectedDeviceData => {
    this.connectDevice('Updating')()
    this.setState({selectedDeviceData: selectedDeviceData})
  }

  private column = columns({
    onEditClick: this.handleRowClick,
    onConsoleClick: this.onClickShellModalOpen,
  })

  private onClickDelete = (checkedArray: string[]) => {
    this.props.openModal({
      title: 'Delete Device',
      isVisible: true,
      message: DELETE_MODAL_INFO.message,
      confirmText: 'Delete',
      btnColor: ComponentColor.Danger,
      onConfirm: () => {
        this.deleteDevicesAJAX(checkedArray)
        this.props.closeModal()
      },
      cancelText: 'Cancel',
    })
  }

  private deleteDevicesAJAX = async (idList: string[]) => {
    const numIdList = idList.map(i => Number(i))
    this.setState({isLoading: true})
    await deleteDevice({devices_id: numIdList})

    this.getDeviceAJAX()
    this.setState({checkedArray: [], isLoading: false})
  }

  private onClickMonitoring = (checkedArray: string[]) => {
    const validArray = this.state.data.filter(
      i => checkedArray.includes(`${i.id}`)
      //create monitoring sql is_modeling_generated=> is_monitoring
    )

    this.props.openModal({
      isVisible: true,
      // message: MONITORING_MODAL_INFO.workHeader,
      message: '',
      btnColor: ComponentColor.Warning,
      onConfirm: () => {
        this.props.closeModal()
      },
      confirmText: 'Confirm',
      cancelText: 'Cancel',
      childNode: (
        <div className="device-modal--childNode">
          <TableComponent
            data={validArray}
            tableTitle="Selected Device List"
            columns={DEVICE_INFO_SELECTED_MONITORING}
            isSearchDisplay={false}
          />

          {MONITORING_MODAL_INFO.workMessage}
        </div>
      ),
    })

    //Apply Monitoring Process -> close Modal
  }

  private connectDevice = (
    deviceConnectionStatus?: DeviceConnectionStatus
  ) => () => {
    this.setState({
      deviceConnectionVisibility: true,
      deviceConnectionStatus: deviceConnectionStatus,
    })
  }

  private handleToggleDeviceConnectionModal = deviceConnectionVisibility => () => {
    this.getDeviceAJAX()
    this.setState({
      deviceConnectionVisibility: deviceConnectionVisibility,
      deviceConnectionStatus: deviceConnectionVisibility
        ? this.state.deviceConnectionStatus
        : 'None',
    })
  }

  private importDevice = () => {
    this.setState({
      importDeviceWizardVisibility: true,
    })
  }

  private handleDismissImportDeviceModalOverlay = (): void => {
    this.getDeviceAJAX()
    this.setState({
      importDeviceWizardVisibility: false,
    })
  }
}

const mstp = ({adminCloudHub: {organizations}, auth, links}) => ({
  organizations,
  isUsingAuth: auth.isUsingAuth,
  auth,
  me: auth.me,
  links,
})

const mdtp = (dispatch: any) => ({
  notify: bindActionCreators(notifyAction, dispatch),
  openShell: bindActionCreators(openShell, dispatch),
  openModal: bindActionCreators(openModal, dispatch),
  closeModal: bindActionCreators(closeModal, dispatch),
})

export default connect(mstp, mdtp, null)(DeviceManagement)
