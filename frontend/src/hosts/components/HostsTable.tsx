import React, {PureComponent} from 'react'
import _ from 'lodash'
import memoize from 'memoize-one'

import SearchBar from 'src/hosts/components/SearchBar'
import HostRow from 'src/hosts/components/HostRow'
import InfiniteScroll from 'src/shared/components/InfiniteScroll'
import PageSpinner from 'src/shared/components/PageSpinner'

import {
  CLOUD_HOSTS_TABLE_SIZING,
  HOSTS_TABLE_SIZING,
} from 'src/hosts/constants/tableSizing'
import {ErrorHandling} from 'src/shared/decorators/errors'
import {Source, RemoteDataState, Host} from 'src/types'

import {HostsPage} from 'src/hosts/containers/HostsPage'

//middlware
import {
  setLocalStorage,
  getLocalStorage,
  verifyLocalStorage,
} from 'src/shared/middleware/localStorage'
import CloudHostRow from './CloudHostRow'

enum SortDirection {
  ASC = 'asc',
  DESC = 'desc',
}

export interface Props {
  hosts: Host[]
  hostsPageStatus: RemoteDataState
  source: Source
  focusedHost: string
  onClickTableRow: HostsPage['handleClickTableRow']
}

interface State {
  searchTerm: string
  sortDirection: SortDirection
  sortKey: string
  selected: string
  items: string[]
  activeEditorTab: string
}

@ErrorHandling
class HostsTable extends PureComponent<Props, State> {
  public getSortedHosts = memoize(
    (
      hosts,
      searchTerm: string,
      sortKey: string,
      sortDirection: SortDirection
    ) => {
      return this.sort(this.filter(hosts, searchTerm), sortKey, sortDirection)
    }
  )

  constructor(props: Props) {
    super(props)

    this.state = {
      searchTerm: '',
      sortDirection: SortDirection.ASC,
      sortKey: 'name',
      selected: 'Select Region',
      items: ['CloudWatch', '2', '3'],
      activeEditorTab: 'snet',
    }
  }

  public filter(allHosts: Host[], searchTerm: string): Host[] {
    const filterText = searchTerm.toLowerCase()
    return allHosts.filter(h => {
      const apps = h.apps ? h.apps.join(', ') : ''

      let tagResult = false
      if (h.tags) {
        tagResult = Object.keys(h.tags).reduce((acc, key) => {
          return acc || h.tags[key].toLowerCase().includes(filterText)
        }, false)
      } else {
        tagResult = false
      }
      return (
        h.name.toLowerCase().includes(filterText) ||
        apps.toLowerCase().includes(filterText) ||
        tagResult
      )
    })
  }

  public sort(hosts: Host[], key: string, direction: SortDirection): Host[] {
    switch (direction) {
      case SortDirection.ASC:
        return _.sortBy<Host>(hosts, e => e[key])
      case SortDirection.DESC:
        return _.sortBy<Host>(hosts, e => e[key]).reverse()
      default:
        return hosts
    }
  }

  public updateSearchTerm = (searchTerm: string): void => {
    this.setState({searchTerm})
  }

  public updateSort = (key: string) => (): void => {
    const {sortKey, sortDirection} = this.state
    if (sortKey === key) {
      const reverseDirection =
        sortDirection === SortDirection.ASC
          ? SortDirection.DESC
          : SortDirection.ASC
      this.setState({sortDirection: reverseDirection})
    } else {
      this.setState({sortKey: key, sortDirection: SortDirection.ASC})
    }
  }

  public sortableClasses = (key: string): string => {
    const {sortKey, sortDirection} = this.state
    if (sortKey === key) {
      if (sortDirection === SortDirection.ASC) {
        return 'hosts-table--th sortable-header sorting-ascending'
      }
      return 'hosts-table--th sortable-header sorting-descending'
    }
    return 'hosts-table--th sortable-header'
  }

  public componentWillMount() {
    verifyLocalStorage(getLocalStorage, setLocalStorage, 'hostsTableState', {
      sortKey: this.state.sortKey,
      sortDirection: this.state.sortDirection,
      focusedHost: this.props.focusedHost,
    })

    const {sortKey, sortDirection} = getLocalStorage('hostsTableState')

    this.setState({
      sortKey,
      sortDirection,
    })
  }

