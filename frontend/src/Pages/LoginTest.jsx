import { useState } from "react";

function LoginTest() {
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState({});
  
  const token = "eyJhbGciOiJIUzI2NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImhlbGxvIGZyb20gIiwiZXhwIjoxNzU0NzY2Mjk5LCJpYXQiOjE3NTQ3NjU2OTl9.sMtK7B7bsMYHHpYJD_yJkCE9tdiH1o0XfB38bfKxfX8";

  // Safe JSON parsing function
  const safeJSONParse = (text) => {
    try {
      // First, try to parse as-is
      return JSON.parse(text);
    } catch (error) {
      console.warn("Initial JSON parse failed:", error.message);
      console.log("Raw response text:", text);
      
      try {
        // Try to extract the first valid JSON object
        const firstBraceIndex = text.indexOf('{');
        const lastBraceIndex = text.lastIndexOf('}');
        
        if (firstBraceIndex !== -1 && lastBraceIndex !== -1 && lastBraceIndex > firstBraceIndex) {
          const jsonSubstring = text.substring(firstBraceIndex, lastBraceIndex + 1);
          return JSON.parse(jsonSubstring);
        }
        
        // If that fails, try to handle multiple concatenated JSON objects
        const jsonObjects = [];
        let currentIndex = 0;
        
        while (currentIndex < text.length) {
          const openBrace = text.indexOf('{', currentIndex);
          if (openBrace === -1) break;
          
          let braceCount = 0;
          let endIndex = openBrace;
          
          for (let i = openBrace; i < text.length; i++) {
            if (text[i] === '{') braceCount++;
            if (text[i] === '}') braceCount--;
            if (braceCount === 0) {
              endIndex = i;
              break;
            }
          }
          
          if (braceCount === 0) {
            const jsonCandidate = text.substring(openBrace, endIndex + 1);
            try {
              jsonObjects.push(JSON.parse(jsonCandidate));
            } catch (e) {
              console.warn("Failed to parse JSON segment:", jsonCandidate);
            }
            currentIndex = endIndex + 1;
          } else {
            break;
          }
        }
        
        // Return the first valid JSON object or the array if multiple
        return jsonObjects.length === 1 ? jsonObjects[0] : jsonObjects;
        
      } catch (fallbackError) {
        // Last resort: return an error object with the raw text
        console.error("All JSON parsing attempts failed:", fallbackError);
        return { 
          error: true, 
          message: "Invalid JSON response from server",
          rawText: text.substring(0, 200) + (text.length > 200 ? '...' : '')
        };
      }
    }
  };

  // Input validation functions (keeping the same as before)
  const validateName = (name) => {
    const trimmedName = name?.trim() || "";
    if (!trimmedName) return "Name is required";
    if (trimmedName.length < 2) return "Name must be at least 2 characters long";
    if (trimmedName.length > 50) return "Name must be less than 50 characters";
    const nameRegex = /^[a-zA-Z0-9\s\-'.]+$/;
    if (!nameRegex.test(trimmedName)) return "Name contains invalid characters";
    if (/\s{2,}/.test(trimmedName)) return "Name cannot contain multiple consecutive spaces";
    return null;
  };

  const validatePassword = (password) => {
    if (!password) return "Password is required";
    if (password.length < 8) return "Password must be at least 8 characters long";
    if (password.length > 128) return "Password is too long";
    return null;
  };

  const handleNameChange = (e) => {
    const value = e.target.value;
    setName(value);
    if (errors.name) {
      setErrors(prev => ({ ...prev, name: null }));
    }
  };

  const handlePasswordChange = (e) => {
    const value = e.target.value;
    setPassword(value);
    if (errors.password) {
      setErrors(prev => ({ ...prev, password: null }));
    }
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    
    setErrors({});
    
    const nameError = validateName(name);
    const passwordError = validatePassword(password);
    
    const newErrors = {};
    if (nameError) newErrors.name = nameError;
    if (passwordError) newErrors.password = passwordError;
    
    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }
    
    setIsLoading(true);

    try {
      const cleanName = name.trim().replace(/\s+/g, ' ');
      
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 30000);
      
      const res = await fetch("http://localhost:8080/api/auth/signup", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`
        },
        body: JSON.stringify({
          name: cleanName,
          password,
        }),
        signal: controller.signal
      });

      clearTimeout(timeoutId);
      
      // Get the raw response text first
      const rawText = await res.text();
      console.log("Raw response:", rawText);
      
      // Use safe JSON parsing
      const data = safeJSONParse(rawText);
      console.log("Parsed response:", data);

      // Handle error responses from safe parser
      if (data.error) {
        alert(`Server response error: ${data.message}`);
        return;
      }

      if (!res.ok) {
        let errorMessage = "Login failed";
        
        if (res.status === 400) {
          errorMessage = data.message || "Invalid input provided";
        } else if (res.status === 401) {
          errorMessage = "Invalid credentials";
        } else if (res.status === 403) {
          errorMessage = "Access denied";
        } else if (res.status === 429) {
          errorMessage = "Too many login attempts. Please try again later";
        } else if (res.status >= 500) {
          errorMessage = "Server error. Please try again later";
        } else {
          errorMessage = data.message || `Error ${res.status}`;
        }
        
        alert(errorMessage);
        return;
      }

      alert("Login successful!");
      setName("");
      setPassword("");
      
    } catch (err) {
      console.error("Error logging in:", err);
      
      let errorMessage = "Something went wrong!";
      
      if (err.name === 'AbortError') {
        errorMessage = "Request timed out. Please try again.";
      } else if (err.name === 'TypeError' && err.message.includes('fetch')) {
        errorMessage = "Network error. Please check your connection.";
      }
      
      alert(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  // Rest of the component remains the same...
  return (
    <div className="h-screen w-screen flex items-center justify-center bg-gray-100">
      <form
        onSubmit={handleLogin}
        className="flex flex-col gap-4 p-6 bg-white rounded-2xl shadow-lg w-80"
        noValidate
      >
        <h2 className="text-xl font-semibold text-center">Login Test</h2>

        <div className="flex flex-col">
          <input
            type="text"
            placeholder="Name"
            value={name}
            onChange={handleNameChange}
            className={`border p-2 rounded-lg focus:outline-none focus:ring-2 transition ${
              errors.name 
                ? 'border-red-500 focus:ring-red-500' 
                : 'border-gray-300 focus:ring-blue-500'
            }`}
            disabled={isLoading}
            maxLength={50}
            autoComplete="username"
          />
          {errors.name && (
            <span className="text-red-500 text-sm mt-1">{errors.name}</span>
          )}
        </div>

        <div className="flex flex-col">
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={handlePasswordChange}
            className={`border p-2 rounded-lg focus:outline-none focus:ring-2 transition ${
              errors.password 
                ? 'border-red-500 focus:ring-red-500' 
                : 'border-gray-300 focus:ring-blue-500'
            }`}
            disabled={isLoading}
            maxLength={128}
            autoComplete="current-password"
          />
          {errors.password && (
            <span className="text-red-500 text-sm mt-1">{errors.password}</span>
          )}
        </div>

        <button
          type="submit"
          disabled={isLoading}
          className={`p-2 rounded-lg text-white font-medium transition ${
            isLoading 
              ? 'bg-gray-400 cursor-not-allowed' 
              : 'bg-blue-500 hover:bg-blue-600 active:bg-blue-700'
          }`}
        >
          {isLoading ? 'Logging in...' : 'Login'}
        </button>
      </form>
    </div>
  );
}

export default LoginTest;

