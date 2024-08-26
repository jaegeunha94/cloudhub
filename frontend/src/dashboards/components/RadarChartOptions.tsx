// Libraries
import React, {PureComponent, ChangeEvent} from 'react'
import {getDeep} from 'src/utils/wrappers'
import _ from 'lodash'

// Components
import Input from 'src/dashboards/components/DisplayOptionsInput'
import OptIn from 'src/shared/components/OptIn'
import {Radio, ButtonShape} from 'src/reusable_ui'
import FancyScrollbar from 'src/shared/components/FancyScrollbar'
import LineGraphColorSelector from 'src/shared/components/LineGraphColorSelector'
import Dropdown from 'src/shared/components/Dropdown'

// Constants
import {AXES_SCALE_OPTIONS} from 'src/dashboards/constants/cellEditor'
import {STATISTICAL_GRAPH_TYPES} from 'src/dashboards/graphics/graph'
import {DEFAULT_STATISTICAL_TIME_FIELD} from 'src/dashboards/constants/'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

// Types
import {Axes, DropdownItem, Template} from 'src/types'
import {GraphOptions, StaticLegendPositionType} from 'src/types/dashboards'
import {ColorString} from 'src/types/colors'

// Utils
import {
  getSelectedShowTemplateVariable,
  getShowTemplateVariable,
} from 'src/shared/utils/staticGraph'
import {
  DropdownOption,
  RenamableField,
  TableOptionsInterface,
} from 'src/types/statisticalgraph'
import GraphOptionsSortBy from 'src/dashboards/components/GraphOptionsSortBy'
import GraphOptionsCustomizeFields from 'src/dashboards/components/GraphOptionsCustomizeFields'

const {LINEAR, BASE_2, BASE_10, BASE_RAW} = AXES_SCALE_OPTIONS
const getInputMin = () => (-Infinity).toString()

interface Props {
  groupByTag: string[]
  tableOptions: TableOptionsInterface
  fieldOptions: RenamableField[]
  type: string
  axes: Axes
  graphOptions: GraphOptions
  staticLegend: boolean
  staticLegendPosition: StaticLegendPositionType
  defaultYLabel: string
  dashboardTemplates?: Template[]
  lineColors: ColorString[]
  onUpdateAxes: (axes: Axes) => void
  onUpdateGraphOptions: (graphOptions: GraphOptions) => void
  onToggleStaticLegend: (isStaticLegend: boolean) => void
  onUpdateStaticLegendPosition: (
    staticLegendPosition: StaticLegendPositionType
  ) => void
  onUpdateLineColors: (colors: ColorString[]) => void
  onUpdateTableOptions: (options: TableOptionsInterface) => void
  onUpdateFieldOptions: (fieldOptions: RenamableField[]) => void
}

interface State {
  rPrefix: string
  rSuffix: string
}

@ErrorHandling
class RadarChartOptions extends PureComponent<Props, State> {
  public static defaultProps: Partial<Props> = {
    axes: {
      y: {
        bounds: ['', ''],
        prefix: '',
        suffix: '',
        base: BASE_RAW,
        scale: LINEAR,
        label: '',
      },
      x: {
        bounds: ['', ''],
        prefix: '',
        suffix: '',
        base: BASE_RAW,
        scale: LINEAR,
        label: '',
      },
    },
  }
  constructor(props) {
    super(props)
    this.state = {
      rPrefix: getDeep<string>(props, 'axes.y.prefix', ''),
      rSuffix: getDeep<string>(props, 'axes.y.suffix', ''),
    }
    this.moveField = this.moveField.bind(this)
  }

