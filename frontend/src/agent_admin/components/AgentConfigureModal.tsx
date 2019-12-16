import React, {PureComponent} from 'react'

import {
  OverlayContainer,
  OverlayHeading,
  OverlayBody,
  Form,
  Button,
  ComponentColor,
} from 'src/reusable_ui'

interface Props {
  onCancel: () => void
  onConfirm: () => void
}

class AgentConfigureModal extends PureComponent<Props> {
  public render() {
    const {onCancel, onConfirm} = this.props

    return (
      <OverlayContainer>
        <OverlayHeading title="Check for You" onDismiss={onCancel} />
        <OverlayBody>
          <Form>
            <Form.Element>
              <div>
                The configuration has changed.
                <br /> Do you want to move without saving?
              </div>
            </Form.Element>
            <Form.Footer>
              <Button
                color={ComponentColor.Success}
                text={`Confirm`}
                titleText="Must choose at least 1 dashboard and set a name"
                onClick={onConfirm}
              />
              <Button text="Cancel" onClick={onCancel} />
            </Form.Footer>
          </Form>
        </OverlayBody>
      </OverlayContainer>
    )
  }
}

export default AgentConfigureModal
