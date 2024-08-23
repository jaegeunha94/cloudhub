import {
  defaultRuleConfigs,
  DEFAULT_RULE_ID,
  HANDLERS_TO_RULE_THEM_ALL,
} from 'src/kapacitor/constants'
import _ from 'lodash'

export default function rules(state = {}, action) {
  switch (action.type) {
    case 'LOAD_DEFAULT_RULE': {
      const {queryID} = action.payload
      return Object.assign({}, state, {
        [DEFAULT_RULE_ID]: {
          id: DEFAULT_RULE_ID,
          queryID,
          trigger: 'threshold',
          values: defaultRuleConfigs.threshold,
          message: '',
          alertNodes: {},
          every: null,
          name: 'Untitled Rule',
          status: 'enabled',
        },
      })
    }

    case 'LOAD_RULES': {
      const theRules = action.payload.rules
      const ruleIDs = theRules.map(r => r.id)
      return _.zipObject(ruleIDs, theRules)
    }

    case 'LOAD_RULE': {
      const {rule} = action.payload
      return Object.assign({}, state, {
        [rule.id]: rule,
      })
    }

    case 'CHOOSE_TRIGGER': {
      const trigger = action.payload.trigger
      const ruleID = action.payload.ruleID
      return Object.assign({}, state, {
        [ruleID]: Object.assign({}, state[ruleID], {
          trigger: trigger.toLowerCase(),
          values: defaultRuleConfigs[trigger.toLowerCase()],
        }),
      })
    }

    case 'ADD_EVERY': {
      const {ruleID, frequency} = action.payload
      return {...state, [ruleID]: {...state[ruleID], every: frequency}}
    }

    case 'REMOVE_EVERY': {
      const {ruleID} = action.payload
      return {...state, [ruleID]: {...state[ruleID], every: null}}
    }

    case 'UPDATE_RULE_VALUES': {
      const {ruleID, trigger, values} = action.payload
      return Object.assign({}, state, {
        [ruleID]: Object.assign({}, state[ruleID], {
          trigger: trigger.toLowerCase(),
          values,
        }),
      })
    }

    case 'UPDATE_RULE_MESSAGE': {
      const {ruleID, message} = action.payload
      return Object.assign({}, state, {
        [ruleID]: Object.assign({}, state[ruleID], {
          message,
        }),
      })
    }

    case 'UPDATE_RULE_ALERT_NODES': {
      const {ruleID, alerts} = action.payload
      const alertNodesByType = {}
      _.forEach(alerts, h => {
        if (h.enabled) {
          if (h.type === 'post') {
            if (h.url === '') {
              return
            }
            h.headers = {[h.headerKey]: h.headerValue}
          }
          if (h.type === 'log' && h.filePath === '') {
            return
          }
          if (h.type === 'tcp' && h.address === '') {
            return
          }
          if (h.type === 'exec' && h.command.length === 0) {
            return
          }
          const existing = _.get(alertNodesByType, h.type, [])
          alertNodesByType[h.type] = [
            ...existing,
            _.pick(h, HANDLERS_TO_RULE_THEM_ALL[h.type]),
          ]
        }
      })
      return Object.assign({}, state, {
        [ruleID]: Object.assign({}, state[ruleID], {
          alertNodes: alertNodesByType,
        }),
      })
    }

    case 'UPDATE_RULE_NAME': {
      const {ruleID, name} = action.payload
      return Object.assign({}, state, {
        [ruleID]: Object.assign({}, state[ruleID], {
          name,
        }),
      })
    }

    case 'DELETE_RULE_SUCCESS': {
      const {ruleID} = action.payload
      delete state[ruleID]
      return Object.assign({}, state)
    }

    case 'UPDATE_RULE_DETAILS': {
      const {ruleID, details} = action.payload

      return {
        ...state,
        ...{
          [ruleID]: {...state[ruleID], details},
        },
      }
    }

    case 'UPDATE_RULE_STATUS_SUCCESS': {
      const {ruleID, status} = action.payload

      return {
        ...state,
        ...{
          [ruleID]: {...state[ruleID], status},
        },
      }
    }
  }
  return state
}
