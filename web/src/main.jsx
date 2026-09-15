import { createRoot } from "react-dom/client";
import { App } from "./framework/app/App.jsx";
import { AuthProvider } from "./framework/auth/AuthProvider.jsx";
import "./framework/style.css";

createRoot(document.getElementById("root")).render(<AuthProvider><App /></AuthProvider>);
