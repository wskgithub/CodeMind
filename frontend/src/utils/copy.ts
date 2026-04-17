/**
 * 复制文本到剪贴板（兼容非 HTTPS / localhost 环境）
 * 生产环境若不是安全上下文，navigator.clipboard 可能不可用，此时降级到 document.execCommand('copy')
 */
export const copyToClipboard = async (text: string): Promise<boolean> => {
  // prefer modern Clipboard API
  if (navigator.clipboard) {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      // fall through to legacy method
    }
  }

  // fallback: hidden textarea + execCommand
  const textArea = document.createElement('textarea');
  textArea.value = text;
  textArea.style.position = 'fixed';
  textArea.style.left = '-9999px';
  textArea.style.top = '-9999px';
  textArea.setAttribute('readonly', '');
  document.body.appendChild(textArea);
  textArea.select();
  textArea.setSelectionRange(0, text.length);

  try {
    const result = document.execCommand('copy');
    document.body.removeChild(textArea);
    return result;
  } catch {
    document.body.removeChild(textArea);
    return false;
  }
};