  public render() {
    const {
      axes: {
        y: {bounds},
      },
      groupByTag,
      type,
      lineColors,
      defaultYLabel,
      fieldOptions,
      tableOptions,
      onUpdateLineColors,
    } = this.props
    const {rPrefix, rSuffix} = this.state

    const [min, max] = bounds
    const {menuOption} = STATISTICAL_GRAPH_TYPES.find(
      graph => graph.type === type
    )
    const tableSortByOptions = fieldOptions
      .map(field => ({
        key: field.internalName,
        text: field.displayName || field.internalName,
      }))
      .filter(field => field?.key !== 'time')
    const customizeFieldOptions = fieldOptions.filter(fieldOption => {
      if (
        fieldOption.internalName === 'time' ||
        groupByTag.includes(fieldOption.internalName)
      ) {
        return false
      }
      return true
    })
    const isValidSelectedSortField =
      tableOptions?.sortBy?.internalName !== 'time' &&
      tableOptions?.sortBy?.internalName !== '' &&
      tableSortByOptions.some(
        tableSortByOption =>
          tableSortByOption?.key === tableOptions?.sortBy?.internalName
      )

    const selectedGraphOptionSortField = this.getSelectedGraphOptionSortField()
    const firstGroupByTag = groupByTag[0]
    const selectedSortFieldByFirstGroupBy =
      _.get(
        _.find(fieldOptions, {internalName: firstGroupByTag}),
        'displayName'
      ) || firstGroupByTag

    const selectedSortFieldByFirstField =
      _.get(
        _.find(fieldOptions, {internalName: defaultYLabel}),
        'displayName'
      ) || defaultYLabel
    const defaultStatisticalTimeField: RenamableField = {
      ...DEFAULT_STATISTICAL_TIME_FIELD,
      internalName:
        firstGroupByTag === undefined
          ? selectedSortFieldByFirstField
          : selectedSortFieldByFirstGroupBy,
    }
    return (
      <FancyScrollbar className="display-options" autoHide={false}>
        <div className="display-options--wrapper">
          <h5 className="display-options--header">{menuOption} Controls</h5>
          <form autoComplete="off" className="form-group-wrapper">
            <GraphOptionsSortBy
              selected={
                isValidSelectedSortField
                  ? selectedGraphOptionSortField
                  : defaultStatisticalTimeField
              }
              selectedDirection={tableOptions?.sortBy?.direction || 'asc'}
              sortByOptions={tableSortByOptions}
              onChooseSortBy={this.handleChooseSortBy}
              onChooseSortByDirection={this.handleChooseSortByDirection}
            />
            <div
              className="form-group col-xs-6"
              style={{width: '100%', marginBottom: '30px'}}
            >
              <GraphOptionsCustomizeFields
                fields={customizeFieldOptions}
                onFieldUpdate={this.handleFieldUpdate}
                moveField={this.moveField}
                isUsingTempVar={false}
              />
            </div>
            <Input
              name="r-prefix"
              id="r-prefix"
              value={rPrefix}
              labelText="R-Value's Prefix"
              onChange={this.handleSetRAxisPrefix}
            />
            <Input
              name="r-suffix"
              id="r-suffix"
              value={rSuffix}
              labelText="R-Value's Suffix"
              onChange={this.handleSetRAxisSuffix}
            />
            <LineGraphColorSelector
              onUpdateLineColors={onUpdateLineColors}
              lineColors={lineColors}
            />
            <div className="form-group col-sm-6">
              <label htmlFor="min">Min</label>
              <OptIn
                customPlaceholder={'min'}
                customValue={min}
                onSetValue={this.handleSetRAxisBoundMin}
                type="number"
                min={getInputMin()}
              />
            </div>
            <div className="form-group col-sm-6">
              <label htmlFor="max">Max</label>
              <OptIn
                customPlaceholder="max"
                customValue={max}
                onSetValue={this.handleSetRAxisBoundMax}
                type="number"
                min={getInputMin()}
              />
            </div>
            {this.yValuesFormatTabs}

            {this.staticLegendTabs}
            {this.staticLegendPositionTabs}
            {this.showCount}
          </form>
        </div>
      </FancyScrollbar>
    )
  }
  private getSelectedGraphOptionSortField(): RenamableField {
    const {fieldOptions, tableOptions} = this.props

    const matchedFieldOption = _.find(fieldOptions, {
      internalName: tableOptions?.sortBy?.internalName,
    })

    if (
      matchedFieldOption &&
      matchedFieldOption?.displayName !== tableOptions?.sortBy?.displayName
    ) {
      return matchedFieldOption
    }

    return tableOptions?.sortBy
  }

