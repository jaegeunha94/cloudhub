import React, {Component} from 'react'
import PropTypes, {object} from 'prop-types'
import {connect} from 'react-redux'
import {bindActionCreators} from 'redux'

import flatten from 'lodash/flatten'
import uniqBy from 'lodash/uniqBy'

import {showDatabases, showQueries} from 'shared/apis/metaQuery'

import QueriesTable from 'src/admin/components/QueriesTable'
import showDatabasesParser from 'shared/parsing/showDatabases'
import showQueriesParser from 'shared/parsing/showQueries'
import {notifyQueriesError} from 'shared/copy/notifications'
import {ErrorHandling} from 'src/shared/decorators/errors'
import AutoRefreshDropdown from 'src/shared/components/dropdown_auto_refresh/AutoRefreshDropdown'

import {
  loadQueries as loadQueriesAction,
  setQueryToKill as setQueryToKillAction,
  setQueriesSort as setQueriesSortAction,
  killQueryAsync,
} from 'src/admin/actions/influxdb'
import {eachRoleQueries} from 'src/admin/utils/eachRoleWorker'

import {notify as notifyAction} from 'shared/actions/notifications'
import queries from 'src/data_explorer/reducers/queries'
import {setCloudAutoRefresh} from 'src/clouds/actions'

class QueriesPage extends Component {
  constructor(props) {
    super(props)
    this.state = {
      updateInterval: props.cloudAutoRefresh.queries || 5000,
    }
  }
  componentDidMount() {
    this.updateQueries()
    if (this.state.updateInterval > 0) {
      this.intervalID = setInterval(
        this.updateQueries,
        this.state.updateInterval
      )
    }
  }

  componentWillUnmount() {
    if (this.intervalID) {
      clearInterval(this.intervalID)
    }
  }

  render() {
    const {queries, queriesSort, changeSort} = this.props
    const {updateInterval, title} = this.state

    return (
      <div className="panel panel-solid">
        <div className="panel-heading">
          <h2 className="panel-title">{title}</h2>
          <div style={{float: 'right'}}>
            <AutoRefreshDropdown
              selected={updateInterval}
              onChoose={this.changeRefreshInterval}
              onManualRefresh={this.updateQueries}
              showManualRefresh={true}
            />
          </div>
        </div>
        <div className="panel-body">
          <QueriesTable
            queries={queries}
            queriesSort={queriesSort}
            changeSort={changeSort}
            onKillQuery={this.handleKillQuery}
          />
        </div>
      </div>
    )
  }

  updateQueries = () => {
    const {source, notify, loadQueries, auth, cloudAutoRefresh} = this.props
    showDatabases(source.links.proxy).then(resp => {
      const {databases, errors} = showDatabasesParser(resp.data)
      const checkDatabases = eachRoleQueries(databases, auth)

      if (errors.length) {
        this.setState(state => ({...state, title: ''}))
        errors.forEach(message => notify(notifyQueriesError(message)))
        return
      }
      this.setState(state => ({
        ...state,
        title:
          databases.length === 1
            ? '1 Database'
            : `${databases.length} Databases`,
      }))

      const fetches = checkDatabases.map(db =>
        showQueries(source.links.proxy, db)
      )

      Promise.allSettled(fetches).then(results => {
        const allQueries = []
        results.forEach((settledResponse, i) => {
          if (!settledResponse.value) {
            console.error(
              `Unable to show queries on '${databases[i]}': `,
              settledResponse.reason
            )
            return
          }
          const result = showQueriesParser(settledResponse.value.data)
          if (result.errors.length) {
            result.errors.forEach(message =>
              notify(notifyQueriesError(message))
            )
          }

          allQueries.push(...result.queries)
        })

        const queries = uniqBy(flatten(allQueries), q => q.id)
        loadQueries(queries)
      })
    })
  }
  changeRefreshInterval = ({milliseconds: updateInterval}) => {
    this.props.onChooseCloudAutoRefresh({queries: updateInterval})
    this.setState(state => ({...state, updateInterval}))
    if (this.intervalID) {
      clearInterval(this.intervalID)
      this.intervalID = undefined
    }
    if (updateInterval > 0) {
      this.updateQueries()
      this.intervalID = setInterval(this.updateQueries, updateInterval)
    }
  }

  handleKillQuery = query => {
    const {source, killQuery} = this.props
    killQuery(source.links.proxy, query)
  }
}

const {
  arrayOf,
  func,
  string,
  number,
  shape,
  oneOfType,
  objectOf,
  oneOf,
} = PropTypes

QueriesPage.propTypes = {
  auth: shape().isRequired,
  source: shape({
    links: shape({
      proxy: string,
    }),
  }),
  queries: arrayOf(shape()),
  queriesSort: string,
  loadQueries: func,
  queryIDToKill: string,
  setQueryToKill: func,
  changeSort: func,
  killQuery: func,
  notify: func.isRequired,
  cloudAutoRefresh: oneOfType([objectOf(number), oneOf([null])]),
  onChooseCloudAutoRefresh: func,
}

const mapStateToProps = ({
  adminInfluxDB: {queries, queriesSort, queryIDToKill},
  auth,
  app: {
    persisted: {cloudAutoRefresh},
  },
}) => ({
  queries,
  queriesSort,
  queryIDToKill,
  auth,
  cloudAutoRefresh,
})

const mapDispatchToProps = dispatch => ({
  loadQueries: bindActionCreators(loadQueriesAction, dispatch),
  setQueryToKill: bindActionCreators(setQueryToKillAction, dispatch),
  changeSort: bindActionCreators(setQueriesSortAction, dispatch),
  killQuery: bindActionCreators(killQueryAsync, dispatch),
  notify: bindActionCreators(notifyAction, dispatch),
  onChooseCloudAutoRefresh: bindActionCreators(setCloudAutoRefresh, dispatch),
})

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(ErrorHandling(QueriesPage))
