import React from 'react';
import fetchError from './fetcherror';

export default class Probe extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      probePath: '',
      failNext: 0,
      history: []
    };
  }

  loadState() {
    fetch(this.props.serverPath+"/api")
    .then(fetchError)
    .then(response => response.json())
    .then(response => this.setState(response))
    .catch(err => this.context.reportConnError());
  }

  componentDidMount() {
    this.loadState()
    this.timer = setInterval(this.loadState.bind(this), 1000);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  configure(e, n) {
    e.preventDefault();
    let payload = JSON.stringify({
      failNext: n
    });
    fetch(this.props.serverPath+"/api", {
      method: "PUT",
      body: payload
    })
    .then(fetchError)
    .then(response => response.json())
    .then(response => this.setState(response))
    .catch(err => this.context.reportConnError());
  }

  render () {
    let probeDesc = null

    if (this.state.failNext == 0) {
      probeDesc = <span> Probe will permanently succeed </span>;
    } else if (this.state.failNext > 0) {
      probeDesc = <span> Probe will fail for next {this.state.failNext} calls</span>;
    } else {
      probeDesc = <span> Probe will permanently fail </span>;
    }

    let history = <p> No recorded probe history </p>
    if (this.state.history.length > 0) {
      let rows = [];
      for (let h of this.state.history) {
        rows.push(
          <tr key={h.id}>
            <td>{h.id}</td>
            <td>{h.when}</td>
            <td>{h.relWhen}</td>
            <td>{h.code}</td>
          </tr>
        )
      }
      history = (
        <table className="table table-condensed table-bordered">
          <thead>
            <tr>
              <th>ID</th><th colSpan="2">When</th><th>Status</th>
            </tr>
          </thead>
          <tbody>
            {rows}
          </tbody>
        </table>
      )
    }

    return (
      <div>
        <p>Probe is being served on <a href={this.props.serverPath}>{this.props.serverPath}</a></p>
        <p>{probeDesc}<br/>
           <span className="small">
             <a className="failn" onClick={e => this.configure(e, 0)} href="#">Succeed</a> | { " " }
             <a className="failn" onClick={e => this.configure(e, -1)} href="#">Fail</a> | { " " }
             Fail for next N calls: { " " }
             <a className="failn" onClick={e => this.configure(e, 1)} href="#">1</a> { " " }
             <a className="failn" onClick={e => this.configure(e, 2)} href="#">2</a> { " " }
             <a className="failn" onClick={e => this.configure(e, 3)} href="#">3</a> { " " }
             <a className="failn" onClick={e => this.configure(e, 5)} href="#">5</a> { " " }
             <a className="failn" onClick={e => this.configure(e, 10)} href="#">10</a>
          </span>
        </p>
        {history}
      </div>
    )
  }
}

Probe.propTypes =  {
  serverPath: React.PropTypes.string.isRequired
}

Probe.contextTypes = {
  reportConnError: React.PropTypes.func
};

