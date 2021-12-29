import React from 'react';
import Navigation from './Navigation';
import Clients from './Clients';
import Form from './Form';
import { Container } from 'react-bootstrap';
// import './App.css';

class App extends React.Component {
  state = {
    error: null,
    isLoaded: false,
    clients: [],
  }

  deleteClient = (index) => {
    const {clients} = this.state

    this.setState({
      clients: clients.filter((clients, i) => {
        return i !== index
      }),
    })
  }

  createClient = () => {
    const {clients} = this.state

    this.setState()
  }

  componentDidMount() {
    fetch("/clients")
      .then((result) => result.json())
      .then((result) => {
          this.setState({
            isLoaded: true,
            clients: result,
          });
        },
        (error) => {
          this.setState({
            isLoaded: true,
            error,
          });
        }
      )
  } 

  render() {
    const {error, isLoaded, clients} = this.state;
    
    if (error) {
      return <div>Error: {error.message}</div>
    } else if (!isLoaded) {
      return <div>Loading...</div>
    } else {
      return (
        <Container fluid>
          <Navigation />
          <Clients clientData={clients} deleteClient={this.deleteClient} />
          <button onClick={() => console.log("not implemented")}>Add new client</button>
          <Form />
        </Container>
      )
    }
  }
}

export default App;
