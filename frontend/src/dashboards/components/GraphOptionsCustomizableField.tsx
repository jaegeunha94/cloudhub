import React, {Component, ChangeEvent} from 'react'
import {findDOMNode} from 'react-dom'
import {
  DragSourceSpec,
  DropTargetConnector,
  DragSourceMonitor,
  DragSource,
  DropTarget,
  DragSourceConnector,
  ConnectDragSource,
  ConnectDropTarget,
  ConnectDragPreview,
} from 'react-dnd'
import {ErrorHandling} from 'src/shared/decorators/errors'

const fieldType = 'field'

interface RenamableField {
  internalName: string
  displayName: string
  visible: boolean
  direction?: '' | 'asc' | 'desc'
  tempVar?: string
}

interface Props {
  internalName: string
  displayName: string
  visible: boolean
  index: number
  id: string
  key: string
  onFieldUpdate: (field: RenamableField) => void
  isDragging?: boolean
  connectDragSource?: ConnectDragSource
  connectDropTarget?: ConnectDropTarget
  connectDragPreview?: ConnectDragPreview
  moveField: (dragIndex: number, hoverIndex: number) => void
  direction: '' | 'asc' | 'desc'
  isUsingTempVar: boolean
  tempVar: string
}

const fieldSource: DragSourceSpec<Props> = {
  beginDrag(props) {
    return {
      id: props.id,
      index: props.index,
    }
  },
}

const fieldTarget = {
  hover(props, monitor, component) {
    const dragIndex = monitor.getItem().index
    const hoverIndex = props.index

    // Don't replace items with themselves
    if (dragIndex === hoverIndex) {
      return
    }

    // Determine rectangle on screen
    const domNode = findDOMNode(component) as Element
    const hoverBoundingRect = domNode.getBoundingClientRect()

    // Get vertical middle
    const hoverMiddleY = (hoverBoundingRect.bottom - hoverBoundingRect.top) / 2

    // Determine mouse position
    const clientOffset = monitor.getClientOffset()

    // Get pixels to the top
    const hoverClientY = clientOffset.y - hoverBoundingRect.top

    // Only perform the move when the mouse has crossed half of the items height
    // When dragging downwards, only move when the cursor is below 50%
    // When dragging upwards, only move when the cursor is above 50%

    // Dragging downwards
    if (dragIndex < hoverIndex && hoverClientY < hoverMiddleY) {
      return
    }

    // Dragging upwards
    if (dragIndex > hoverIndex && hoverClientY > hoverMiddleY) {
      return
    }
    // Time to actually perform the action
    props.moveField(dragIndex, hoverIndex)

    // Note: we're mutating the monitor item here!
    // Generally it's better to avoid mutations,
    // but it's good here for the sake of performance
    // to avoid expensive index searches.
    monitor.getItem().index = hoverIndex
  },
}

function MyDropTarget(dropv1, dropv2, dropfunc1) {
  return target => DropTarget(dropv1, dropv2, dropfunc1)(target) as any
}

function MyDragSource(dragv1, dragv2, dragfunc1) {
  return target => DragSource(dragv1, dragv2, dragfunc1)(target) as any
}

@ErrorHandling
@MyDropTarget(fieldType, fieldTarget, (connect: DropTargetConnector) => ({
  connectDropTarget: connect.dropTarget(),
}))
@MyDragSource(
  fieldType,
  fieldSource,
  (connect: DragSourceConnector, monitor: DragSourceMonitor) => ({
    connectDragSource: connect.dragSource(),
    connectDragPreview: connect.dragPreview(),
    isDragging: monitor.isDragging(),
  })
)
export default class GraphOptionsCustomizableField extends Component<Props> {
  constructor(props) {
    super(props)

    this.handleFieldRename = this.handleFieldRename.bind(this)
    this.handleToggleVisible = this.handleToggleVisible.bind(this)
    this.handleTemplateVariable = this.handleTemplateVariable.bind(this)
  }
  public render(): JSX.Element | null {
    const {
      internalName,
      displayName,
      isDragging,
      connectDragSource,
      connectDragPreview,
      connectDropTarget,
      visible,
      tempVar,
      isUsingTempVar,
    } = this.props

    const fieldClass = `customizable-field${isDragging ? ' dragging' : ''}`
    const labelClass = `${
      visible
        ? 'customizable-field--label'
        : 'customizable-field--label__hidden'
    }`

    const visibilityTitle = visible
      ? `Click to HIDE ${internalName}`
      : `Click to SHOW ${internalName}`

    return connectDragPreview(
      connectDropTarget(
        <div className={fieldClass}>
          <div
            className={labelClass}
            style={isUsingTempVar ? {width: '50%'} : null}
          >
            {connectDragSource(
              <div className="customizable-field--drag">
                <span className="hamburger" />
              </div>
            )}
            <div
              className="customizable-field--visibility"
              onClick={this.handleToggleVisible}
              title={visibilityTitle}
            >
              <span className={visible ? 'icon eye-open' : 'icon eye-closed'} />
            </div>
            <div className="customizable-field--name">{internalName}</div>
          </div>
          <input
            className="form-control input-sm customizable-field--input"
            type="text"
            spellCheck={false}
            id="internalName"
            value={displayName}
            data-test="custom-time-format"
            onChange={this.handleFieldRename}
            placeholder={`Rename ${internalName}`}
            disabled={!visible}
            style={
              isUsingTempVar
                ? internalName === 'time'
                  ? {width: 'calc(50% - 4px)'}
                  : {width: 'calc(25% - 4px)'}
                : null
            }
          />
          <input
            className="form-control input-sm customizable-field--input"
            type="text"
            spellCheck={false}
            id="tempVar"
            value={tempVar}
            data-test="custom-time-format"
            onChange={this.handleTemplateVariable}
            placeholder={`Template Variables`}
            disabled={!visible}
            style={
              !isUsingTempVar || internalName === 'time'
                ? {display: 'none'}
                : {width: 'calc(25% - 4px)'}
            }
          />
        </div>
      )
    )
  }

  private handleFieldRename(e: ChangeEvent<HTMLInputElement>) {
    const {
      onFieldUpdate,
      internalName,
      visible,
      direction,
      tempVar,
    } = this.props
    onFieldUpdate({
      internalName,
      displayName: e.target.value,
      visible,
      direction,
      tempVar,
    })
  }

  private handleToggleVisible() {
    const {
      onFieldUpdate,
      internalName,
      displayName,
      visible,
      direction,
      tempVar,
    } = this.props
    onFieldUpdate({
      internalName,
      displayName,
      visible: !visible,
      direction,
      tempVar,
    })
  }

  private handleTemplateVariable(e: ChangeEvent<HTMLInputElement>) {
    const {
      onFieldUpdate,
      internalName,
      displayName,
      visible,
      direction,
    } = this.props
    onFieldUpdate({
      internalName,
      displayName,
      tempVar: e.target.value,
      visible,
      direction,
    })
  }
}
