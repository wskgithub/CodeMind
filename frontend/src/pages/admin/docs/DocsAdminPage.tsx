import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Table,
  Button,
  Space,
  Tag,
  Popconfirm,
  message,
  Card,
  Typography,
  Alert,
  Tooltip,
  Modal,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  FileTextOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { documentService, Document } from '@/services/documentService';


const { Title } = Typography;

const DocsAdminPage: React.FC = () => {
  const navigate = useNavigate();

  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(false);
  const [initModalVisible, setInitModalVisible] = useState(false);
  const [initLoading, setInitLoading] = useState(false);

  useEffect(() => {
    loadDocuments();
  }, []);

  const loadDocuments = async () => {
    try {
      setLoading(true);
      const list = await documentService.listAll();
      setDocuments(list);
    } catch (error) {
      message.error('加载文档列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    navigate('/admin/docs/create');
  };

  const handleEdit = (id: number) => {
    navigate(`/admin/docs/edit/${id}`);
  };

  const handleView = (slug: string) => {
    navigate(`/docs/${slug}`);
  };

  const handleDelete = async (id: number) => {
    try {
      await documentService.delete(id);
      message.success('删除成功');
      loadDocuments();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleInitialize = async () => {
    try {
      setInitLoading(true);
      const result = await documentService.initialize();
      message.success(`初始化成功，共创建 ${result.count} 篇文档`);
      setInitModalVisible(false);
      loadDocuments();
    } catch (error: any) {
      if (error.response?.data?.error === '文档已存在，无法初始化') {
        message.warning('文档已存在，无法初始化');
      } else {
        message.error('初始化失败');
      }
    } finally {
      setInitLoading(false);
    }
  };

  const columns = [
    {
      title: '排序',
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 80,
      sorter: (a: Document, b: Document) => a.sort_order - b.sort_order,
    },
    {
      title: '标识',
      dataIndex: 'slug',
      key: 'slug',
      width: 150,
    },
    {
      title: '图标',
      dataIndex: 'icon',
      key: 'icon',
      width: 80,
      render: (icon: string) => <span style={{ fontSize: 20 }}>{icon || '📄'}</span>,
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (title: string, record: Document) => (
        <Space direction="vertical" size={0}>
          <span style={{ fontWeight: 500 }}>{title}</span>
          <span style={{ fontSize: 12, color: '#8c8c8c' }}>{record.subtitle}</span>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'is_published',
      key: 'is_published',
      width: 100,
      render: (isPublished: boolean) =>
        isPublished ? (
          <Tag color="success">已发布</Tag>
        ) : (
          <Tag color="default">未发布</Tag>
        ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: any, record: Document) => (
        <Space size="small">
          <Tooltip title="查看">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleView(record.slug)}
              disabled={!record.is_published}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record.id)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除这篇文档吗？"
            description="删除后无法恢复"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button type="text" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="docs-admin-page">
      <Card>
        <div className="docs-admin-header">
          <Title level={4} className="docs-admin-title">
            <FileTextOutlined style={{ marginRight: 8 }} />
            文档管理
          </Title>
          <Space>
            {documents.length === 0 && (
              <Button onClick={() => setInitModalVisible(true)}>
                初始化默认文档
              </Button>
            )}
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              新建文档
            </Button>
          </Space>
        </div>

        {documents.length === 0 && !loading && (
          <Alert
            message="暂无文档"
            description="您可以点击「初始化默认文档」按钮快速创建开发工具接入文档，或手动创建新文档。"
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <Table
          columns={columns}
          dataSource={documents}
          rowKey="id"
          loading={loading}
          pagination={false}
        />
      </Card>

      <Modal
        title={
          <Space>
            <ExclamationCircleOutlined style={{ color: '#faad14' }} />
            确认初始化
          </Space>
        }
        open={initModalVisible}
        onOk={handleInitialize}
        onCancel={() => setInitModalVisible(false)}
        confirmLoading={initLoading}
        okText="确认初始化"
        cancelText="取消"
      >
        <p>初始化将创建以下默认文档：</p>
        <ul>
          <li>Claude Code</li>
          <li>Claude Code IDE 插件</li>
          <li>Cursor</li>
          <li>TRAE</li>
          <li>Cline</li>
          <li>Kilo Code</li>
          <li>Roo Code</li>
          <li>OpenCode</li>
          <li>Factory Droid</li>
          <li>Crush</li>
          <li>Goose</li>
          <li>OpenClaw</li>
          <li>Cherry Studio</li>
          <li>其他工具</li>
        </ul>
        <p style={{ color: '#ff4d4f' }}>注意：此操作仅在文档表为空时可用。</p>
      </Modal>
    </div>
  );
};

export default DocsAdminPage;
