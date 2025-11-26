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
      sans: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
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
