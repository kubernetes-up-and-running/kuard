import React from 'react';

class Env extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      commandLine: [],
      env: {}
    };
  }

  loadState() {
    fetch(this.props.apiPath)
    .then(response => response.json())
    .then(response => this.setState(response));
  }

  componentDidMount() {
    this.loadState()
  }


  render () {
    let args = [];
    for (let [idx, arg] of this.state.commandLine.entries()) {
      args.push(<code key={idx}>{arg}</code>)
      args.push(" ")
    }

    let rows = [];
    for (let k in this.state.env) {
      rows.push(
        <tr key={k}>
          <td><samp>{k}</samp></td>
          <td><samp>{this.state.env[k]}</samp></td>
        </tr>
      )
    }

    return (
      <div>
        <dl>
          <dt>Command Line</dt>
          <dd>{args}</dd>
        </dl>
        <table className="table table-condensed table-bordered">
          <thead>
            <tr>
              <th>Key</th><th>Value</th>
            </tr>
          </thead>
          <tbody>
            {rows}
          </tbody>
        </table>
      </div>
    )
  }
}

Env.propTypes =  {
  apiPath: React.PropTypes.string.isRequired,
}

module.exports = Env;
