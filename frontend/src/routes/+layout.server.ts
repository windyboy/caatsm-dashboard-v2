import type { LayoutServerLoad } from './$types';

// Locale normalization function (consistent with lib/i18n/index.ts)
function normalizeLocale(locale: string | null | undefined): string {
  if (!locale) return 'en';
  const lower = locale.toLowerCase();
  if (lower.startsWith('zh')) return 'zh-CN';
  if (lower.startsWith('en')) return 'en';
  return 'en';
}

export const load: LayoutServerLoad = async ({ cookies }) => {
  // Priority: cookie → fallback
  // Cookie is set by hooks.server.ts based on Accept-Language header or user preference
  const cookieLocale = cookies.get('locale');
  const locale = normalizeLocale(cookieLocale);
  
  return { locale };
};

