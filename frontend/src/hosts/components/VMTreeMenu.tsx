import React from 'react'
import {debounce} from 'lodash'

import walk, {
  TreeNode,
  Item,
  TreeNodeInArray,
  LocaleFunction,
  MatchSearchFunction,
} from 'src/reusable_ui/components/treemenu/TreeMenu/walk'
import {
  defaultChildren,
  TreeMenuChildren,
  TreeMenuItem,
} from 'src/reusable_ui/components/treemenu/TreeMenu/renderProps'
import KeyDown from 'src/reusable_ui/components/treemenu/KeyDown'

export type TreeMenuProps = {
  data: {[name: string]: TreeNode} | TreeNodeInArray[]
  activeKey?: string
  focusKey?: string
  initialActiveKey?: string
  initialFocusKey?: string
  initialOpenNodes?: string[]
  openNodes?: string[]
  hasSearch?: boolean
  onClickItem?: (props: Item) => void
  onClickToggle?: (node: string) => void
  debounceTime?: number
  children?: TreeMenuChildren
  locale?: LocaleFunction
  matchSearch?: MatchSearchFunction
}

type TreeMenuState = {
  openNodes: string[]
  searchTerm: string
  activeKey: string
  focusKey: string
}

const defaultOnClick = (props: Item) => console.log(props) // eslint-disable-line no-console
const defaultOnClickToggle = (node: string) => console.log(node) // eslint-disable-line no-console

class VMTreeMenu extends React.PureComponent<TreeMenuProps, TreeMenuState> {
  constructor(props) {
    super(props)

    this.state = {
      openNodes: this.props.initialOpenNodes || [],
      searchTerm: '',
      activeKey: this.props.initialActiveKey || '',
      focusKey: this.props.initialFocusKey || '',
    }
  }

  static defaultProps: TreeMenuProps = {
    data: {},
    onClickItem: defaultOnClick,
    onClickToggle: defaultOnClickToggle,
    debounceTime: 125,
    children: defaultChildren,
    hasSearch: true,
  }

  public reset = (newOpenNodes?: string[]) => {
    const {initialOpenNodes} = this.props
    const openNodes =
      (Array.isArray(newOpenNodes) && newOpenNodes) || initialOpenNodes || []
    this.setState({openNodes, searchTerm: ''})
  }

  public search = (value: string) => {
    const {debounceTime} = this.props
    const search = debounce(
      (searchTerm: string) => this.setState({searchTerm}),
      debounceTime
    )
    search(value)
  }

  public toggleNode = (node: string) => {
    if (!this.props.openNodes) {
      const {openNodes} = this.state
      const newOpenNodes = openNodes.includes(node)
        ? openNodes.filter((openNode) => openNode !== node)
        : [...openNodes, node]
      this.setState({openNodes: newOpenNodes})
    } else {
      if (this.props.onClickToggle) {
        this.props.onClickToggle(node)
      }
    }
  }

  public generateItems = (): TreeMenuItem[] => {
    const {data, onClickItem, locale, matchSearch} = this.props
    const {searchTerm} = this.state
    const openNodes = this.props.openNodes || this.state.openNodes
    const activeKey = this.props.activeKey || this.state.activeKey
    const focusKey = this.props.focusKey || this.state.focusKey

    const items: Item[] = data
      ? walk({data, openNodes, searchTerm, locale, matchSearch})
      : []

    return items.map((item) => {
      const focused = item.key === focusKey
      const active = item.key === activeKey

      const onClick = () => {
        const newActiveKey = this.props.activeKey || item.key
        this.setState({activeKey: newActiveKey, focusKey: newActiveKey})
        onClickItem && onClickItem(item)
      }

      const toggleNode = item.hasNodes
        ? () => this.toggleNode(item.key)
        : undefined

      const newItem = {...item, focused, active, onClick, toggleNode}
      return newItem
    })
  }

  public getKeyDownProps = (items: TreeMenuItem[]) => {
    const {onClickItem} = this.props
    const {focusKey, activeKey} = this.state
    const focusIndex = items.findIndex(
      (item) => item.key === (focusKey || activeKey)
    )
    const getFocusKey = (item: TreeMenuItem) => {
      const keyArray = item.key.split('/')

      return keyArray.length > 1
        ? keyArray.slice(0, keyArray.length - 1).join('/')
        : item.key
    }

    return {
      up: () => {
        this.setState(({focusKey}) => ({
          focusKey: focusIndex > 0 ? items[focusIndex - 1].key : focusKey,
        }))
      },
      down: () => {
        this.setState(({focusKey}) => ({
          focusKey:
            focusIndex < items.length - 1
              ? items[focusIndex + 1].key
              : focusKey,
        }))
      },
      left: () => {
        this.setState(({openNodes, ...rest}) => {
          const item = items[focusIndex]
          const newOpenNodes = openNodes.filter((node) => node !== item.key)

          return item.isOpen
            ? {...rest, openNodes: newOpenNodes, focusKey: item.key}
            : {...rest, focusKey: getFocusKey(item)}
        })
      },
      right: () => {
        const {hasNodes, key} = items[focusIndex]
        if (hasNodes)
          this.setState(({openNodes}) => ({openNodes: [...openNodes, key]}))
      },
      enter: () => {
        this.setState(({focusKey}) => ({activeKey: focusKey}))
        onClickItem && onClickItem(items[focusIndex])
      },
    }
  }

  public render() {
    const {children, hasSearch} = this.props
    const {searchTerm} = this.state

    const items = this.generateItems()
    const renderedChildren = children || defaultChildren
    const keyDownProps = this.getKeyDownProps(items)

    return (
      <KeyDown {...keyDownProps}>
        {renderedChildren(
          hasSearch
            ? {search: this.search, items, reset: this.reset, searchTerm}
            : {items, reset: this.reset}
        )}
      </KeyDown>
    )
  }
}

export default VMTreeMenu
