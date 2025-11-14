import { defineConfig } from 'unocss'
import presetWind4 from '@unocss/preset-wind4'
import presetTypography from '@unocss/preset-typography'
import transformerDirectives from '@unocss/transformer-directives'

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
  
  // 主题配置
  theme: {
    // PresetWind4 使用 font 而不是 fontFamily
    // @ts-ignore - PresetWind4 使用 font 而不是 fontFamily，但类型定义可能还未更新
    font: {
      sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
    },
  },
  
  // Safelist 暂时移除，先测试基本配置是否能正常工作
  // 如果布局正常，再根据需要添加必要的 safelist
})

