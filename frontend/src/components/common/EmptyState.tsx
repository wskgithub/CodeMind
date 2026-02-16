import { Empty, Button } from 'antd';

interface EmptyStateProps {
  /** 描述文本 */
  description?: string;
  /** 操作按钮文本 */
  actionText?: string;
  /** 操作回调 */
  onAction?: () => void;
}

/** 通用空状态组件 — 统一的空数据占位展示 */
const EmptyState: React.FC<EmptyStateProps> = ({
  description = '暂无数据',
  actionText,
  onAction,
}) => {
  return (
    <div style={{ padding: '60px 0', textAlign: 'center' }}>
      <Empty description={description}>
        {actionText && onAction && (
          <Button type="primary" onClick={onAction}>
            {actionText}
          </Button>
        )}
      </Empty>
    </div>
  );
};

export default EmptyState;
