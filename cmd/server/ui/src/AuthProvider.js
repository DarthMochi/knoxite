import React, { useState, createContext, useContext, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { fetchData } from "./utils.js";

const AuthContext = createContext(null);

const AuthProvider = ({client, children }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [token, setToken] = useState(null);

  useEffect(() => {
    const token = localStorage.getItem("token");
    if(token) {
      setToken(token);
      if(!location.pathname.includes("/admin")) {
        navigate("/admin/clients");
      }
    }
  }, [navigate, token, location]);

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
        if (response.ok) {
          setToken(userToken);
          localStorage.setItem("token", userToken);
          navigate("/admin/clients");
        }
      },
      err => {
        console.log("Error logging in:", err);
      }
    );
  };

  const handleLogout = () => {
    setToken(null);
    localStorage.removeItem("token");
    localStorage.removeItem("client");
    navigate("/admin/login");
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
