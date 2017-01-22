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
      for (var hi of this.state.history) {
        rows.push(
          <tr key="{hi.when}">
            <td>{hi.when}</td>
            <td>{hi.relWhen}</td>
            <td>{hi.code}</td>
          </tr>
        )
      }
      history = (
        <table>
          <tr>
            <th>When</th><th></th><th>Status</th>
          </tr>
          {rows}
        </table>
      )
    }

    return (
      <Details title={this.props.path}>
        <p>Probe is being served on <a href="{this.props.path}">{this.props.path}</a></p>
        <p>{probeDesc}<br/>
           <span className="small">
             <a className="failn" href="#">Succeed</a> |
             <a className="failn" href="#">Fail</a> |
             Fail for next N calls:
             <a className="failn" href="#">1</a>
             <a className="failn" href="#">2</a>
             <a className="failn" href="#">3</a>
             <a className="failn" href="#">5</a>
             <a className="failn" href="#">10</a>
          </span>
        </p>
        {history}
      </Details>
    )
  }
}

Probe.propTypes =  {
  path: React.PropTypes.string.isRequired,
  open: React.PropTypes.bool
}

Probe.defaultProps = {
  open: false
}


module.exports = Probe;
