import React, {Component} from 'react'
import PropTypes from 'prop-types'

class RedactedInput extends Component {
  constructor(props) {
    super(props)
    this.state = {
      editing: false,
    }
  }

  handleClick = () => {
    this.setState({editing: true})
  }

  render() {
    const {defaultValue, id, refFunc, disableTest} = this.props
    const {editing} = this.state

    if (defaultValue === true && !editing) {
      return (
        <div className="form-control-static redacted-input">
          <span className="alert-value-set">
            <span className="icon checkmark" /> Value set
          </span>
          <button className="btn btn-xs btn-link" onClick={this.handleClick}>
            Change
          </button>
          <input
            className="form-control"
            id={id}
            type="hidden"
            ref={refFunc}
            defaultValue={defaultValue}
          />
        </div>
      )
    }

    return (
      <input
        className="form-control"
        id={id}
        type="text"
        ref={refFunc}
        defaultValue={''}
        onChange={disableTest}
      />
    )
  }
}

const {bool, func, string} = PropTypes

RedactedInput.propTypes = {
  id: string.isRequired,
  defaultValue: bool,
  refFunc: func.isRequired,
  disableTest: func,
}

export default RedactedInput
