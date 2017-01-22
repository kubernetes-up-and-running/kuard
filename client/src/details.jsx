import React from 'react';

class Details extends React.Component {
  propTypes: {
    title: React.PropTypes.string.isRequired,
    open: React.PropTypes.bool
  }

  defaultProps: {
    open: false
  }

  render () {
    return (
      <details open={this.props.open}>
        <summary>{this.props.title}</summary>
        <div>
          {this.props.children}
        </div>
      </details>
    )
  }
}

module.exports = Details;
