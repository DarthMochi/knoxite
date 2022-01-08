import React, { useState, useEffect, useRef } from "react";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrashAlt, faEdit } from '@fortawesome/free-solid-svg-icons';
import { Container, Table, Button } from "react-bootstrap";
import ErrorMessage from "./ErrorMessage";
import Loading from "./Loading";
import ClientForm from "./ClientForm";

const Clients = (props) => {
    const [clients, setClients] = useState([]);
    const [selectedClient, setSelectedClient] = useState(null);
    const [isLoading, setIsLoading] = useState(false);
    const [isFormMode, setIsFormMode] = useState(false);
    const [error, setError] = useState(null);
    const [alert, setAlert] = useState(null);
    const isMounted = useRef(false);
    const token = props.token;

    useEffect(() => {
        isMounted.current = true;
        setIsLoading(true);
        setError(false);

        const fetchUrl = "/clients";
        const fetchOptions = {
            headers: {
                'Authorization': 'Basic ' + token,
            },
        };
        fetchData(fetchUrl, fetchOptions)
        .then(result => result.json())
        .then(
            (result) => {
                setClients(result);
                setIsLoading(false);
            },
            (err) => {
                console.log(err);
                setError(err);
                setAlert(err.message);
            }
        );
        setIsLoading(false);

        return () => {
            isMounted.current = false;
        };
    }, [token]);

    // TODO: Exctract `getClients` from useEffect as a function
    // TODO: Move all API calls out of the component into separate functions.

    const deleteClient = (index) => {
        const fetchUrl = "/clients/" + clients[index].ID;
        const fetchOptions = {
            method: 'DELETE',
            headers: {
                'Authorization': 'Basic ' + token,
            },
        };
        fetchData(fetchUrl, fetchOptions)
        .catch((err) => {
            console.log(err);
        });
        setClients(clients.filter((_, i) => {
            return i !== index;
          }));
    };

    const createClient = (data) => {
        // let currentDate = new Date().toLocaleDateString("en-US", {day: "2-digit", month: "2-digit", hour: "numeric", minute: "numeric", second: "numeric"});
        // data = 'react-testclient ' + currentDate;
        setIsLoading(true);
        let newClientID = '';

        const fetchUrl = "/clients";
        const fetchOptions = {
            method: 'POST',
            headers: {
                'Authorization': 'Basic ' + token,
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: "name=" + data,
        };
        fetchData(fetchUrl, fetchOptions)
        .then(result => {
            newClientID = result.headers.get("Location").slice(-2);
            console.log("client id is ", newClientID);
            return fetchClient(token, newClientID);
        })
        .then(
            (result) => {
                console.log("called the new client: ", result);
                setClients(clients => [...clients, result]);
                setIsLoading(false);
                setAlert("success");
                setError(false);
            },
            (err) => {
                console.log(err);
                setError(err);
                setAlert(err.message);
            }
        );
        setIsLoading(false);
        console.log("clients are:", clients);
    };

    const updateClient = (data) => {
        const url = "/clients/" + selectedClient.ID;
        console.log(url);
        const options = {
            method: 'PUT',
            headers: {
                'Authorization': 'Basic ' + token,
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: "name=" + data,
        };
        fetchData(url, options)
        .then(response => {
            if(response.status == 200) {
                setClients(clients.map((client, index) => {
                    if (selectedClient.ID === client.ID) {
                        client.Name = data;
                    }
                    return client;
                }));
            }
            
        }, err => {
            console.log(err);
        });
    }

    const toggleClientForm = () => {
        setIsFormMode(!isFormMode);
        setSelectedClient(null);
    }

    const editClient = (client) => {
        setSelectedClient(client);
        setIsFormMode(!isFormMode);
    }

    if (isLoading) {
      return <Loading />;
    } else if (isFormMode) {
        return (
            <ClientForm reset={toggleClientForm} createClient={createClient} updateClient={updateClient} client={selectedClient} setSelectedClient={setSelectedClient} />
        );
    } else {
        return (
            <Container fluid>
                <ErrorMessage message={alert} err={error} />
                <Table hover size="sm">
                    <TableHeader />
                    <TableBody clientData={clients} deleteClient={deleteClient} editClient={editClient} token={token} />
                </Table>
                <Button variant="light" onClick={toggleClientForm}>
                    <FontAwesomeIcon icon={faEdit} /> Create new client
                </Button>
            </Container>
        );
    }
};

const TableHeader = () => {
    return (
        <thead>
            <tr>
                <th>ID</th>
                <th>Name</th>
                <th>AuthCode</th>
                <th></th>
            </tr>
        </thead>
    );
};

const TableBody = (props) => {
    const clients = props.clientData.map((client, index) => {
        return (
            <tr key={index}>
                <td>{client.ID}</td>
                <td>{client.Name}</td>
                <td>{client.AuthCode}</td>
                <td>
                    <Button variant="light" onClick={() => props.deleteClient(index)}>
                        <FontAwesomeIcon icon={faTrashAlt} />
                    </Button>{' '}
                    <Button variant="light" onClick={() => props.editClient(client)}>
                        <FontAwesomeIcon icon={faEdit} />
                    </Button>
                </td>
            </tr>
        );
    });
  return <tbody>{clients}</tbody>;
};

const fetchClient = async (token, id) => {
    const url = "/clients/" + id;
    const options = {
        headers: {
            'Authorization': 'Basic ' + token,
        },
    };
    const response = await fetchData(url, options);
    return await response.json();
};

const fetchData = async (url, options) => {
    const response = await fetch(url, options);
    if (!response.ok) {
        throw new Error(`Error fetching data. Server replied ${response.status}`);
    }
    console.log(response);
    return response;
};

export default Clients;
