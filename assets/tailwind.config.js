/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "../views/**/*.templ",
    "../views/**/*_templ.go",
    "../internal/handlers/**/*.go",
    "../cmd/**/*.go",
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
  // Safelist to ensure critical classes are included even if not detected
  safelist: [
    // Use pattern matching for better coverage
    {
      pattern: /^(grid|gap|md:grid-cols)/,
    },
    {
      pattern: /^(bg|text|border)-(slate|gray)-(50|100|200|300|400|500|600|700|800|900|950)/,
      variants: ['hover'],
    },
    {
      pattern: /^(bg|text)-(slate|gray)-(900|950)\/(40|60)/,
    },
    {
      pattern: /^(p|px|py|m|mx|my|mt|mb|ml|mr)-(0|1|2|3|4|6|8)/,
    },
    {
      pattern: /^(text)-(xs|sm|base|lg|xl|2xl|3xl)/,
    },
    {
      pattern: /^(font)-(normal|medium|semibold|bold)/,
    },
    {
      pattern: /^(rounded|border|shadow|overflow|h|w)-(xl|lg|md|sm|72|auto)/,
    },
    {
      pattern: /^(space)-(x|y)-(2|4|6)/,
    },
    {
      pattern: /^(uppercase|antialiased|backdrop-blur)/,
    },
  ],
};
