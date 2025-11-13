import { defineConfig } from 'unocss'
import presetWind3 from '@unocss/preset-wind3'
import presetTypography from '@unocss/preset-typography'
import transformerDirectives from '@unocss/transformer-directives'

export default defineConfig({
  // 使用 Wind preset（提供 Tailwind 兼容的工具类）
  presets: [
    presetWind3(),
    presetTypography(),
  ],
  
  // 转换器
  transformers: [
    transformerDirectives(), // 支持 @apply 指令
  ],
  
  // 主题配置
  theme: {
    // @ts-ignore - theme 配置支持 fontFamily
    fontFamily: {
      sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
    },
  },
  
  // Safelist - 确保动态生成的类被包含
  // @ts-ignore - UnoCSS safelist 支持正则表达式，但 TypeScript 类型定义可能不完整
  safelist: [
    // 实际使用的工具类
    'scroll-smooth', 'leading-relaxed', 'backdrop-blur-sm', 'backdrop-blur-lg',
    'bg-clip-text', 'text-transparent', 'shadow-xl',
    // 模式匹配 - 支持所有颜色值和透明度（覆盖动态生成的颜色类）
    /^(bg|text|border)-(sky|rose|emerald|violet|amber|slate|purple|green|blue|red|gray)-(50|100|200|300|400|500|600|700|800|900|950)\/?(10|20|30|40|50|60|70|80|90|95)?$/,
    // 间距和尺寸
    /^(p|px|py|pt|pb|pl|pr|m|mx|my|mt|mb|ml|mr)-(0|0\.5|1|1\.5|2|2\.5|3|4|6|8|12)$/,
    /^(gap|space-x|space-y)-(0|0\.5|1|1\.5|2|2\.5|3|4|6)$/,
    /^space-y-2\.5$/,
    // 文字和字体
    /^(text)-(xs|sm|base|lg|xl|2xl|3xl)$/,
    /^(font)-(normal|medium|semibold|bold)$/,
    // 圆角、边框、阴影、溢出
    /^(rounded|border|shadow|overflow|h|w)-(xl|lg|md|sm|full|auto)$/,
    /^shadow-(green|purple|slate|emerald|sky|rose|violet|amber)-(500|900)\/(20|50)$/,
    /^overflow-(x|y)-(auto|scroll|hidden|visible)$/,
    // 任意值语法
    /^(h|w|max-h|min-h)-\[.+\]$/,
    // 布局
    /^(grid|gap|md:grid-cols|lg:grid-cols|lg:col-span)/,
    // 渐变
    /^bg-gradient-to-(r|l|t|b|tr|tl|br|bl)$/,
    /^from-(sky|purple|blue|emerald|slate|indigo)-(400|500|600|50|100|200)\/?(20|80|90|95)?$/,
    /^via-(blue|sky|purple|emerald|slate|indigo)-(400|500|600)$/,
    /^to-(blue|purple|emerald|slate|indigo|white)-(500|600|50|100|200|90)\/?(20|80|90|95)?$/,
    // 动画和过渡
    /^animate-(pulse|spin|bounce)$/,
    /^transition-(all|colors|opacity|transform)$/,
    /^duration-(75|100|150|200|300|500|700|1000)$/,
    // 悬停和焦点状态
    /^hover:(bg|text|border|shadow)-(sky|rose|emerald|violet|amber|slate|purple|green|blue|red|gray)-(50|100|200|300|400|500|600|700|800|900|950)\/?(10|20|30|40|50|60|70|80|90|95)?$/,
    /^group-hover:(bg|text|border)-(sky|rose|emerald|violet|amber|slate|purple|green|blue|red|gray)-(50|100|200|300|400|500|600|700|800|900|950)\/?(10|20|30|40|50|60|70|80|90|95)?$/,
    /^focus:(border|outline|ring)-(sky|slate|gray|emerald|purple)-(50|100|200|300|400|500|600|700|800|900|950)\/?(10|20|30|40|50|60|70|80|90|95)?$/,
    // placeholder 颜色
    /^placeholder:(text)-(slate|gray)-(400|500|600)$/,
  ] as any,
})