  public componentDidUpdate() {
    setLocalStorage('hostsTableState', {
      sortKey: this.state.sortKey,
      sortDirection: this.state.sortDirection,
      focusedHost: this.props.focusedHost,
    })
  }

  public render() {
    return (
      <div className="panel">
        <div className="panel-heading">
          <div>
            <h2 className="panel-title">{this.HostsTitle}</h2>
          </div>

          <div>
            <SearchBar
              placeholder="Filter by Host..."
              onSearch={this.updateSearchTerm}
            />
          </div>
        </div>
        <div className="panel-body">
          {this.state.activeEditorTab === 'snet' ? this.TableContents : null}
          {this.state.activeEditorTab === 'aws'
            ? this.CloudTableContents
            : null}
        </div>
      </div>
    )
  }

  private get TableContents(): JSX.Element {
    const {hosts, hostsPageStatus} = this.props
    const {sortKey, sortDirection, searchTerm} = this.state
    const sortedHosts = this.getSortedHosts(
      hosts,
      searchTerm,
      sortKey,
      sortDirection
    )
    if (hostsPageStatus === RemoteDataState.Loading) {
      return this.LoadingState
    }
    if (hostsPageStatus === RemoteDataState.Error) {
      return this.ErrorState
    }
    if (hosts.length === 0) {
      return this.NoHostsState
    }
    if (sortedHosts.length === 0) {
      return this.NoSortedHostsState
    }
    return this.TableWithHosts
  }

  private get CloudTableContents(): JSX.Element {
    const {hosts, hostsPageStatus} = this.props
    const {sortKey, sortDirection, searchTerm} = this.state
    const sortedHosts = this.getSortedHosts(
      hosts,
      searchTerm,
      sortKey,
      sortDirection
    )
    if (hostsPageStatus === RemoteDataState.Loading) {
      return this.LoadingState
    }
    if (hostsPageStatus === RemoteDataState.Error) {
      return this.ErrorState
    }
    if (hosts.length === 0) {
      return this.NoHostsState
    }
    if (sortedHosts.length === 0) {
      return this.NoSortedHostsState
    }
    return this.CloudTableWithHosts
  }

  private get TableWithHosts(): JSX.Element {
    const {source, hosts, focusedHost, onClickTableRow} = this.props
    const {sortKey, sortDirection, searchTerm} = this.state
    const sortedHosts = this.getSortedHosts(
      hosts,
      searchTerm,
      sortKey,
      sortDirection
    )

    return (
      <div className="hosts-table">
        {this.HostsTableHeader}
        <InfiniteScroll
          items={sortedHosts.map(h => (
            <HostRow
              key={h.name}
              host={h}
              sourceID={source.id}
              focusedHost={focusedHost}
              onClickTableRow={onClickTableRow}
            />
          ))}
          itemHeight={26}
          className="hosts-table--tbody"
        />
      </div>
    )
  }

  private get CloudTableWithHosts(): JSX.Element {
    const {source, hosts, focusedHost, onClickTableRow} = this.props
    const {sortKey, sortDirection, searchTerm} = this.state
    const sortedHosts = this.getSortedHosts(
      hosts,
      searchTerm,
      sortKey,
      sortDirection
    )

    return (
      <div className="hosts-table">
        {this.CloudHostsTableHeader}
        <InfiniteScroll
          items={sortedHosts.map(h => (
            <CloudHostRow
              key={h.name}
              host={h}
              sourceID={source.id}
              focusedHost={focusedHost}
              onClickTableRow={onClickTableRow}
            />
          ))}
          itemHeight={26}
          className="hosts-table--tbody"
        />
      </div>
    )
  }

  private get LoadingState(): JSX.Element {
    return <PageSpinner />
  }

  private get ErrorState(): JSX.Element {
    return (
      <div className="generic-empty-state">
        <h4 style={{margin: '90px 0'}}>There was a problem loading hosts</h4>
      </div>
    )
  }

