import React, { useState } from 'react';
import Navigation from './Navigation';
import Clients from './Clients';
import Login from './Login';
import { Container, Row, Col } from 'react-bootstrap';
// import './App.css';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [token, setToken] = useState(localStorage.getItem('basetoken') || '');

  const logout = () => {
    localStorage.setItem('basetoken', '');
    setToken('');
    setIsLoggedIn(false);
  }

  // Only checking for token seems weak. By manually entering any token to local storage
  // the app thinks you are authenticated and authorizes the main view.
  // Probably solved by using an additional flag 'isLoggedIn' that gets set (unset) on
  // login (logout)
  if (!token || !isLoggedIn) {
    return (
      <Container fluid>
        <Row className="justify-content-md-center">
          <Col md="auto"> <Login setToken={setToken} setIsLoggedIn={setIsLoggedIn} /> </Col>
        </Row>
      </Container>
    )
  }

  return (
    <Container fluid>
      <Navigation logout={logout} />
      <Row className="justify-content-md-center">
        <Clients token={token} />
      </Row>
    </Container>
  );
}

export default App;
