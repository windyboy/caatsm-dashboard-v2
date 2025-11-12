/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../views/**/*.{templ,go}",
    "../internal/handlers/**/*.go",
    "./css/**/*.css",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui"],
      },
    },
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};

