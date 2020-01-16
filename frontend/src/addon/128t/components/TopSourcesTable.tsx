// libraries
import React, {PureComponent} from 'react'
import _ from 'lodash'
import memoize from 'memoize-one'

// components
import TopSourcesTableRow from 'src/addon/128t/components/TopSourcesTableRow'
import FancyScrollbar from 'src/shared/components/FancyScrollbar'

// type
import {TopSource} from 'src/addon/128t/types'
import {ErrorHandling} from 'src/shared/decorators/errors'

// constants
import {TOPSOURCES_TABLE_SIZING} from 'src/addon/128t/constants'

enum SortDirection {
  ASC = 'asc',
  DESC = 'desc',
}

export interface Props {
  topSources: TopSource[]
}

interface State {
  searchTerm: string
  sortDirection: SortDirection
  sortKey: string
  topSourceCount: string
}

@ErrorHandling
class TopSourcesTable extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      searchTerm: '',
      sortDirection: SortDirection.ASC,
      sortKey: 'ip',
      topSourceCount: '0',
    }
  }

  public getSortedTopSources = memoize(
    (
      topSources: TopSource[],
      searchTerm: string,
      sortKey: string,
      sortDirection: SortDirection
    ) => this.sort(this.filter(topSources, searchTerm), sortKey, sortDirection)
  )

  public componentWillMount() {
    const {topSources} = this.props
    const {sortKey, sortDirection, searchTerm} = this.state

    this.setSourceCount(topSources, searchTerm, sortKey, sortDirection)
  }

  public componentWillReceiveProps(nextProps: Props) {
    const {topSources} = this.props
    if (topSources === nextProps.topSources) return
    const {sortKey, sortDirection, searchTerm} = this.state

    this.setSourceCount(
      nextProps.topSources,
      searchTerm,
      sortKey,
      sortDirection
    )
  }

  private setSourceCount(
    topSources: TopSource[],
    searchTerm: string,
    sortKey: string,
    sortDirection: SortDirection
  ) {
    const sortedTopSources = this.getSortedTopSources(
      topSources,
      searchTerm,
      sortKey,
      sortDirection
    )
    this.setState({topSourceCount: sortedTopSources.length})
  }

  public render() {
    return (
      <div className={`panel`}>
        <div className="panel-body">
          <div className="hosts-table">
            <div className="hosts-table--thead">
              <div className={'hosts-table--tr'}>{this.TableHeader}</div>
            </div>
            {this.TableData}
          </div>
        </div>
      </div>
    )
  }

  private get TableHeader() {
    const {
      IP,
      TENANT,
      CURRENTBANDWIDTH,
      TOTALDATA,
      SESSIONCOUNT,
    } = TOPSOURCES_TABLE_SIZING
    return (
      <>
        <div
          onClick={this.updateSort('ip')}
          className={this.sortableClasses('ip')}
          style={{width: IP}}
        >
          IP
          <span className="icon caret-up" />
        </div>
        <div
          onClick={this.updateSort('tenant')}
          className={this.sortableClasses('tenant')}
          style={{width: TENANT}}
        >
          Tenant
          <span className="icon caret-up" />
        </div>
        <div
          onClick={this.updateSort('sessioncnt')}
          className={this.sortableClasses('sessioncnt')}
          style={{width: SESSIONCOUNT}}
          title="Session Count"
        >
          S/C
          <span className="icon caret-up" />
        </div>
        <div
          onClick={this.updateSort('bandwidth')}
          className={this.sortableClasses('bandwidth')}
          style={{width: CURRENTBANDWIDTH}}
          title="Band Width"
        >
          B/W
          <span className="icon caret-up" />
        </div>
        <div
          onClick={this.updateSort('totaldata')}
          className={this.sortableClasses('totaldata')}
          style={{width: TOTALDATA}}
          title="Total Data"
        >
          T/D
          <span className="icon caret-up" />
        </div>
      </>
    )
  }

  // data add
  private get TableData() {
    const {topSources} = this.props
    const {sortKey, sortDirection, searchTerm} = this.state

    //const sortedTopSources = topSources
    const sortedTopSources = this.getSortedTopSources(
      topSources,
      searchTerm,
      sortKey,
      sortDirection
    )

    return (
      <FancyScrollbar
        children={sortedTopSources.map(
          (r: TopSource, i: number): JSX.Element => (
            <TopSourcesTableRow topSources={r} key={i} />
          )
        )}
      />
    )
  }

  public filter(alltopsources: TopSource[], searchTerm: string) {
    const filterText = searchTerm.toLowerCase()
    return alltopsources.filter(h => {
      return h.tenant.toLowerCase().includes(filterText)
    })
  }

  public sort(alltopsources: TopSource[], key: string, direction: string) {
    switch (direction) {
      case SortDirection.ASC:
        return _.sortBy(alltopsources, e => e[key])
      case SortDirection.DESC:
        return _.sortBy(alltopsources, e => e[key]).reverse()
      default:
        return alltopsources
    }
  }

  public updateSearchTerm = (searchTerm: string) => {
    this.setState({searchTerm})
  }

  public updateSort = (key: string) => () => {
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
}

export default TopSourcesTable
