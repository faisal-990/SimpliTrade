import {  Route, Routes } from "react-router-dom";
import { LoginForm } from "./components/ui/login-form";
import { SignUp } from "./components/selfMade/signup";
export default function App() {
   return (
       <Routes>
        <Route path="/login" element={
            <LoginForm />
        }/>
       <Route path="/signup" element ={
            <SignUp/>
       }/>
       </Routes>
   ) 
}
