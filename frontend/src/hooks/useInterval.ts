import { useEffect, useRef } from 'react';

/**
 * 定时器 Hook
 * @param callback 回调函数
 * @param delay 延迟时间（毫秒），为 null 时暂停
 */
export function useInterval(callback: () => void, delay: number | null) {
  const savedCallback = useRef<() => void>();

  // 保存最新的回调函数
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  // 设置定时器
  useEffect(() => {
    function tick() {
      savedCallback.current?.();
    }

    if (delay !== null) {
      const id = setInterval(tick, delay);
      return () => clearInterval(id);
    }
  }, [delay]);
}
