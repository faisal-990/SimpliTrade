import { Routes, Route } from "react-router-dom";
import Login from "./Pages/Login";
import TradingDashboard from "./Pages/demo";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<TradingDashboard/>} />
      <Route path="/login" element={<Login />} />
    </Routes>
  );
}

