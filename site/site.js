const languageToggle = document.querySelector('#language-toggle');
const translatedElements = [...document.querySelectorAll('[data-i18n]')];
const englishContent = new Map(translatedElements.map(element => [element, element.innerHTML]));
let language = 'en';

function setLanguage(value) {
  language = value === 'zh' ? 'zh' : 'en';
  document.documentElement.lang = language === 'zh' ? 'zh-CN' : 'en';
  for (const element of translatedElements) {
    element.innerHTML = language === 'zh' ? element.dataset.zh : englishContent.get(element);
  }
  languageToggle.textContent = language === 'zh' ? 'English' : '中文';
  languageToggle.setAttribute('aria-label', language === 'zh' ? 'Switch to English' : '切换为中文');
  for (const button of document.querySelectorAll('[data-copy]')) {
    button.textContent = language === 'zh' ? '复制' : 'Copy';
    button.setAttribute('aria-label', language === 'zh' ? '复制安装命令' : 'Copy install command');
  }
  document.querySelector('meta[name="description"]').content = language === 'zh'
    ? 'jlink-cli 是基于 Go 的命令行工具，为 AI Agent 提供 SEGGER J-Link 工具支持。'
    : 'jlink-cli is a Go-first command-line interface for AI agents working with SEGGER J-Link tooling.';
  try { localStorage.setItem('jlink-cli-language', language); } catch { /* Storage may be disabled. */ }
}

let initialLanguage = navigator.language.startsWith('zh') ? 'zh' : 'en';
try { initialLanguage = localStorage.getItem('jlink-cli-language') || initialLanguage; } catch { /* Use browser language. */ }
setLanguage(initialLanguage);
languageToggle.addEventListener('click', () => setLanguage(language === 'zh' ? 'en' : 'zh'));

for (const button of document.querySelectorAll('[data-copy]')) {
  button.addEventListener('click', async () => {
    const command = button.getAttribute('data-copy');
    if (!command) return;
    try {
      await navigator.clipboard.writeText(command);
      button.textContent = language === 'zh' ? '已复制' : 'Copied';
    } catch {
      button.textContent = language === 'zh' ? '请手动复制' : 'Copy manually';
    }
    window.setTimeout(() => { button.textContent = language === 'zh' ? '复制' : 'Copy'; }, 1800);
  });
}
