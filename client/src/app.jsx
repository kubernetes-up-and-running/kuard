import React from 'react';
import Env from './env';
import Mem from './mem';
import Probe from './probe';
import Dns from './dns';
import KeyGen from './keygen';
import Request from './request';
import HighlightLink from './highlightlink'
import Disconnected from './disconnected'
import MemQ from './memq'
import { Location, Locations } from 'react-router-component';

function createElement(Component, props) {
  console.log(props)
  return <Component {...props}/>
}

class App extends React.Component {
  getChildContext() {
    return {
      reportConnError: () => {
        if (this.disconnected) {
          this.disconnected.reportConnError()
        }
      }
    }
  }

  handleNavigation() {
    this.forceUpdate()
  }

  reportConnError() {
      this.disconnected.
      this.timer = setInterval(this.loadState.bind(this), 1000);
  }

  render () {
    let addrs = [];
    for (let a of this.props.page.addrs) {
      addrs.push(<span key={a}>{a}</span>, " ")
    }

    return (
      <div className="top">
        <div className="title">
          <div className="alert alert-danger" role="alert">
            <svg className="icon icon-notification"><use xlinkHref="#icon-notification"></use></svg> { " " }
            <b>WARNING:</b> This server may expose sensitive and secret information. Be careful.
          </div>
          <Disconnected ref={el => this.disconnected = el}/>
          <h2><samp>{this.props.page.hostname}</samp></h2>
          <div>Demo application version <i>{this.props.page.version}</i></div>
          <div>Serving on {addrs}</div>
        </div>

        <div className="nav-container">
          <div className="nav">
            <HighlightLink href="/" className="nav-item">Request Details</HighlightLink>
            <HighlightLink href="/-/env" className="nav-item">Server Env</HighlightLink>
            <HighlightLink href="/-/mem" className="nav-item">Memory</HighlightLink>
            <HighlightLink href="/-/liveness" className="nav-item">Liveness Probe</HighlightLink>
            <HighlightLink href="/-/readiness" className="nav-item">Readiness Probe</HighlightLink>
            <HighlightLink href="/-/dns" className="nav-item">DNS Query</HighlightLink>
            <HighlightLink href="/-/keygen" className="nav-item">KeyGen Workload</HighlightLink>
            <HighlightLink href="/-/memq" className="nav-item">MemQ Server</HighlightLink>
            <a className="nav-item" href="/fs/">File system browser</a>
          </div>
          <div className="content">
            <Locations onNavigation={this.handleNavigation.bind(this)}>
              <Location path="/" handler={Request} page={this.props.page}/>
              <Location path="/-/env" apiPath="/env/api" handler={Env}/>
              <Location path="/-/mem" apiPath="/mem/api" handler={Mem}/>
              <Location path="/-/liveness" serverPath="/healthy" handler={Probe}/>
              <Location path="/-/readiness" serverPath="/ready" handler={Probe}/>
              <Location path="/-/dns" serverPath="/dns" handler={Dns}/>
              <Location path="/-/keygen" serverPath="/keygen" handler={KeyGen}/>
              <Location path="/-/memq" serverPath="/memq" handler={MemQ}/>
            </Locations>
          </div>
        </div>
      </div>
    )
  }
}

App.childContextTypes = {
  reportConnError: React.PropTypes.func
}

module.exports = App;
