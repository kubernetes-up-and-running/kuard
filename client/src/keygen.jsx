import React from 'react';
import Details from './details'

class KeyGen extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      config: {
        enable: false,
        numToGen: 0,
        timeToRun: 0,
        exitOnComplete: false,
        exitCode: 0
      },
      history: []
    };

    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  loadState(initial) {
    fetch(this.props.path)
    .then(response => response.json())
    .then(response => {
      if(initial) {
        this.setState(response);
      } else {
        this.setState(state => state.history = response.history)
      }
    });
  }

  componentDidMount() {
    this.loadState(true)
    this.timer = setInterval(this.loadState.bind(this, false), 1000);
  }

  componentWillUnmount() {
    this.clearInterval(this.timer);
  }

  handleChange(event) {
    const target = event.target;
    const k = target.name;
    const v = target.type === 'checkbox' ? target.checked : target.value;
    this.setState(previousState => {
      previousState.config[k] = v;
      return previousState;
    })
  }

  handleSubmit(event) {
    event.preventDefault();

    let payload = JSON.stringify(this.state.config);
    fetch(this.props.path, {
      method: "PUT",
      body: payload
    })
    .then(response => response.json())
    .then(response => this.setState(previousState => {
      previousState = response
      return previousState
    }));
  }

  render () {
    let history = <p> No recorded workload history </p>
    if (this.state.history.length > 0) {
      history = [];
      for (let h of this.state.history) {
        history.push(<pre key={h.id}>{h.data}</pre>)
      }
      history.reverse()
    }

    return (
      <Details title="Key Generation Artificial Workload" open={this.props.open}>
        <form autoComplete="off" onSubmit={this.handleSubmit}>
          <label>Enabled <input
              type="checkbox"
              name="enable"
              checked={this.state.config.enable}
              onChange={this.handleChange}
            /></label> { "" }
          <label>Number To Generate <input
              type="text"
              name="numToGen"
              value={this.state.config.numToGen}
              onChange={this.handleChange}
            /></label> { " " }
          <label>Runtime (s) <input
              type="text"
              name="timeToRun"
              value={this.state.config.timeToRun}
              onChange={this.handleChange}
            /></label> { " " }
          <label>Exit when done <input
              type="checkbox"
              name="exitOnComplete"
              checked={this.state.config.exitOnComplete}
              onChange={this.handleChange}
            /></label> { "" }
          <label>Exit code <input
              type="text"
              name="exitCode"
              value={this.state.config.exitCode}
              onChange={this.handleChange}
            /></label> { " " }
          <input
            type="submit"
            value="Submit" />
        </form>
        {history}
      </Details>
    )
  }
}

KeyGen.propTypes =  {
  path: React.PropTypes.string.isRequired,
  open: React.PropTypes.bool
}

KeyGen.defaultProps = {
  open: false
}


module.exports = KeyGen;
