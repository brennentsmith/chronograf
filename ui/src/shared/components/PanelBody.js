import React from 'react'
import PropTypes from 'prop-types'

const {node} = PropTypes
const PanelBody = React.createClass({
  propTypes: {
    children: node.isRequired,
  },

  render() {
    return (
      <div className="panel-body text-center">
        <h3 className="deluxe">How to resolve:</h3>
        <p>{this.props.children}</p>
      </div>
    )
  },
})

export default PanelBody
