import { createRoot } from "react-dom/client";
import { App } from "./app/App.jsx";
import { AuthProvider } from "./modules/auth/AuthProvider.jsx";
import "./style.css";

createRoot(document.getElementById("root")).render(<AuthProvider><App /></AuthProvider>);