  private get staticLegendTabs(): JSX.Element {
    const {staticLegend, onToggleStaticLegend} = this.props

    return (
      <div className="form-group col-sm-6">
        <label>Static Legend</label>
        <Radio shape={ButtonShape.StretchToFit}>
          <Radio.Button
            id="static-legend-tab--show"
            value={true}
            active={staticLegend === true}
            titleText="Show static legend below graph"
            onClick={onToggleStaticLegend}
          >
            Show
          </Radio.Button>
          <Radio.Button
            id="static-legend-tab--hide"
            value={false}
            active={staticLegend === false}
            titleText="Hide static legend"
            onClick={onToggleStaticLegend}
          >
            Hide
          </Radio.Button>
        </Radio>
      </div>
    )
  }

  private get showCount(): JSX.Element {
    const {graphOptions, dashboardTemplates} = this.props
    const selectedShowCount = getSelectedShowTemplateVariable(graphOptions)
    const showCountItems = getShowTemplateVariable(dashboardTemplates)
    return (
      <div className="form-group col-sm-6">
        <label>Show Count</label>
        <div className="show-count-field">
          <Dropdown
            items={showCountItems}
            selected={selectedShowCount}
            buttonColor="btn-default"
            buttonSize="btn-sm"
            className="dropdown-stretch"
            onChoose={this.handleUpdateShowCount}
          />
        </div>
      </div>
    )
  }

  private get staticLegendPositionTabs(): JSX.Element {
    const {staticLegendPosition, onUpdateStaticLegendPosition} = this.props

    return (
      <div className="form-group col-sm-6">
        <label>Static Legend Position</label>
        <Radio shape={ButtonShape.StretchToFit}>
          <Radio.Button
            id="static-legend-tab--top"
            value={true}
            active={staticLegendPosition === 'top'}
            titleText="Show static legend on the top side"
            onClick={() => onUpdateStaticLegendPosition('top')}
          >
            Top
          </Radio.Button>
          <Radio.Button
            id="static-legend-tab--bottom"
            value={false}
            active={staticLegendPosition === 'bottom'}
            titleText="Show static legend on the bottom side"
            onClick={() => onUpdateStaticLegendPosition('bottom')}
          >
            Bottom
          </Radio.Button>
          <Radio.Button
            id="static-legend-tab--left"
            value={false}
            active={staticLegendPosition === 'left'}
            titleText="Show static legend on the left side"
            onClick={() => onUpdateStaticLegendPosition('left')}
          >
            Left
          </Radio.Button>
          <Radio.Button
            id="static-legend-tab--right"
            value={false}
            active={staticLegendPosition === 'right'}
            titleText="Show static legend on the right side"
            onClick={() => onUpdateStaticLegendPosition('right')}
          >
            Right
          </Radio.Button>
        </Radio>
      </div>
    )
  }

  private get yValuesFormatTabs(): JSX.Element {
    const {
      axes: {
        y: {base},
      },
    } = this.props

    return (
      <div className="form-group col-sm-6">
        <label>Y-Value's Format</label>
        <Radio shape={ButtonShape.StretchToFit}>
          <Radio.Button
            id="y-values-format-tab--raw"
            value={BASE_RAW}
            active={base === '' || base === BASE_RAW}
            titleText="Don't format values"
            onClick={this.handleSetBase}
          >
            Raw
          </Radio.Button>
          <Radio.Button
            id="y-values-format-tab--kmb"
            value={BASE_10}
            active={base === BASE_10}
            titleText="Thousand / Million / Billion"
            onClick={this.handleSetBase}
          >
            K/M/B
          </Radio.Button>
          <Radio.Button
            id="y-values-format-tab--kmg"
            value={BASE_2}
            active={base === BASE_2}
            titleText="Kilo / Mega / Giga"
            onClick={this.handleSetBase}
          >
            K/M/G
          </Radio.Button>
        </Radio>
      </div>
    )
  }
  private handleSetRAxisPrefix = (e: ChangeEvent<HTMLInputElement>): void => {
    const {onUpdateAxes, axes} = this.props
    const rPrefix = e.target.value

    const newAxes = {
      ...axes,
      y: {
        ...axes.y,
        prefix: rPrefix,
      },
    }

    this.setState({rPrefix: rPrefix}, () => onUpdateAxes(newAxes))
  }

