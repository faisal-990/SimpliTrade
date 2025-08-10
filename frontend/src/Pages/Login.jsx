import { useState } from "react";
import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { TrendingUp, Eye, EyeOff, Mail, Lock, User } from "lucide-react";

export default function Login() {
    const [isLogin, setIsLogin] = useState(true);
    const [showPassword, setShowPassword] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [formData, setFormData] = useState({
        email: "",
        password: "",
        confirmPassword: "",
        firstName: "",
        lastName: "",
    });

    const handleInputChange = (e) => {
        setFormData((prev) => ({
            ...prev,
            [e.target.name]: e.target.value,
        }));
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsLoading(true);
        // Simulate loading
        await new Promise((resolve) => setTimeout(resolve, 1500));
        setIsLoading(false);
        console.log("Form submitted:", formData);
    };

    const handleModeSwitch = () => {
        setIsLogin(!isLogin);
        setFormData({
            email: "",
            password: "",
            confirmPassword: "",
            firstName: "",
            lastName: "",
        });
    };

    return (
        <div className="min-h-screen bg-background flex items-center justify-center p-4 relative overflow-hidden">
            {/* Animated background elements */}
            <div className="absolute inset-0 overflow-hidden">
                <div className="absolute -top-40 -right-40 w-80 h-80 bg-primary/10 rounded-full blur-3xl animate-pulse"></div>
                <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-primary/5 rounded-full blur-3xl animate-pulse delay-700"></div>
            </div>

            <div className="w-full max-w-md relative z-10">
                {/* Logo Section */}
                <div className="text-center mb-8 animate-fade-in">
                    <div className="inline-flex items-center gap-3 mb-4 group">
                        <div className="w-12 h-12 bg-primary rounded-xl flex items-center justify-center transition-all duration-300 group-hover:scale-110 group-hover:rotate-3">
                            <TrendingUp className="h-6 w-6 text-primary-foreground transition-transform duration-300" />
                        </div>
                        <div className="text-left">
                            <h1 className="text-2xl font-bold text-foreground">SimpliTrade</h1>
                            <p className="text-sm text-muted-foreground">
                                Where investing meets ideology
                            </p>
                        </div>
                    </div>
                </div>

                <Card className="border-border bg-card/80 backdrop-blur-sm shadow-2xl animate-scale-in">
                    <CardHeader className="text-center pb-4">
                        <CardTitle className="text-2xl font-bold text-card-foreground transition-all duration-300">
                            {isLogin ? "Welcome back" : "Create account"}
                        </CardTitle>
                        <CardDescription className="text-muted-foreground transition-all duration-300">
                            {isLogin
                                ? "Sign in to your account to continue"
                                : "Join thousands of smart investors"}
                        </CardDescription>
                    </CardHeader>

                    <CardContent className="space-y-6">
                        <form onSubmit={handleSubmit} className="space-y-4">
                            {/* Name fields for signup */}
                            <div
                                className={`grid grid-cols-2 gap-4 transition-all duration-500 ease-in-out ${!isLogin
                                        ? "opacity-100 max-h-20 animate-fade-in"
                                        : "opacity-0 max-h-0 overflow-hidden"
                                    }`}
                            >
                                <div className="space-y-2">
                                    <Label
                                        htmlFor="firstName"
                                        className="text-sm font-medium text-foreground flex items-center gap-2"
                                    >
                                        <User className="h-3 w-3" />
                                        First name
                                    </Label>
                                    <Input
                                        id="firstName"
                                        name="firstName"
                                        type="text"
                                        value={formData.firstName}
                                        onChange={handleInputChange}
                                        className="bg-background/50 border-input focus:border-primary focus:ring-primary transition-all duration-200 hover:bg-background/70"
                                        placeholder="John"
                                        required={!isLogin}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label
                                        htmlFor="lastName"
                                        className="text-sm font-medium text-foreground flex items-center gap-2"
                                    >
                                        <User className="h-3 w-3" />
                                        Last name
                                    </Label>
                                    <Input
                                        id="lastName"
                                        name="lastName"
                                        type="text"
                                        value={formData.lastName}
                                        onChange={handleInputChange}
                                        className="bg-background/50 border-input focus:border-primary focus:ring-primary transition-all duration-200 hover:bg-background/70"
                                        placeholder="Doe"
                                        required={!isLogin}
                                    />
                                </div>
                            </div>

                            {/* Email field */}
                            <div className="space-y-2 animate-fade-in">
                                <Label
                                    htmlFor="email"
                                    className="text-sm font-medium text-foreground flex items-center gap-2"
                                >
                                    <Mail className="h-3 w-3" />
                                    Email address
                                </Label>
                                <Input
                                    id="email"
                                    name="email"
                                    type="email"
                                    value={formData.email}
                                    onChange={handleInputChange}
                                    className="bg-background/50 border-input focus:border-primary focus:ring-primary transition-all duration-200 hover:bg-background/70 focus:scale-[1.02]"
                                    placeholder="john@example.com"
                                    required
                                />
                            </div>

                            {/* Password field */}
                            <div className="space-y-2 animate-fade-in">
                                <Label
                                    htmlFor="password"
                                    className="text-sm font-medium text-foreground flex items-center gap-2"
                                >
                                    <Lock className="h-3 w-3" />
                                    Password
                                </Label>
                                <div className="relative group">
                                    <Input
                                        id="password"
                                        name="password"
                                        type={showPassword ? "text" : "password"}
                                        value={formData.password}
                                        onChange={handleInputChange}
                                        className="bg-background/50 border-input focus:border-primary focus:ring-primary pr-10 transition-all duration-200 hover:bg-background/70 focus:scale-[1.02]"
                                        placeholder="••••••••"
                                        required
                                    />
                                    <button
                                        type="button"
                                        onClick={() => setShowPassword(!showPassword)}
                                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-all duration-200 hover:scale-110"
                                    >
                                        {showPassword ? (
                                            <EyeOff className="h-4 w-4" />
                                        ) : (
                                            <Eye className="h-4 w-4" />
                                        )}
                                    </button>
                                </div>
                            </div>

                            {/* Confirm password for signup */}
                            <div
                                className={`space-y-2 transition-all duration-500 ease-in-out ${!isLogin
                                        ? "opacity-100 max-h-20 animate-fade-in"
                                        : "opacity-0 max-h-0 overflow-hidden"
                                    }`}
                            >
                                <Label
                                    htmlFor="confirmPassword"
                                    className="text-sm font-medium text-foreground flex items-center gap-2"
                                >
                                    <Lock className="h-3 w-3" />
                                    Confirm password
                                </Label>
                                <Input
                                    id="confirmPassword"
                                    name="confirmPassword"
                                    type="password"
                                    value={formData.confirmPassword}
                                    onChange={handleInputChange}
                                    className="bg-background/50 border-input focus:border-primary focus:ring-primary transition-all duration-200 hover:bg-background/70 focus:scale-[1.02]"
                                    placeholder="••••••••"
                                    required={!isLogin}
                                />
                            </div>

                            {/* Forgot password link for login */}
                            <div
                                className={`text-right transition-all duration-500 ease-in-out ${isLogin
                                        ? "opacity-100 max-h-6 animate-fade-in"
                                        : "opacity-0 max-h-0 overflow-hidden"
                                    }`}
                            >
                                <button
                                    type="button"
                                    className="text-sm text-primary hover:text-primary/80 transition-all duration-200 hover:scale-105 story-link"
                                >
                                    Forgot password?
                                </button>
                            </div>

                            {/* Submit button */}
                            <Button
                                type="submit"
                                disabled={isLoading}
                                className="w-full bg-primary text-primary-foreground hover:bg-primary/90 transition-all duration-200 hover:scale-[1.02] hover:shadow-lg hover:shadow-primary/20 group"
                            >
                                {isLoading ? (
                                    <div className="flex items-center gap-2">
                                        <div className="w-4 h-4 border-2 border-primary-foreground/30 border-t-primary-foreground rounded-full animate-spin"></div>
                                        <span>Processing...</span>
                                    </div>
                                ) : (
                                    <span className="group-hover:scale-105 transition-transform duration-200">
                                        {isLogin ? "Sign in" : "Create account"}
                                    </span>
                                )}
                            </Button>
                        </form>

                        {/* Divider */}
                        <div className="relative animate-fade-in">
                            <div className="absolute inset-0 flex items-center">
                                <span className="w-full border-t border-border" />
                            </div>
                            <div className="relative flex justify-center text-xs uppercase">
                                <span className="bg-card px-2 text-muted-foreground">or</span>
                            </div>
                        </div>

                        {/* Toggle between login/signup */}
                        <div className="text-center animate-fade-in">
                            <p className="text-sm text-muted-foreground">
                                {isLogin
                                    ? "Don't have an account?"
                                    : "Already have an account?"}{" "}
                                <button
                                    type="button"
                                    onClick={handleModeSwitch}
                                    className="text-primary hover:text-primary/80 font-medium transition-all duration-200 hover:scale-105 story-link"
                                >
                                    {isLogin ? "Create account" : "Sign in"}
                                </button>
                            </p>
                        </div>

                        {/* Back to dashboard link */}
                        <div className="text-center pt-4 border-t border-border animate-fade-in">
                            <Link
                                to="/"
                                className="text-sm text-muted-foreground hover:text-foreground transition-all duration-200 hover:scale-105 inline-flex items-center gap-1 group"
                            >
                                <span className="group-hover:-translate-x-1 transition-transform duration-200">
                                    ←
                                </span>
                                <span>Back to dashboard</span>
                            </Link>
                        </div>
                    </CardContent>
                </Card>

                {/* Footer */}
                <div className="text-center mt-8 animate-fade-in">
                    <p className="text-xs text-muted-foreground">
                        By continuing, you agree to our{" "}
                        <a
                            href="#"
                            className="text-primary hover:text-primary/80 transition-all duration-200 hover:scale-105 story-link"
                        >
                            Terms of Service
                        </a>{" "}
                        and{" "}
                        <a
                            href="#"
                            className="text-primary hover:text-primary/80 transition-all duration-200 hover:scale-105 story-link"
                        >
                            Privacy Policy
                        </a>
                    </p>
                </div>
            </div>
        </div>
    );
}
