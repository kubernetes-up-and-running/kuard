import React from 'react';
import Details from './details'
import Form from "react-jsonschema-form";

const schema = {
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "enable": {
      "type": "boolean"
    },
    "numToGen": {
      "type": "integer"
    },
    "timeToRun": {
      "type": "integer"
    },
    "exitOnComplete": {
      "type": "boolean"
    },
    "exitCode": {
      "type": "integer"
    }
  }
};

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
        this.setState(state => {
          state.history = response.history
          return state
        })
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

  handleSubmit(event) {
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

  handleChange({formData}) {
    this.setState(state => {
      state.config = formData;
      return state
    })
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
        <Form
          schema={schema}
          formData={this.state.config}
          onChange={this.handleChange}
          onSubmit={this.handleSubmit}>
        </Form>
        <div>
        {history}
        </div>
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
