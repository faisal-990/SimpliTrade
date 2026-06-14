/// <reference types="vite/client" />

// The vendored shadcn UI components are JavaScript (.jsx); declare the modules
// so our TypeScript code can import them without type errors. They render fine
// at runtime via Vite; we just don't get prop types for them.
declare module "@/components/ui/*";
declare module "@/components/selfMade/*";
