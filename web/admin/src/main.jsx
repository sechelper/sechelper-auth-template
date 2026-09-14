import { createRoot } from "react-dom/client";
import { AdminApp } from "./app/AdminApp.jsx";
import { AuthProvider } from "./modules/auth/AuthProvider.jsx";
import "./style.css";
createRoot(document.getElementById("root")).render(<AuthProvider><AdminApp /></AuthProvider>);
