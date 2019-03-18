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

export default class App extends React.Component {
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

    let base = this.props.page.urlBase;

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
            <HighlightLink href={base+"/"} className="nav-item">Request Details</HighlightLink>
            <HighlightLink href={base+"/-/env"} className="nav-item">Server Env</HighlightLink>
            <HighlightLink href={base+"/-/mem"} className="nav-item">Memory</HighlightLink>
            <HighlightLink href={base+"/-/liveness"} className="nav-item">Liveness Probe</HighlightLink>
            <HighlightLink href={base+"/-/readiness"} className="nav-item">Readiness Probe</HighlightLink>
            <HighlightLink href={base+"/-/dns"} className="nav-item">DNS Query</HighlightLink>
            <HighlightLink href={base+"/-/keygen"} className="nav-item">KeyGen Workload</HighlightLink>
            <HighlightLink href={base+"/-/memq"} className="nav-item">MemQ Server</HighlightLink>
            <a className="nav-item" href={base+"/fs/"}>File system browser</a>
          </div>
          <div className="content">
            <Locations onNavigation={this.handleNavigation.bind(this)}>
              <Location path={base+"/"} handler={Request} page={this.props.page}/>
              <Location path={base+"/-/env"} apiPath={base+"/env/api"} handler={Env}/>
              <Location path={base+"/-/mem"} apiPath={base+"/mem/api"} handler={Mem}/>
              <Location path={base+"/-/liveness"} serverPath={base+"/healthy"} handler={Probe}/>
              <Location path={base+"/-/readiness"} serverPath={base+"/ready"} handler={Probe}/>
              <Location path={base+"/-/dns"} serverPath={base+"/dns"} handler={Dns}/>
              <Location path={base+"/-/keygen"} serverPath={base+"/keygen"} handler={KeyGen}/>
              <Location path={base+"/-/memq"} serverPath={base+"/memq"} handler={MemQ}/>
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

