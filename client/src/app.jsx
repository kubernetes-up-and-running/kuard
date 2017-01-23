import React from 'react';
import Details from './details';
import Env from './env';
import Probe from './probe';

class App extends React.Component {
  render () {
    return (
      <div>
        <div className="title">
          <h1>Kubernetes Up and Running</h1>
          <div>Demo application version <i>{this.props.page.version}</i></div>
        </div>

        <div className="warning">WARNING: This server may expose sensitive and secret information. Be careful.</div>

        <Details open={true} title="Request Details">
          <div><b>Proto:</b> <code>{this.props.page.requestProto}</code></div>
          <div><b>Client addr:</b> <code>{this.props.page.requestAddr}</code></div>
          <div><b>Dump:</b></div>
          <pre>
            {this.props.page.requestDump}
          </pre>
        </Details>

        <Env path="/env" />

        <Probe title="Liveness Check" path="/healthy" />

        <Probe title="Readiness Check" path="/ready" />


        <Details title="File System">
            <a href="/fs/">Browse the root file system for this server.</a>
        </Details>
      </div>
    )
  }
}

module.exports = App;
