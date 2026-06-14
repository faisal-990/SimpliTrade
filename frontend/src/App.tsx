import { Navigate, Route, Routes } from "react-router-dom";
import { RequireAuth } from "@/components/RequireAuth";
import { AppLayout } from "@/components/layout/AppLayout";
import Login from "@/pages/Login";
import Signup from "@/pages/Signup";
import ForgotPassword from "@/pages/ForgotPassword";
import OAuthCallback from "@/pages/OAuthCallback";
import Profile from "@/pages/Profile";
import Dashboard from "@/pages/Dashboard";
import StockDetail from "@/pages/StockDetail";
import Portfolio from "@/pages/Portfolio";
import Analytics from "@/pages/Analytics";
import Investors from "@/pages/Investors";
import InvestorDetail from "@/pages/InvestorDetail";
import CreateInvestor from "@/pages/CreateInvestor";
import Feed from "@/pages/Feed";

export default function App() {
  return (
    <Routes>
      {/* public */}
      <Route path="/" element={<Navigate to="/login" replace />} />
      <Route path="/login" element={<Login />} />
      <Route path="/signup" element={<Signup />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/oauth/callback" element={<OAuthCallback />} />

      {/* authenticated app */}
      <Route element={<RequireAuth />}>
        <Route path="/app" element={<AppLayout />}>
          <Route index element={<Navigate to="/app/dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="stock/:symbol" element={<StockDetail />} />
          <Route path="portfolio" element={<Portfolio />} />
          <Route path="analytics" element={<Analytics />} />
          <Route path="investors" element={<Investors />} />
          <Route path="investors/new" element={<CreateInvestor />} />
          <Route path="investors/:id" element={<InvestorDetail />} />
          <Route path="feed" element={<Feed />} />
          <Route path="profile" element={<Profile />} />
        </Route>
      </Route>

      {/* fallback */}
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  );
}
