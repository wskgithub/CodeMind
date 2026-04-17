/**
 * Copy text to clipboard (compatible with non-HTTPS / localhost environments).
 * Falls back to document.execCommand('copy') when navigator.clipboard is unavailable
 * outside a secure context in production.
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
