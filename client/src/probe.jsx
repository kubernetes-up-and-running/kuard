import React from 'react';
import Details from './details'

class Probe extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      probePath: 'foo',
      failNext: 0,
      history: []
    };
  }

  loadState() {
    fetch(this.props.path+"/api")
    .then(response => response.json())
    .then(response => this.setState(response));
  }

  componentDidMount() {
    this.loadState()
    this.timer = setInterval(this.loadState.bind(this), 1000);
  }

  componentWillUnmount() {
    this.clearInterval(this.timer);
  }

  configure(e, n) {
    e.preventDefault();
    var payload = JSON.stringify({
      failNext: n
    });
    fetch(this.props.path+"/api", {
      method: "PUT",
      body: payload
    })
    .then(response => response.json())
    .then(response => this.setState(response));
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
      for (var h of this.state.history) {
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
        <table>
          <thead>
            <tr>
              <th>ID</th><th>When</th><th></th><th>Status</th>
            </tr>
          </thead>
          <tbody>
            {rows}
          </tbody>
        </table>
      )
    }

    return (
      <Details title={this.props.title}>
        <p>Probe is being served on <a href={this.props.path}>{this.props.path}</a></p>
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
      </Details>
    )
  }
}

Probe.propTypes =  {
  title: React.PropTypes.string.isRequired,
  path: React.PropTypes.string.isRequired,
  open: React.PropTypes.bool
}

Probe.defaultProps = {
  open: false
}


module.exports = Probe;
