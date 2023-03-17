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
  const [loader, setLoader] = useState({});
  const [error, setError] = useState(null);
  const [alert, setAlert] = useState(null);
  const [client, setClient] = useState({
    Name: "",
    Quota: 0,
    UsedSpace: 0,
  });
  const [clients, setClients] = useState([]);
  const [totalQuota, setTotalQuota] = useState(null);
  const [usedSpace, setUsedSpace] = useState(null);
  const [storageSpace, setStorageSpace] = useState(null);
  const [token, setToken] = useState(null);

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

  const loadingHandler = (source, loading_type) => {
    var l = loader;
    console.log("source: " + source);
    console.log("loading_type: " + loading_type);
    console.log(l);
    if(loading_type === "push") {
      l[source] = "";
    } else if (loading_type === "pop") {
      delete l[source];
    }
    setLoader(l);
    console.log(l);
    if (Object.keys(l).length === 0) {
      setIsLoading(false);
    } else {
      setIsLoading(true);
    }
  }

  return (
    <>
      <AuthProvider token={token} setToken={setToken} loadingHandler={loadingHandler}>  
        <Navigation />

        <Container fluid>
          <Row className="justify-content-md-center">
            <ErrorMessage message={alert} err={error} />
            <Col md="90%">
              <Routes>
                <Route index element={<Login
                  loadingHandler={loadingHandler}  />} />
                <Route path="/admin/login" element={<Login
                  loadingHandler={loadingHandler}  />} />
                <Route path="/admin/clients">
                  <Route index element={<Clients 
                    error={error} 
                    setError={setError} 
                    setAlert={setAlert} 
                    loadingHandler={loadingHandler} 
                    clients={clients} 
                    setClients={wrapperSetClients} 
                    setClient={wrapperSetClient}
                    totalQuota={totalQuota} 
                    setTotalQuota={setTotalQuota}
                    usedSpace={usedSpace} 
                    setUsedSpace={setUsedSpace}
                    storageSpace={storageSpace} 
                    setStorageSpace={setStorageSpace} />} />
                  <Route path="new" element={<ClientForm 
                    client={client} 
                    setClient={wrapperSetClient} 
                    loadingHandler={loadingHandler} 
                    setError={setError} />} />
                  <Route path=":id" element={<ClientInfo 
                    client={client} />} />
                  <Route path=":id/edit" element={<ClientForm 
                    client={client} 
                    setClient={wrapperSetClient} 
                    loadingHandler={loadingHandler} 
                    setError={setError} />} />
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
