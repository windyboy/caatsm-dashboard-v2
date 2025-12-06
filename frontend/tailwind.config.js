/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: ["./src/**/*.{html,js,svelte,ts}"],
  // TailwindCSS v4 uses CSS-first configuration via @theme in app.css
  // Theme colors and other design tokens are defined in app.css using @theme inline
  // This config file is kept minimal for compatibility
};
