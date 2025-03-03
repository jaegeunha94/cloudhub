// Types
import {Source} from 'src/types'

export interface Query {
  text: string
  id: string
  queryConfig: QueryConfig
  type: string
  groupbys?: string[]
}

export interface QueryConfig {
  tags: Tags
  areTagsAccepted: boolean
  id?: string
  database?: string
  measurement?: string
  retentionPolicy?: string
  fields?: Field[]
  groupBy?: GroupBy
  rawText?: string
  range?: DurationRange | null
  source?: Source | null // doesn't come from server -- is set in CellEditorOverlay
  fill?: string
  status?: Status
  shifts?: TimeShift[]
  lower?: string
  upper?: string
  isQuerySupportedByExplorer?: boolean // doesn't come from server -- is set in CellEditorOverlay
  originalQuery?: string
}

export interface QueryStatus {
  queryID: string
  status: Status
}

export interface Field {
  value: string
  type: string
  alias?: string
  args?: FieldArg[]
  desc?: string
  subFunc?: string
}

export interface FieldArg {
  value: string
  type: string
  alias?: string
  args?: FieldArg[]
}

export interface FieldFunc extends Field {
  args?: FuncArg[]
}
export interface FuncArg {
  type: string
  value: string
  alias?: string
}

export interface ApplyFuncsToFieldArgs {
  field: Field
  funcs: FuncArg[]
  subFunc?: SelectedSubFunction
}

export interface Tag {
  key: string
  value: string
}

export type TagValues = string[]

export interface Tags {
  [key: string]: TagValues
}

export interface GroupBy {
  time?: string | null
  tags?: string[]
}

export interface Namespace {
  database: string
  retentionPolicy: string
}

export interface Status {
  loading?: boolean
  error?: string
  warn?: string
  success?: string
}

export interface TimeRange {
  lower: string
  lowerFlux?: string
  upper?: string | null
  seconds?: number
  format?: string
}

export interface TimeRangeWithType extends TimeRange {
  menuOption?: typeof INPUT_TIME_TYPE
}

export const INPUT_TIME_TYPE = {
  RELATIVE_TIME: 'relativeTime',
  //now()-2d

  TIMESTAMP: 'timestamp',
  //"2024-07-08T06:34:00.000Z"
} as const

export interface DurationRange {
  lower: string
  upper?: string
}

export interface TimeShift {
  label: string
  unit: string
  quantity: string
}

export interface TimeRangeOption extends TimeRange {
  defaultGroupBy: string
  seconds: number
  inputValue: string
  menuOption: string
}

export interface GroupBys {
  defaultGroupBy: number
  lower: string
}

export interface SelectedSubFunction {
  [key: string]: string[]
}

export type DashTimeV1Range = TimeRangeOption & {dashboardID: string}
