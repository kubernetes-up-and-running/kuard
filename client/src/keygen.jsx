import React from 'react';
import Form from "react-jsonschema-form";
import fetchError from './fetcherror';

const schema = {
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "object",
  "properties": {
    "enable": {
      "title": "Enabled?",
      "type": "boolean"
    },
    "exitOnComplete": {
      "title": "Exit server on completion?",
      "type": "boolean"
    },
    "exitCode": {
      "title": "Exit code when exiting. 0 is success.",
      "type": "integer"
    },
    "numToGen": {
      "title": "Number of keys to generate. 0 is infinite.",
      "type": "integer"
    },
    "timeToRun": {
      "title": "Time to run, in seconds. 0 is infinite.",
      "type": "integer"
    },
    "memQServer": {
      "title": "Base URL of the MemQ server to draw from. Can be http://localhost:8080/memq/server.",
      "type": "string"
    },
    "memQQueue": {
      "title": "The Queue to pull work items from.",
      "type": "string"
    }
  }
};

const uiSchema = {
  enable: {
    classNames: "foo"
  }
}

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
    fetch(this.props.serverPath)
    .then(fetchError)
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
    })
    .catch(err => this.context.reportConnError());
  }

  componentDidMount() {
    this.loadState(true)
    this.timer = setInterval(this.loadState.bind(this, false), 1000);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  handleSubmit(event) {
    let payload = JSON.stringify(this.state.config);
    fetch(this.props.serverPath, {
      method: "PUT",
      body: payload
    })
    .then(fetchError)
    .then(response => response.json())
    .then(response => this.setState(previousState => {
      previousState = response
      return previousState
    }))
    .catch(err => this.context.reportConnError());
  }

  handleChange({formData}) {
    this.setState(state => {
      state.config = formData;
      return state
    })
  }

  render () {
    let history = <p>No recorded workload history</p>
    if (this.state.history.length > 0) {
      let historyItems = [];
      for (let h of this.state.history) {
        historyItems.push(<span key={h.id}>{h.data}{"\n"}</span>)
      }
      historyItems.reverse()
      history = (<pre>{historyItems}</pre>)
    }

    return (
      <div>
        <div className="panel panel-default">
          <div className="panel-heading">KeyGen Synthetic Workload</div>
          <div className="panel-body">
            <div>This controls a synthetic workload on the server: creating 4096
                 bit RSA key pairs.  These parameters control how many to create
                 and, optionally, cause the server to exit with a specific exit
                 code.
            </div>
            <Form
              schema={schema}
              uiSchema={uiSchema}
              className="form"
              formData={this.state.config}
              onChange={this.handleChange}
              onSubmit={this.handleSubmit}>
              <input
                className="btn btn-default"
                type="submit"
                value="Submit" />
            </Form>
          </div>
        </div>
        {history}
      </div>
    )
  }
}

KeyGen.propTypes =  {
  serverPath: React.PropTypes.string.isRequired,
}

KeyGen.contextTypes = {
  reportConnError: React.PropTypes.func
};


module.exports = KeyGen;
