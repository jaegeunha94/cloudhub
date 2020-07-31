import React, {ChangeEvent, KeyboardEvent, MouseEvent} from 'react'
import _ from 'lodash'
// Components
import {Form, Button, ComponentColor, Input, InputType} from 'src/reusable_ui'
import {ComponentStatus} from 'src/reusable_ui/types'

import {ShellInfo} from 'src/types'

interface Props {
  shells: ShellInfo[]
  host: string
  addr: string
  user: string
  pwd: string
  port: string
  getIP: string
  isNewEditor: boolean
  handleShellRemove: (nodename: ShellInfo['nodename']) => void
  handleShellUpdate: (shell: ShellInfo) => void
  handleChangeHost: (e: ChangeEvent<HTMLInputElement>) => void
  handleChangeAddress: (e: ChangeEvent<HTMLInputElement>) => void
  handleChangeID: (e: ChangeEvent<HTMLInputElement>) => void
  handleChangePassword: (e: ChangeEvent<HTMLInputElement>) => void
  handleChangePort: (e: ChangeEvent<HTMLInputElement>) => void
  handleOpenTerminal: () => void
}
const ShellForm = (props: Props) => {
  const {
    shells,
    host,
    addr,
    user,
    pwd,
    port,
    getIP,
    isNewEditor,
    handleOpenTerminal,
    handleChangeHost,
    handleChangeAddress,
    handleChangeID,
    handleChangePassword,
    handleChangePort,
  } = props

  const onKeyPressEnter = function(event: KeyboardEvent<HTMLInputElement>) {
    const enterKeyCode = 13
    const enterKey = 'Enter'

    if (
      event.keyCode === enterKeyCode ||
      event.charCode === enterKeyCode ||
      event.key === enterKey
    ) {
      if (!props.isNewEditor && host && addr && user && pwd && port) {
        handleOpenTerminal()
      }
    }
  }

  const nodenameChecker = (shells: ShellInfo[], nodename: string) => {
    const isNodeNameOverlap =
      _.findIndex(shells, shell => shell.nodename === nodename) < 0

    return isNodeNameOverlap
  }

  const btnControl = () => {
    const check = isNewEditor && !nodenameChecker(shells, host)

    if (!host || check) {
      return ComponentStatus.Disabled
    } else {
      return ComponentStatus.Default
    }
  }

  const handleMouseDown = (e: MouseEvent<HTMLInputElement>): void => {
    e.stopPropagation()
  }

  return (
    <Form>
      <Form.Element label="SESSION NAME">
        <>
          <Input
            value={host}
            onChange={handleChangeHost}
            onKeyPress={onKeyPressEnter}
            placeholder={'Connect Host'}
            type={InputType.Text}
            status={
              isNewEditor ? ComponentStatus.Default : ComponentStatus.Disabled
            }
            onMouseDown={handleMouseDown}
          />
          {isNewEditor && !nodenameChecker(shells, host) ? (
            <div className="alert alert-error">
              <span className="icon alerts" />
              <div className="alert-message">Change the Host name.</div>
            </div>
          ) : null}
        </>
      </Form.Element>
      <Form.Element label="ADDRESS">
        <>
          <Input
            value={addr}
            onChange={
              isNewEditor
                ? e => {
                    handleChangeAddress(e)
                    handleChangeHost(e)
                  }
                : handleChangeAddress
            }
            onKeyPress={onKeyPressEnter}
            placeholder={'Connect Address'}
            type={InputType.Text}
            onMouseDown={handleMouseDown}
          />
          {getIP ? (
            <div className="alert alert-success">
              <span className="icon checkmark" />
              <div className="alert-message">
                Auto changed to the WAN IP used
              </div>
            </div>
          ) : null}
        </>
      </Form.Element>
      <Form.Element label="ID">
        <Input
          value={user}
          onChange={handleChangeID}
          onKeyPress={onKeyPressEnter}
          placeholder={'Connect ID'}
          type={InputType.Text}
          onMouseDown={handleMouseDown}
        />
      </Form.Element>
      <Form.Element label="PASSWORD">
        <Input
          value={pwd}
          onChange={handleChangePassword}
          onKeyPress={onKeyPressEnter}
          placeholder={'Connect Password'}
          type={InputType.Password}
          onMouseDown={handleMouseDown}
        />
      </Form.Element>
      <Form.Element label="PORT">
        <Input
          value={port}
          onChange={handleChangePort}
          onKeyPress={onKeyPressEnter}
          placeholder={'Connect Port'}
          type={InputType.Text}
          onMouseDown={handleMouseDown}
        />
      </Form.Element>
      <Form.Footer>
        <Button
          color={ComponentColor.Default}
          text={`Remove`}
          onClick={e => {
            e.stopPropagation()
            props.handleShellRemove(host)
          }}
        />
        {isNewEditor ? (
          <Button
            color={ComponentColor.Primary}
            text={`Add Config`}
            onClick={() => {
              props.handleShellUpdate({isNewEditor, nodename: host, addr: addr})
            }}
            status={btnControl()}
          />
        ) : (
          <Button
            color={ComponentColor.Success}
            text={`Connect`}
            onClick={handleOpenTerminal}
          />
        )}
      </Form.Footer>
    </Form>
  )
}

export default ShellForm
