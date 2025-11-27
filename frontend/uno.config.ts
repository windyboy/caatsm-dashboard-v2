import { defineConfig } from "unocss";
import presetWind4 from "@unocss/preset-wind4";
import presetTypography from "@unocss/preset-typography";
import transformerDirectives from "@unocss/transformer-directives";

export default defineConfig({
  // 使用 Wind4 preset（提供 Tailwind4 兼容的工具类）
  presets: [
    presetWind4({
      preflights: {
        reset: true, // 启用内置的 reset 样式
      },
    }),
    presetTypography(),
  ],

  // 转换器
  transformers: [
    transformerDirectives(), // 支持 @apply 指令
  ],

  // 内容扫描
  content: {
    filesystem: ["src/**/*.{html,js,svelte,ts}"],
  },

  // 主题配置
  theme: {
    // PresetWind4 使用 font 而不是 fontFamily
    // @ts-ignore - PresetWind4 使用 font 而不是 fontFamily，但类型定义可能还未更新
    font: {
      family: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
    },

    // Design System Tokens - CSS Variables for consistent theming
    spacing: {
      'xs': 'var(--spacing-xs, 0.25rem)',    // 4px
      'sm': 'var(--spacing-sm, 0.5rem)',     // 8px
      'md': 'var(--spacing-md, 0.75rem)',    // 12px
      'lg': 'var(--spacing-lg, 1rem)',       // 16px
      'xl': 'var(--spacing-xl, 1.25rem)',    // 20px
      '2xl': 'var(--spacing-2xl, 1.5rem)',   // 24px
      '3xl': 'var(--spacing-3xl, 2rem)',     // 32px
      '4xl': 'var(--spacing-4xl, 2.5rem)',   // 40px
      '5xl': 'var(--spacing-5xl, 3rem)',     // 48px
      '6xl': 'var(--spacing-6xl, 4rem)',     // 64px
    },

    borderRadius: {
      'none': 'var(--radius-none, 0)',
      'sm': 'var(--radius-sm, 0.125rem)',    // 2px
      'md': 'var(--radius-md, 0.375rem)',    // 6px
      'lg': 'var(--radius-lg, 0.5rem)',      // 8px
      'xl': 'var(--radius-xl, 0.75rem)',     // 12px
      '2xl': 'var(--radius-2xl, 1rem)',      // 16px
      '3xl': 'var(--radius-3xl, 1.5rem)',    // 24px
      'full': 'var(--radius-full, 9999px)',
    },

    fontSize: {
      'xs': ['var(--text-xs, 0.75rem)', 'var(--leading-xs, 1rem)'],       // 12px
      'sm': ['var(--text-sm, 0.875rem)', 'var(--leading-sm, 1.25rem)'],    // 14px
      'base': ['var(--text-base, 1rem)', 'var(--leading-base, 1.5rem)'],   // 16px
      'lg': ['var(--text-lg, 1.125rem)', 'var(--leading-lg, 1.75rem)'],    // 18px
      'xl': ['var(--text-xl, 1.25rem)', 'var(--leading-xl, 1.75rem)'],     // 20px
      '2xl': ['var(--text-2xl, 1.5rem)', 'var(--leading-2xl, 2rem)'],      // 24px
      '3xl': ['var(--text-3xl, 1.875rem)', 'var(--leading-3xl, 2.25rem)'], // 30px
      '4xl': ['var(--text-4xl, 2.25rem)', 'var(--leading-4xl, 2.5rem)'],   // 36px
      '5xl': ['var(--text-5xl, 3rem)', 'var(--leading-5xl, 1)'],           // 48px
    },

    fontWeight: {
      'thin': 'var(--font-thin, 100)',
      'extralight': 'var(--font-extralight, 200)',
      'light': 'var(--font-light, 300)',
      'normal': 'var(--font-normal, 400)',
      'medium': 'var(--font-medium, 500)',
      'semibold': 'var(--font-semibold, 600)',
      'bold': 'var(--font-bold, 700)',
      'extrabold': 'var(--font-extrabold, 800)',
      'black': 'var(--font-black, 900)',
    },

    boxShadow: {
      'xs': 'var(--shadow-xs, 0 1px 2px 0 rgb(0 0 0 / 0.05))',
      'sm': 'var(--shadow-sm, 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1))',
      'md': 'var(--shadow-md, 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1))',
      'lg': 'var(--shadow-lg, 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1))',
      'xl': 'var(--shadow-xl, 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1))',
      '2xl': 'var(--shadow-2xl, 0 25px 50px -12px rgb(0 0 0 / 0.25))',
      'inner': 'var(--shadow-inner, inset 0 2px 4px 0 rgb(0 0 0 / 0.05))',
      'none': 'var(--shadow-none, 0 0 #0000)',
    },
    // 扩展颜色系统 - 添加更鲜艳的配色方案
    colors: {
      // 品牌色 - 更鲜艳的蓝色系
      brand: {
        50: '#eff6ff',
        100: '#dbeafe',
        200: '#bfdbfe',
        300: '#93c5fd',
        400: '#60a5fa',
        500: '#3b82f6',
        600: '#2563eb',
        700: '#1d4ed8',
        800: '#1e40af',
        900: '#1e3a8a',
        950: '#172554',
      },
      // 强调色 - 紫色系
      accent: {
        50: '#faf5ff',
        100: '#f3e8ff',
        200: '#e9d5ff',
        300: '#d8b4fe',
        400: '#c084fc',
        500: '#a855f7',
        600: '#9333ea',
        700: '#7c3aed',
        800: '#6b21a8',
        900: '#581c87',
        950: '#3b0764',
      },
      // 成功色 - 更鲜艳的绿色
      success: {
        50: '#f0fdf4',
        100: '#dcfce7',
        200: '#bbf7d0',
        300: '#86efac',
        400: '#4ade80',
        500: '#22c55e',
        600: '#16a34a',
        700: '#15803d',
        800: '#166534',
        900: '#14532d',
        950: '#052e16',
      },
      // 警告色 - 橙色系
      warning: {
        50: '#fff7ed',
        100: '#ffedd5',
        200: '#fed7aa',
        300: '#fdba74',
        400: '#fb923c',
        500: '#f97316',
        600: '#ea580c',
        700: '#c2410c',
        800: '#9a3412',
        900: '#7c2d12',
        950: '#431407',
      },
      // 危险色 - 红色系
      danger: {
        50: '#fef2f2',
        100: '#fee2e2',
        200: '#fecaca',
        300: '#fca5a5',
        400: '#f87171',
        500: '#ef4444',
        600: '#dc2626',
        700: '#b91c1c',
        800: '#991b1b',
        900: '#7f1d1d',
        950: '#450a0a',
      },
    },
    // 动画配置
    animation: {
      'fade-in': 'fadeIn 0.5s ease-in-out',
      'slide-in-up': 'slideInUp 0.3s ease-out',
      'slide-in-down': 'slideInDown 0.3s ease-out',
      'scale-in': 'scaleIn 0.2s ease-out',
      'bounce-in': 'bounceIn 0.6s ease-out',
      'glow': 'glow 2s ease-in-out infinite alternate',
      'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
    },
    keyframes: {
      fadeIn: {
        '0%': { opacity: '0' },
        '100%': { opacity: '1' },
      },
      slideInUp: {
        '0%': { transform: 'translateY(20px)', opacity: '0' },
        '100%': { transform: 'translateY(0)', opacity: '1' },
      },
      slideInDown: {
        '0%': { transform: 'translateY(-20px)', opacity: '0' },
        '100%': { transform: 'translateY(0)', opacity: '1' },
      },
      scaleIn: {
        '0%': { transform: 'scale(0.9)', opacity: '0' },
        '100%': { transform: 'scale(1)', opacity: '1' },
      },
      bounceIn: {
        '0%': { transform: 'scale(0.3)', opacity: '0' },
        '50%': { transform: 'scale(1.05)' },
        '70%': { transform: 'scale(0.9)' },
        '100%': { transform: 'scale(1)', opacity: '1' },
      },
      glow: {
        '0%': { boxShadow: '0 0 20px rgba(59, 130, 246, 0.5)' },
        '100%': { boxShadow: '0 0 30px rgba(59, 130, 246, 0.8)' },
      },
    },
  },
});
