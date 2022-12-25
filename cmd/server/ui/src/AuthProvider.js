import React, { useState, createContext, useContext } from "react";
import { useNavigate } from "react-router-dom";
import { fetchData } from "./utils.js";

const AuthContext = createContext(null);

const AuthProvider = ({ children }) => {
  const navigate = useNavigate();
  const [token, setToken] = useState(null);

  const handleLogin = (event) => {
    event.preventDefault();
    var username = event.target[0].value;
    var password = event.target[1].value;
    var bcrypt = require('bcryptjs');
    var hash = bcrypt.hashSync(password, 14);  // has to be 14 (why?)
    const userToken = btoa(username + ':' + hash);

    const fetchOptions = {
      method: 'POST',
      headers: {
          'Authorization': 'Basic ' + userToken,
      },
    }
    fetchData("/login", fetchOptions)
    .then(
      response => {
        if (response.status === 200) {
          setToken(userToken);
          navigate("/admin/clients");
        }
      },
      err => {
        console.log("Error logging in:", err);
      }
    );
    }

  const handleLogout = () => {
    setToken(null);
    navigate("/login");
  };

  const value = {
    token: token,
    onLogin: handleLogin,
    onLogout: handleLogout,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  return useContext(AuthContext);
}

export default AuthProvider;
