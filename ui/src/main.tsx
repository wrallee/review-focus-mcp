import React from "react";
import { createRoot } from "react-dom/client";
import ReviewFocusApp from "./App";
import { connectApp } from "./mcp";

async function main() {
  await connectApp();
  createRoot(document.getElementById("root")!).render(<React.StrictMode><ReviewFocusApp /></React.StrictMode>);
}

void main();
