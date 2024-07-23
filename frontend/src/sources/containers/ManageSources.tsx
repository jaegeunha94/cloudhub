import React, {PureComponent} from 'react'
import {connect} from 'react-redux'
import {bindActionCreators} from 'redux'
import {ErrorHandling} from 'src/shared/decorators/errors'

import * as sourcesActions from 'src/shared/actions/sources'
import {notify as notifyAction} from 'src/shared/actions/notifications'
import {UserRole, ForceSessionAbortInputRole} from 'src/shared/actions/session'
import {SUPERADMIN_ROLE} from 'src/auth/Authorized'

import {Page} from 'src/reusable_ui'
import InfluxTable from 'src/sources/components/InfluxTable'
import ConnectionWizard from 'src/sources/components/ConnectionWizard'
import {connectedSourceAction, connectedSource} from 'src/sources/actions'

import {
  notifySourceDeleted,
  notifySourceDeleteFailed,
} from 'src/shared/copy/notifications'

import {Me, Source, Notification, Organization} from 'src/types'
import {ToggleWizard} from 'src/types/wizard'

interface State {
  wizardVisibility: boolean
  sourceInWizard: Source
  jumpStep: number
  showNewKapacitor: boolean
}

interface Props {
  source: Source
  sources: Source[]
  me: Me
  organizations: Organization[]
  isUsingAuth: boolean
  notify: (n: Notification) => void
  deleteKapacitor: sourcesActions.DeleteKapacitor
  fetchKapacitors: sourcesActions.FetchKapacitorsAsync
  removeAndLoadSources: sourcesActions.RemoveAndLoadSources
  setActiveKapacitor: sourcesActions.SetActiveKapacitorAsync
  connectedSource: connectedSourceAction
  ForceSessionAbortInputRole: (
    requireRole: UserRole,
    isNoAuthOuting?: boolean
  ) => void
}

const VERSION = process.env.npm_package_version

@ErrorHandling
class ManageSources extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      wizardVisibility: false,
      sourceInWizard: null,
      jumpStep: null,
      showNewKapacitor: null,
    }
  }

  public async componentDidMount() {
    const {ForceSessionAbortInputRole} = this.props

    ForceSessionAbortInputRole(SUPERADMIN_ROLE)
    this.fetchKapacitors()
  }

  public componentDidUpdate(prevProps: Props) {
    if (prevProps.sources.length !== this.props.sources.length) {
      this.fetchKapacitors()
    }
  }

  public render() {
    const {
      me,
      organizations,
      isUsingAuth,
      sources,
      source,
      deleteKapacitor,
      connectedSource,
    } = this.props
    const {
      wizardVisibility,
      sourceInWizard,
      jumpStep,
      showNewKapacitor,
    } = this.state
    return (
      <Page>
        <Page.Header>
          <Page.Header.Left>
            <Page.Title title="Configuration" />
          </Page.Header.Left>
          <Page.Header.Right showSourceIndicator={true} />
        </Page.Header>
        <Page.Contents>
          <InfluxTable
            source={source}
            sources={sources}
            deleteKapacitor={deleteKapacitor}
            onDeleteSource={this.handleDeleteSource}
            setActiveKapacitor={this.handleSetActiveKapacitor}
            toggleWizard={this.toggleWizard}
            connectedSource={connectedSource}
          />
          <p className="version-number">CloudHub Version: {VERSION}</p>
        </Page.Contents>
        <ConnectionWizard
          me={me}
          organizations={organizations}
          isUsingAuth={isUsingAuth}
          isVisible={wizardVisibility}
          toggleVisibility={this.toggleWizard}
          source={sourceInWizard}
          jumpStep={jumpStep}
          showNewKapacitor={showNewKapacitor}
        />
      </Page>
    )
  }

  private handleDeleteSource = (source: Source) => {
    const {notify} = this.props

    try {
      this.props.removeAndLoadSources(source)
      notify(notifySourceDeleted(source.name))
    } catch (e) {
      notify(notifySourceDeleteFailed(source.name))
    }
  }

  private fetchKapacitors = () => {
    this.props.sources.forEach(source => {
      this.props.fetchKapacitors(source)
    })
  }

  private toggleWizard: ToggleWizard = (
    isVisible,
    source = null,
    jumpStep = null,
    showNewKapacitor = null
  ) => () => {
    if (!isVisible) {
      this.fetchKapacitors()
    }
    this.setState({
      wizardVisibility: isVisible,
      sourceInWizard: source,
      jumpStep,
      showNewKapacitor,
    })
  }

  private handleSetActiveKapacitor = kapacitor => {
    this.props.setActiveKapacitor(kapacitor)
  }
}

const mstp = ({
  adminCloudHub: {organizations},
  auth: {isUsingAuth, me},
  sources,
}) => ({
  organizations,
  isUsingAuth,
  me,
  sources,
})

const mdtp = (dispatch: any) => ({
  notify: bindActionCreators(notifyAction, dispatch),
  removeAndLoadSources: bindActionCreators(
    sourcesActions.removeAndLoadSources,
    dispatch
  ),
  fetchKapacitors: bindActionCreators(
    sourcesActions.fetchKapacitorsAsync,
    dispatch
  ),
  setActiveKapacitor: bindActionCreators(
    sourcesActions.setActiveKapacitorAsync,
    dispatch
  ),
  deleteKapacitor: bindActionCreators(
    sourcesActions.deleteKapacitorAsync,
    dispatch
  ),
  connectedSource: bindActionCreators(connectedSource, dispatch),
  ForceSessionAbortInputRole: bindActionCreators(
    ForceSessionAbortInputRole,
    dispatch
  ),
})

export default connect(mstp, mdtp, null)(ManageSources)
