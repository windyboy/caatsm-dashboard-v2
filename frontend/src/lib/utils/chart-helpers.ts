import { type ChartOptions, type ChartData } from "chart.js";

export type { ChartOptions, ChartData };

// shadcn灰色系配色方案
export const chartColors = {
  zinc: {
    50: "#fafafa",
    100: "#f4f4f5",
    200: "#e4e4e7",
    300: "#d4d4d8",
    400: "#a1a1aa",
    500: "#71717a",
    600: "#52525b",
    700: "#3f3f46",
    800: "#27272a",
    900: "#18181b",
  },
  accent: {
    blue: "#2563eb",
    green: "#16a34a",
    yellow: "#eab308",
    amber: "#f59e0b",
    red: "#dc2626",
    purple: "#9333ea",
  },
};

// 默认图表配置（shadcn风格）
export const defaultChartOptions: Partial<ChartOptions<"line" | "bar" | "doughnut">> = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: "top" as const,
      labels: {
        usePointStyle: true,
        padding: 12,
        font: {
          size: 12,
        },
        color: chartColors.zinc[900],
      },
    },
    tooltip: {
      backgroundColor: chartColors.zinc[800],
      titleColor: chartColors.zinc[100],
      bodyColor: chartColors.zinc[200],
      borderColor: chartColors.zinc[700],
      borderWidth: 1,
      padding: 12,
      cornerRadius: 6,
      displayColors: true,
    },
  },
  scales: {
    x: {
      grid: {
        color: chartColors.zinc[200],
      },
      ticks: {
        color: chartColors.zinc[600],
        font: {
          size: 11,
        },
      },
    },
    y: {
      grid: {
        color: chartColors.zinc[200],
      },
      ticks: {
        color: chartColors.zinc[600],
        font: {
          size: 11,
        },
      },
    },
  },
};

// 生成饼图/环形图颜色数组
export function generateColors(count: number): string[] {
  const colors = [
    chartColors.accent.blue,
    chartColors.accent.green,
    chartColors.accent.purple,
    chartColors.accent.amber,
    chartColors.zinc[500],
    chartColors.zinc[400],
    chartColors.accent.yellow,
    chartColors.zinc[600],
  ];

  // 如果需要的颜色超过预设，循环使用
  return Array.from({ length: count }, (_, i) => colors[i % colors.length]);
}

// 格式化数字（用于图表显示）
export function formatChartNumber(value: number): string {
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(1)}M`;
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}K`;
  }
  return value.toString();
}
