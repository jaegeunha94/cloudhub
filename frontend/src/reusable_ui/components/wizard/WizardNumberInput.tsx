import React, {ChangeEvent, KeyboardEvent, PureComponent} from 'react'
import {ErrorHandling} from 'src/shared/decorators/errors'

interface Validation {
  status: boolean
  reason: string
}

interface Props {
  value: string
  label: string
  isValid?: (value) => Validation
  isDisabled?: boolean
  onChange: (value: string) => void
  valueModifier?: (value: string) => string
  onSubmit?: (value: string) => void
  placeholder?: string
  autoFocus?: boolean
  type?: string
  newClassName?: string
  min?: number
  labelClassName?: string
}

interface State {
  initialValue: string
}

@ErrorHandling
class WizardNumberInput extends PureComponent<Props, State> {
  public static defaultProps: Partial<Props> = {
    value: '',
    isDisabled: false,
    isValid: () => ({
      status: true,
      reason: '',
    }),
    valueModifier: x => x,
    autoFocus: false,
    type: 'number',
    onSubmit: () => null,
    min: 0,
  }

  constructor(props) {
    super(props)

    this.state = {
      initialValue: this.props.value,
    }
  }

  public render() {
    const {
      isDisabled,
      placeholder,
      isValid,
      value,
      autoFocus,
      label,
      type,
      newClassName,
      min,
      labelClassName,
    } = this.props

    let inputClass = ''
    let errorText = ''
    const validation = isValid(value)
    if (validation.status === false) {
      inputClass = 'form-volcano'
      errorText = validation.reason
    }
    const divClassName = newClassName || 'form-group col-xs-6'

    return (
      <div className={divClassName}>
        <label htmlFor={label} className={labelClassName}>
          {label}
        </label>
        <input
          type={type}
          id={label}
          className={`form-control input-sm ${inputClass}`}
          value={value}
          placeholder={placeholder}
          onBlur={this.handleBlur}
          onKeyDown={this.handleKeyDown}
          onChange={this.handleChange}
          disabled={isDisabled}
          autoFocus={autoFocus}
          spellCheck={false}
          autoComplete={'off'}
          {...(type === 'number' && min !== undefined ? {min} : {})}
        />
        {errorText}
      </div>
    )
  }

  private handleBlur = (e: ChangeEvent<HTMLInputElement>) => {
    this.submit(e.target.value)
  }

  private handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      this.submit(e.currentTarget.value)
    }
    if (e.key === 'Escape') {
      this.handleEscape()
    }
  }

  private handleEscape = () => {
    const {onChange} = this.props
    const {initialValue} = this.state
    onChange(initialValue)
  }

  private submit = incomingValue => {
    const {onChange, value, valueModifier, onSubmit} = this.props
    const newValue = valueModifier(incomingValue)
    onSubmit(newValue)

    if (value !== newValue) {
      onChange(newValue)
    }

    this.setState({
      initialValue: newValue,
    })
  }

  private handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const {onChange} = this.props
    if (onChange) {
      onChange(e.target.value)
    }
  }
}

export default WizardNumberInput
