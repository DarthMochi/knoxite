import React, { useState } from 'react';
import { Button, Card, Form, Container } from 'react-bootstrap';
import ErrorMessage from './ErrorMessage';

function Login(props) {
    const setToken = props.setToken;
    const setIsLoggedIn = props.setIsLoggedIn;

    const [username, setUserName] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState(null);

    const handleLogin = (event) => {
        event.preventDefault();
        setIsLoggedIn(false);

        var bcrypt = require('bcryptjs');
        var hash = bcrypt.hashSync(password, 14);  // has to be 14 (why?)
        const userToken = btoa(username + ':' + hash);
        
        /* try {
            let response = await fetch("/testUser", {
                headers: {
                    'Authorization': 'Basic ' + btoa(username + ':' + password),
                },
            });
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            if (response.status === 200) {
                localStorage.setItem('basetoken', userToken);
                setToken(userToken);
                setIsLoggedIn(true);
            }
        } catch(err) {
            setError(err);
            console.log(err);
        } */

        const fetchOptions = {
            headers: {
                'Authorization': 'Basic ' + userToken,
            },
        }
        fetchData("/login", fetchOptions)
        .then(
            response => {
                if (response.status === 200) {
                    localStorage.setItem('basetoken', userToken);
                    setToken(userToken);
                    setIsLoggedIn(true);
                }
            },
            err => {
                console.log("Error logging in:", err);
                setError(err);
            }
        );
    }

    return (
        <Container>
            <Card>
                <Card.Header>Login</Card.Header>
                <Card.Body>
                <Form onSubmit={handleLogin}>
                    <Form.Group className="mb-3" controlId="formBasicName">
                        <Form.Label>Name</Form.Label>
                        <Form.Control type="text" placeholder="Username" name="user" onChange={e => setUserName(e.target.value)} />
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="formBasicPassword">
                        <Form.Label>Password</Form.Label>
                        <Form.Control type="password" placeholder="Password" name="passwd" onChange={e => setPassword(e.target.value)} />
                    </Form.Group>
                    <Button variant="light" type="submit" >
                        Submit
                    </Button>
                </Form>
                </Card.Body>
            </Card>
            <ErrorMessage error={error} />
        </Container>
    );
}

const fetchData = async (url, options) => {
    const response = await fetch(url, options);
    if (!response.ok) {
        throw new Error(`Error fetching data. Server replied ${response.status}`);
    }
    console.log(response);
    return response;
}

/* class LoginForm extends React.Component {
    state = {
        user: '',
        passwd: '',
    };

    handleChange = (event) => {
        this.setState({
            [event.target.name]: event.target.value
        });
    }

    handleSubmit = (event, props) => {
        alert('Submitted value: ' + this.state.user + ' ' + this.state.passwd);
        props.login();
        event.preventDefault();
    }

    render() {
        return (
            <Card>
                <Card.Header>Login</Card.Header>
                <Card.Body>
                <Form onSubmit={this.handleSubmit}>
                    <Form.Group className="mb-3" controlId="formBasicName">
                        <Form.Label>Name</Form.Label>
                        <Form.Control type="text" placeholder="Username" name="user" onChange={this.handleChange} />
                    </Form.Group>
                    <Form.Group className="mb-3" controlId="formBasicPassword">
                        <Form.Label>Password</Form.Label>
                        <Form.Control type="password" placeholder="Password" name="passwd" onChange={this.handleChange} />
                    </Form.Group>
                    <Button variant="light" type="submit" >
                        Submit
                    </Button>
                </Form>
                </Card.Body>
            </Card>
        )
    }

    render() {
        return (
            <form onSubmit={this.handleSubmit}>
                <label>
                    Name:
                    <input type="text" value={this.state.name} onChange={this.handleChange} />
                </label>
                <input type="submit" value="Submit" />
            </form>
        );
    }
} */

export default Login;