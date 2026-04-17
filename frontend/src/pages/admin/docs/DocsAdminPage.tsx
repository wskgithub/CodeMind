import React, { useEffect, useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
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
  Tooltip,
  Empty,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import { documentService, Document } from '@/services/documentService';

const { Title } = Typography;

const DocsAdminPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadDocuments();
  }, []);

  const loadDocuments = async () => {
    try {
      setLoading(true);
      const list = await documentService.listAll();
      setDocuments(list);
    } catch {
      message.error(t('docsAdmin.loadFailed'));
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
      message.success(t('docsAdmin.deleteSuccess'));
      loadDocuments();
    } catch {
      message.error(t('docsAdmin.deleteFailed'));
    }
  };

  const columns = useMemo(() => [
    {
      title: t('docsAdmin.table.sortOrder'),
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 80,
      sorter: (a: Document, b: Document) => a.sort_order - b.sort_order,
    },
    {
      title: t('docsAdmin.table.icon'),
      dataIndex: 'icon',
      key: 'icon',
      width: 60,
      render: (icon: string) => <span style={{ fontSize: 20 }}>{icon || '📄'}</span>,
    },
    {
      title: t('docsAdmin.table.title'),
      dataIndex: 'title',
      key: 'title',
      render: (title: string, record: Document) => (
        <Space direction="vertical" size={0}>
          <span style={{ fontWeight: 500 }}>{title}</span>
          {record.subtitle && (
            <span style={{ fontSize: 12, color: 'var(--ant-color-text-secondary)' }}>{record.subtitle}</span>
          )}
        </Space>
      ),
    },
    {
      title: t('docsAdmin.table.slug'),
      dataIndex: 'slug',
      key: 'slug',
      width: 150,
      render: (slug: string) => <code style={{ fontSize: 12 }}>{slug}</code>,
    },
    {
      title: t('common.status'),
      dataIndex: 'is_published',
      key: 'is_published',
      width: 100,
      render: (isPublished: boolean) =>
        isPublished ? (
          <Tag color="success">{t('docsAdmin.table.published')}</Tag>
        ) : (
          <Tag>{t('docsAdmin.table.unpublished')}</Tag>
        ),
    },
    {
      title: t('docsAdmin.table.updatedAt'),
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
      render: (date: string) => new Date(date).toLocaleString('zh-CN'),
    },
    {
      title: t('common.actions'),
      key: 'action',
      width: 160,
      render: (_: unknown, record: Document) => (
        <Space size="small">
          <Tooltip title={t('docsAdmin.tooltips.view')}>
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleView(record.slug)}
              disabled={!record.is_published}
            />
          </Tooltip>
          <Tooltip title={t('docsAdmin.tooltips.edit')}>
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record.id)}
            />
          </Tooltip>
          <Popconfirm
            title={t('docsAdmin.confirmDelete')}
            description={t('docsAdmin.confirmDeleteDesc')}
            onConfirm={() => handleDelete(record.id)}
            okText={t('common.confirm')}
            cancelText={t('common.cancel')}
          >
            <Tooltip title={t('docsAdmin.tooltips.delete')}>
              <Button type="text" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ], [t]);

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <Title level={4} style={{ margin: 0 }}>
            <FileTextOutlined style={{ marginRight: 8 }} />
            {t('docsAdmin.title')}
          </Title>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            {t('docsAdmin.newDocument')}
          </Button>
        </div>

        <Table
          columns={columns}
          dataSource={documents}
          rowKey="id"
          loading={loading}
          pagination={false}
          locale={{
            emptyText: (
              <Empty
                description={t('docsAdmin.noDocuments')}
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              >
                <Button type="primary" onClick={handleCreate}>
                  {t('docsAdmin.newDocument')}
                </Button>
              </Empty>
            ),
          }}
        />
      </Card>
    </div>
  );
};

export default DocsAdminPage;
