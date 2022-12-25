import React, { useState } from 'react';
import Navigation from './Navigation';
import Clients from './Clients';
import ClientForm from './ClientForm';
import ClientInfo from './ClientInfo';
import Login from './Login';
import AuthProvider from './AuthProvider';
import ErrorMessage from './ErrorMessage';
import { Routes, Route } from 'react-router-dom';
import { Container, Row, Col } from "react-bootstrap";
import Loading from './Loading';
// import './App.css';


function App() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [alert, setAlert] = useState(null);
  const [clients, setClients] = useState([]);
  const [selectedClient, setSelectedClient] = useState(null);
  const [storageSize, setStorageSize] = useState(0);
  const [storageSizeLabel, setStorageSizeLabel] = useState("");

  // Only checking for token seems weak. By manually entering any token to local storage
  // the app thinks you are authenticated and authorizes the main view.
  // Probably solved by using an additional flag 'isLoggedIn' that gets set (unset) on
  // login (logout)
  // if (!token || !isLoggedIn) {
  //   return (
  //     <>
  //       <Container fluid>
  //         <Row className="justify-content-md-center">
  //           <Col md="auto"> <Login setToken={setToken} setIsLoggedIn={setIsLoggedIn} setIsLoading={setIsLoading} /> </Col>
  //         </Row>
  //       </Container>
  //       <Loading isLoading={isLoading} />
  //     </>
  //   )
  // }

  return (
    <>
      <AuthProvider>  
        <Navigation />

        <Container fluid>
          <Row className="justify-content-md-center">
            <ErrorMessage message={alert} err={error} />
            <Col md="auto">
              <Routes>
                <Route index element={<Login />} />
                <Route path="/admin/login" element={<Login />} />
                <Route path="/admin/clients">
                  <Route index element={<Clients error={error} setError={setError} setAlert={setAlert} setIsLoading={setIsLoading} clients={clients} setClients={setClients} setStorageSize={setStorageSize} setStorageSizeLabel={setStorageSizeLabel} />} />
                  <Route path="new" element={<ClientForm clients={clients} setClients={setClients} setError={setError} setIsLoading={setIsLoading} client={selectedClient} setSelectedClient={setSelectedClient} storageSize={storageSize} storageSizeLabel={storageSizeLabel} />} />
                  <Route path=":id" element={<ClientInfo setIsLoading={setIsLoading} />} />
                  <Route path=":id/edit" element={<ClientForm clients={clients} setClients={setClients} client={selectedClient} setSelectedClient={setSelectedClient} storageSize={storageSize} storageSizeLabel={storageSizeLabel} />} />
                </Route>
              </Routes>
            </Col>
          </Row>
        </Container>
        <Loading isLoading={isLoading} />
      </AuthProvider>
    </>
  );
}

export default App;
