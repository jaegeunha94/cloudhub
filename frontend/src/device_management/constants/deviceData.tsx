import {
  DeviceData,
  DropdownItem,
  LearningOption,
  SNMPConfig,
  SSHConfig,
} from 'src/types'

export const DEFAULT_SNMP_CONFIG: SNMPConfig = {
  community: '',
  port: 161,
  version: '1',
  protocol: 'UDP',
}

export const DEFAULT_SSH_CONFIG: SSHConfig = {
  user_id: '',
  password: '',
  en_password: '',
  port: 22,
}

export const DEFAULT_NETWORK_DEVICE_DATA: DeviceData = {
  device_ip: '',
  organization: 'Default',
  snmp_config: DEFAULT_SNMP_CONFIG,
  ssh_config: DEFAULT_SSH_CONFIG,
}

export const SNMP_VERSION: DropdownItem[] = [{text: '1'}, {text: '2c'}]

export const SNMP_PROTOCOL: DropdownItem[] = [
  {text: 'UDP'},
  {text: 'UDP4'},
  {text: 'UDP6'},
  {text: 'TCP'},
  {text: 'TCP4'},
  {text: 'TCP6'},
]

export const IMPORT_DEVICE_CSV_Template =
  'device_ip,organization,snmp_community,snmp_port,snmp_version,snmp_protocol,ssh_user_id,ssh_password,ssh_en_password,ssh_port'

export const SNMP_CONNECTION_URL = '/cloudhub/v1/snmp/validation'

export const DEVICE_MANAGEMENT_URL =
  '/cloudhub/v1//ai/network/managements/devices'

export const APPLY__MONITORING_URL =
  '/cloudhub/v1/ai/network/managements/monitoring/config'

export const APPLY_LEARNING_ENABLE_STATUS_URL =
  '/cloudhub/v1/ai/network/managements/learning/config'

export const DELETE_MODAL_INFO = {
  message: `Are you sure you want to delete this?`,
}

export const SYSTEM_MODAL = {
  LEARNING: 'learning',
  DELETE: 'delete',
  MONITORING_DELETE: 'monitoring_delete',
} as const

export const MLFunctionMsg = {
  ml_multiplied: 'Correlation Coefficient',
  ml_scaling_normalized: 'Scaling Normalized',
  ml_gaussian_std: 'Gaussian Standard Deviation',
} as const

export const DEFAULT_LEARNING_OPTION: LearningOption = {
  organization: '',
  data_duration: 1,
  ml_function: 'ml_gaussian_std' as typeof MLFunctionMsg[keyof typeof MLFunctionMsg],
  relearn_cycle: '',
}

export const NETWORK_MANAGEMENT_ORGANIZATIONS_URL =
  '/cloudhub/v1/ai/network/managements/orgs'
