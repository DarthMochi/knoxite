import React, { useCallback, useState } from 'react';
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


function App() {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);
  const [alert, setAlert] = useState(null);
  const [client, setClient] = useState({
    Name: "",
    Quota: 0,
    UsedSpace: 0,
  });
  const [clients, setClients] = useState([]);
  const [token, setToken] = useState(null);
  const [storageSizeInBytes, setStorageSizeInBytes] = useState(0);
  const [storageSizeLabel, setStorageSizeLabel] = useState("");

  const wrapperSetClient = useCallback(val => {
    var c = {
      ID: val ? val.ID : 0,
      Name: val ? val.Name : "",
      Quota: val ? val.Quota : 0,
      UsedSpace: val ? val.UsedSpace : 0,
      AuthCode: val ? val.AuthCode : "",
    };
    setClient(c);
  }, [setClient]);

  const wrapperSetClients = useCallback(val => {
    setClients(val);
  }, [setClients]);

  return (
    <>
      <AuthProvider token={token} setToken={setToken}>  
        <Navigation token={token} />

        <Container fluid>
          <Row className="justify-content-md-center">
            <ErrorMessage message={alert} err={error} />
            <Col md="90%">
              <Routes>
                <Route index element={<Login
                  setIsLoading={setIsLoading}  />} />
                <Route path="/admin/login" element={<Login
                  setIsLoading={setIsLoading}  />} />
                <Route path="/admin/clients">
                  <Route index element={<Clients 
                    error={error} 
                    setError={setError} 
                    setAlert={setAlert} 
                    setIsLoading={setIsLoading} 
                    clients={clients} 
                    setClients={wrapperSetClients} 
                    storageSizeLabel={storageSizeLabel} 
                    setStorageSizeLabel={setStorageSizeLabel} 
                    setClient={wrapperSetClient}
                    setStorageSizeInBytes={setStorageSizeInBytes}
                    storageSizeInBytes={storageSizeInBytes} />} />
                  <Route path="new" element={<ClientForm 
                    client={client} 
                    setClient={wrapperSetClient} 
                    setIsLoading={setIsLoading} 
                    setError={setError} 
                    storageSizeInBytes={storageSizeInBytes} 
                    storageSizeLabel={storageSizeLabel} />} />
                  <Route path=":id" element={<ClientInfo 
                    client={client} />} />
                  <Route path=":id/edit" element={<ClientForm 
                    client={client} 
                    setClient={wrapperSetClient} 
                    setIsLoading={setIsLoading} 
                    storageSizeInBytes={storageSizeInBytes}
                    storageSizeLabel={storageSizeLabel} />} />
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