  private handleSetRAxisSuffix = (e: ChangeEvent<HTMLInputElement>): void => {
    const {onUpdateAxes, axes} = this.props
    const rSuffix = e.target.value

    const newAxes = {
      ...axes,
      y: {
        ...axes.y,
        suffix: rSuffix,
      },
    }
    this.setState({rSuffix: rSuffix}, () => onUpdateAxes(newAxes))
  }

  private handleSetRAxisBoundMin = (min: string): void => {
    const {onUpdateAxes, axes} = this.props
    const {
      y: {
        bounds: [, max],
      },
    } = this.props.axes

    const bounds: [string, string] = [min, max]
    const newAxes = {...axes, y: {...axes.y, bounds}}

    onUpdateAxes(newAxes)
  }

  private handleSetRAxisBoundMax = (max: string): void => {
    const {onUpdateAxes, axes} = this.props
    const {
      y: {
        bounds: [min],
      },
    } = axes

    const bounds: [string, string] = [min, max]
    const newAxes = {...axes, y: {...axes.y, bounds}}

    onUpdateAxes(newAxes)
  }

  private handleSetBase = (base: string): void => {
    const {onUpdateAxes, axes} = this.props
    const newAxes = {...axes, y: {...axes.y, base}}

    onUpdateAxes(newAxes)
  }

  private handleFieldUpdate = field => {
    const {
      onUpdateTableOptions,
      onUpdateFieldOptions,
      tableOptions,
      fieldOptions,
    } = this.props
    const {sortBy} = tableOptions

    const updatedFieldOptions = fieldOptions.map(f =>
      f.internalName === field.internalName ? field : f
    )

    if (sortBy.internalName === field.internalName) {
      const updatedSortBy = {...sortBy, displayName: field.displayName}
      onUpdateTableOptions({
        ...tableOptions,
        sortBy: updatedSortBy,
      })
    }

    onUpdateFieldOptions(updatedFieldOptions)
  }

  private filterExcludedFields(fieldOption) {
    const {groupByTag} = this.props

    if (
      fieldOption.internalName === 'time' ||
      groupByTag.includes(fieldOption?.internalName)
    ) {
      return false
    }
    return true
  }

  private findActualIndex(filteredFieldOptions, filteredIndex) {
    return _.get(filteredFieldOptions, `[${filteredIndex}].originalIndex`, 0)
  }

  private moveField(dragIndex, hoverIndex) {
    const {onUpdateFieldOptions, fieldOptions} = this.props

    const filteredFieldOptions = fieldOptions
      .map((field, index) => ({field, originalIndex: index}))
      .filter(item => this.filterExcludedFields(item.field))

    const actualDragIndex = this.findActualIndex(
      filteredFieldOptions,
      dragIndex
    )
    const actualHoverIndex = this.findActualIndex(
      filteredFieldOptions,
      hoverIndex
    )
    const draggedField = fieldOptions[actualDragIndex]
    let newFieldOptions = [...fieldOptions]

    newFieldOptions.splice(actualDragIndex, 1)
    newFieldOptions.splice(actualHoverIndex, 0, draggedField)

    onUpdateFieldOptions(newFieldOptions)
  }

  private handleChooseSortBy = (option: DropdownOption) => {
    const {tableOptions, onUpdateTableOptions, fieldOptions} = this.props
    const sortBy = fieldOptions.find(f => f.internalName === option.key)

    onUpdateTableOptions({...tableOptions, sortBy})
  }

  private handleChooseSortByDirection = (direction: 'asc' | 'desc') => {
    const {tableOptions, onUpdateTableOptions, fieldOptions} = this.props
    const sortBy = fieldOptions.find(
      f => f.internalName === tableOptions.sortBy.internalName
    )
    const updatedSortBy = {...sortBy, direction: direction}

    onUpdateTableOptions({...tableOptions, sortBy: updatedSortBy})
  }

  private handleUpdateShowCount = (item: DropdownItem): void => {
    const {onUpdateGraphOptions, graphOptions} = this.props
    const newGraphOptions = {...graphOptions, showTempVarCount: item.text}

    onUpdateGraphOptions(newGraphOptions)
  }
}

export default RadarChartOptions