  private get NoHostsState(): JSX.Element {
    return (
      <div className="generic-empty-state">
        <h4 style={{margin: '90px 0'}}>No Hosts found</h4>
      </div>
    )
  }

  private get NoSortedHostsState(): JSX.Element {
    return (
      <div className="generic-empty-state">
        <h4 style={{margin: '90px 0'}}>
          There are no hosts that match the search criteria
        </h4>
      </div>
    )
  }

  private get HostsTitle(): string {
    const {hostsPageStatus, hosts} = this.props
    const {sortKey, sortDirection, searchTerm} = this.state
    const sortedHosts = this.getSortedHosts(
      hosts,
      searchTerm,
      sortKey,
      sortDirection
    )
    const hostsCount = sortedHosts.length
    if (hostsPageStatus === RemoteDataState.Loading) {
      return 'Loading Hosts...'
    }
    if (hostsCount === 1) {
      return `1 Host`
    }
    return `${hostsCount} Hosts`
  }
  private get CloudHostsTableHeader(): JSX.Element {
    const {
      CloudNameWidth,
      CloudInstanceIDWidth,
      CloudInstanceStateWidth,
      CloudInstanceTypeWidth,
      CloudStatusCheckWidth,
      CloudAlarmStatusWidth,
      CloudAppsWidth,
      CloudCPUWidth,
      CloudMemoryWidth,
      CloudDiskWidth,
    } = CLOUD_HOSTS_TABLE_SIZING

    return (
      <div className="hosts-table--thead">
        <div className="hosts-table--tr">
          <div
            onClick={this.updateSort('name')}
            className={this.sortableClasses('name')}
            style={{width: CloudNameWidth}}
          >
            Name
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('instanceID')}
            className={this.sortableClasses('instanceID')}
            style={{width: CloudInstanceIDWidth}}
          >
            Instance ID
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('instanceState')}
            className={this.sortableClasses('instanceState')}
            style={{width: CloudInstanceStateWidth}}
          >
            Instance state
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('instanceType')}
            className={this.sortableClasses('instanceType')}
            style={{width: CloudInstanceTypeWidth}}
          >
            Instance type
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('statusCheck')}
            className={this.sortableClasses('statusCheck')}
            style={{width: CloudStatusCheckWidth}}
          >
            Status check
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('alarmStatus')}
            className={this.sortableClasses('alarmStatus')}
            style={{width: CloudAlarmStatusWidth}}
          >
            Alarm status
            <span className="icon caret-up" />
          </div>
          <div
            className="hosts-table--th list-type"
            style={{width: CloudAppsWidth}}
          >
            Apps
          </div>
          <div
            onClick={this.updateSort('cpu')}
            className={this.sortableClasses('cpu')}
            style={{width: CloudCPUWidth}}
          >
            CPU
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('memory')}
            className={this.sortableClasses('memory')}
            style={{width: CloudMemoryWidth}}
          >
            Memory
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('disk')}
            className={this.sortableClasses('disk')}
            style={{width: CloudDiskWidth}}
          >
            Disk
            <span className="icon caret-up" />
          </div>
        </div>
      </div>
    )
  }

  private get HostsTableHeader(): JSX.Element {
    const {NameWidth, StatusWidth, CPUWidth, LoadWidth} = HOSTS_TABLE_SIZING

    return (
      <div className="hosts-table--thead">
        <div className="hosts-table--tr">
          <div
            onClick={this.updateSort('name')}
            className={this.sortableClasses('name')}
            style={{width: NameWidth}}
          >
            Host
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('deltaUptime')}
            className={this.sortableClasses('deltaUptime')}
            style={{width: StatusWidth}}
          >
            Status
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('cpu')}
            className={this.sortableClasses('cpu')}
            style={{width: CPUWidth}}
          >
            CPU
            <span className="icon caret-up" />
          </div>
          <div
            onClick={this.updateSort('load')}
            className={this.sortableClasses('load')}
            style={{width: LoadWidth}}
          >
            Load
            <span className="icon caret-up" />
          </div>
          <div className="hosts-table--th list-type">Apps</div>
        </div>
      </div>
    )
  }
}

export default HostsTable
