import React from 'react';

export default class Request extends React.Component {
  render () {
    return (
      <div>
        <div><b>Proto:</b> <samp>{this.props.page.requestProto}</samp></div>
        <div><b>Client addr:</b> <samp>{this.props.page.requestAddr}</samp></div>
        <div><b>Dump:</b></div>
        <pre>
          {this.props.page.requestDump}
        </pre>
      </div>
    )
  }
}
