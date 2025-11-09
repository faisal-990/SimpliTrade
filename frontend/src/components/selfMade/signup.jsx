import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
    CardFooter, // Added CardFooter for the sign-up link
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function SignUp({ className, ...props }) {
    return (
        // Outermost Wrapper: Sets the fixed width and centers the entire form horizontally
        <div
            className={cn("flex flex-col gap-6 mx-auto max-w-72", className)}
            {...props}
        >
            <Card>
                <CardHeader className="text-center">
                    <CardTitle className="text-xl">Create Account</CardTitle>
                    <CardDescription>
                        Enter your details below to create your account
                    </CardDescription>
                </CardHeader>

                <CardContent>
                    <form>
                        <div className="grid gap-4">
                            {/* Full Name Field */}
                            <div className="grid gap-2">
                                <Label htmlFor="name">Full Name</Label>
                                <Input id="name" placeholder="Max Robinson" required />
                            </div>

                            {/* Email Field */}
                            <div className="grid gap-2">
                                <Label htmlFor="email">Email</Label>
                                <Input
                                    id="email"
                                    type="email"
                                    placeholder="m@example.com"
                                    required
                                />
                            </div>

                            {/* Password Field */}
                            <div className="grid gap-2">
                                <Label htmlFor="password">Password</Label>
                                <Input id="password" type="password" required />
                            </div>

                            {/* Sign Up Button */}
                            <Button type="submit" className="w-full mt-2">
                                Sign Up
                            </Button>

                            {/* Google Sign Up Button */}
                            <div className="relative text-center text-sm after:absolute after:inset-0 after:top-1/2 after:z-0 after:flex after:items-center after:border-t after:border-border">
                                <span className="bg-card text-muted-foreground relative z-10 px-2">
                                    Or sign up with
                                </span>
                            </div>
                            <Button variant="outline" className="w-full">
                                {/* Replace with a proper Google SVG/Icon component */}
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    viewBox="0 0 24 24"
                                    className="w-4 h-4 mr-2"
                                >
                                    <path
                                        d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z"
                                        fill="currentColor"
                                    />
                                </svg>
                                Sign up with Google
                            </Button>
                        </div>
                    </form>
                </CardContent>

                <CardFooter>
                    <div className="text-center text-sm w-full">
                        Already have an account?{" "}
                        {/* CHANGE THIS 'href' to your main login page route */}
                        <a href="/login" className="underline underline-offset-4">
                            Login
                        </a>
                    </div>
                </CardFooter>
            </Card>

            {/* Terms and Policy Footer */}
            <div className="text-muted-foreground *:[a]:hover:text-primary text-center text-xs text-balance *:[a]:underline *:[a]:underline-offset-4">
                By clicking sign up, you agree to our <a href="#">Terms of Service</a>{" "}
                and <a href="#">Privacy Policy</a>.
            </div>
        </div>
    );
}
