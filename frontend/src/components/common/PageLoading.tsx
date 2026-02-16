import { Spin } from 'antd';

/** 页面级加载状态组件 — 全屏居中 Spinner */
const PageLoading: React.FC = () => {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: 400,
        width: '100%',
      }}
    >
      <Spin size="large" tip="加载中..." />
    </div>
  );
};

export default PageLoading;
