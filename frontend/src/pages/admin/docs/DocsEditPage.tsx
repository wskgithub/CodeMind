import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Form,
  Input,
  Button,
  Card,
  Typography,
  message,
  Space,
  Switch,
  InputNumber,
  Skeleton,
  Tag,
  Tooltip,
} from 'antd';
import {
  SaveOutlined,
  ArrowLeftOutlined,
  SendOutlined,
  CloudOutlined,
  LoadingOutlined,
  PictureOutlined,
} from '@ant-design/icons';
import MDEditor from '@uiw/react-md-editor';
import rehypeSanitize from 'rehype-sanitize';
import { documentService, CreateDocumentRequest, UpdateDocumentRequest } from '@/services/documentService';
import { uploadService } from '@/services/uploadService';
import useAppStore from '@/store/appStore';

const { Title } = Typography;

// 自动保存间隔（毫秒）
const AUTO_SAVE_INTERVAL = 30000;

const DocsEditPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = !!id;
  const [form] = Form.useForm();
  const themeMode = useAppStore((s) => s.themeMode);

  const [loading, setLoading] = useState(isEdit);
  const [saving, setSaving] = useState(false);
  const [content, setContent] = useState('');
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [lastSavedAt, setLastSavedAt] = useState<Date | null>(null);
  const [autoSaveStatus, setAutoSaveStatus] = useState<'idle' | 'saving' | 'saved'>('idle');
  const [isUploading, setIsUploading] = useState(false);
  const [isDragging, setIsDragging] = useState(false);

  const autoSaveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const contentRef = useRef(content);
  const isMountedRef = useRef(true);
  const imageInputRef = useRef<HTMLInputElement>(null);
  const editorRef = useRef<HTMLDivElement>(null);

  // 保持 contentRef 同步
  useEffect(() => {
    contentRef.current = content;
  }, [content]);

  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    if (isEdit) {
      loadDocument();
    }
  }, [id]);

  const loadDocument = async () => {
    try {
      setLoading(true);
      const doc = await documentService.getById(Number(id));
      if (doc) {
        form.setFieldsValue({
          slug: doc.slug,
          title: doc.title,
          subtitle: doc.subtitle,
          icon: doc.icon,
          sort_order: doc.sort_order,
          is_published: doc.is_published,
        });
        setContent(doc.content);
        setLastSavedAt(new Date(doc.updated_at));
      } else {
        message.error(t('docsAdmin.editPage.docNotFound'));
        navigate('/admin/docs');
      }
    } catch {
      message.error(t('docsAdmin.editPage.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  // 保存逻辑（同时支持手动保存和自动保存）
  const doSave = useCallback(async (isAutoSave = false) => {
    if (saving) return;

    try {
      const values = form.getFieldsValue();
      if (!values.title) {
        if (!isAutoSave) message.warning(t('docsAdmin.editPage.titleRequired'));
        return;
      }

      setSaving(true);
      if (isAutoSave) setAutoSaveStatus('saving');

      const currentContent = contentRef.current;

      if (isEdit) {
        const data: UpdateDocumentRequest = {
          title: values.title,
          subtitle: values.subtitle || '',
          icon: values.icon || '',
          content: currentContent,
          sort_order: values.sort_order || 0,
          is_published: values.is_published ?? false,
        };
        await documentService.update(Number(id), data);
        if (!isAutoSave) {
          message.success(t('docsAdmin.editPage.updateSuccess'));
        }
      } else {
        if (!values.slug) {
          if (!isAutoSave) message.warning(t('docsAdmin.editPage.slugRequired'));
          return;
        }
        const data: CreateDocumentRequest = {
          slug: values.slug,
          title: values.title,
          subtitle: values.subtitle || '',
          icon: values.icon || '',
          content: currentContent,
          sort_order: values.sort_order || 0,
          is_published: values.is_published ?? false,
        };
        const doc = await documentService.create(data);
        if (!isAutoSave) {
          message.success(t('docsAdmin.editPage.createSuccess'));
          navigate(`/admin/docs/edit/${doc.id}`, { replace: true });
        }
      }

      if (isMountedRef.current) {
        setHasUnsavedChanges(false);
        setLastSavedAt(new Date());
        if (isAutoSave) {
          setAutoSaveStatus('saved');
          setTimeout(() => {
            if (isMountedRef.current) setAutoSaveStatus('idle');
          }, 2000);
        }
      }
    } catch (error: any) {
      if (!isAutoSave) {
        const errorMsg = error.response?.data?.error;
        message.error(errorMsg || (isEdit ? t('docsAdmin.editPage.updateFailed') : t('docsAdmin.editPage.createFailed')));
      }
      if (isMountedRef.current) setAutoSaveStatus('idle');
    } finally {
      if (isMountedRef.current) setSaving(false);
    }
  }, [saving, form, isEdit, id, navigate, t]);

  // 手动保存（快捷键或按钮）
  const handleSave = useCallback(() => doSave(false), [doSave]);

  // 保存并发布
  const handlePublish = useCallback(async () => {
    form.setFieldValue('is_published', true);
    await doSave(false);
  }, [doSave, form]);

  // Ctrl+S 快捷键
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        handleSave();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleSave]);

  // 自动保存：编辑模式下，内容有变化时定时保存
  useEffect(() => {
    if (!isEdit || !hasUnsavedChanges) return;

    if (autoSaveTimerRef.current) {
      clearTimeout(autoSaveTimerRef.current);
    }

    autoSaveTimerRef.current = setTimeout(() => {
      doSave(true);
    }, AUTO_SAVE_INTERVAL);

    return () => {
      if (autoSaveTimerRef.current) {
        clearTimeout(autoSaveTimerRef.current);
      }
    };
  }, [isEdit, hasUnsavedChanges, content, doSave]);

  // 离开页面前提示未保存
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasUnsavedChanges) {
        e.preventDefault();
      }
    };
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [hasUnsavedChanges]);

  const handleContentChange = (val?: string) => {
    const newContent = val || '';
    setContent(newContent);
    setHasUnsavedChanges(true);
  };

  const handleBack = () => {
    if (hasUnsavedChanges) {
      if (!window.confirm(t('docsAdmin.editPage.unsavedConfirm'))) return;
    }
    navigate('/admin/docs');
  };

  // 获取编辑器 textarea 元素
  const getTextarea = useCallback((): HTMLTextAreaElement | null => {
    return editorRef.current?.querySelector('textarea') || null;
  }, []);

  // 上传图片并在光标位置插入 Markdown 图片语法
  const handleImageUpload = useCallback(async (file: File) => {
    if (!file.type.startsWith('image/')) {
      message.warning(t('docsAdmin.editPage.imageOnly'));
      return;
    }

    setIsUploading(true);
    try {
      const result = await uploadService.uploadImage(file);
      const markdown = `![${file.name}](${result.url})`;
      const textarea = getTextarea();

      if (textarea) {
        const start = textarea.selectionStart;
        const newContent =
          contentRef.current.substring(0, start) +
          '\n' + markdown + '\n' +
          contentRef.current.substring(start);
        setContent(newContent);
        setHasUnsavedChanges(true);

        setTimeout(() => {
          const pos = start + markdown.length + 2;
          textarea.focus();
          textarea.setSelectionRange(pos, pos);
        }, 0);
      } else {
        setContent((prev) => prev + '\n' + markdown + '\n');
        setHasUnsavedChanges(true);
      }
    } catch {
      message.error(t('docsAdmin.editPage.uploadFailed'));
    } finally {
      setIsUploading(false);
    }
  }, [t, getTextarea]);

  // 粘贴处理：拦截图片粘贴并上传到服务端
  const handlePaste = useCallback((e: React.ClipboardEvent) => {
    const items = Array.from(e.clipboardData.items);
    for (const item of items) {
      if (item.type.startsWith('image/')) {
        e.preventDefault();
        const file = item.getAsFile();
        if (file) handleImageUpload(file);
        break;
      }
    }
  }, [handleImageUpload]);

  // 文件选择器回调
  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    handleImageUpload(file);
    e.target.value = '';
  }, [handleImageUpload]);

  // 拖拽事件
  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!editorRef.current?.contains(e.relatedTarget as Node)) {
      setIsDragging(false);
    }
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    const file = e.dataTransfer.files[0];
    if (file) {
      handleImageUpload(file);
    }
  }, [handleImageUpload]);

  if (loading) {
    return (
      <div style={{ padding: 24 }}>
        <Skeleton active paragraph={{ rows: 20 }} />
      </div>
    );
  }

  const isPublished = form.getFieldValue('is_published');

  return (
    <div style={{ padding: 24 }} data-color-mode={themeMode === 'dark' ? 'dark' : 'light'}>
      <Card>
        {/* 顶部操作栏 */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <Space align="center">
            <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
              {t('docsAdmin.editPage.back')}
            </Button>
            <Title level={4} style={{ margin: 0 }}>
              {isEdit ? t('docsAdmin.editPage.title') : t('docsAdmin.editPage.createTitle')}
            </Title>
            {isPublished ? (
              <Tag color="success">{t('docsAdmin.table.published')}</Tag>
            ) : (
              <Tag>{t('docsAdmin.table.unpublished')}</Tag>
            )}
          </Space>

          <Space>
            {/* 保存状态指示 */}
            {isEdit && (
              <span style={{ fontSize: 12, color: 'var(--ant-color-text-secondary)' }}>
                {autoSaveStatus === 'saving' && (
                  <><LoadingOutlined style={{ marginRight: 4 }} />{t('docsAdmin.editPage.autoSaving')}</>
                )}
                {autoSaveStatus === 'saved' && (
                  <><CloudOutlined style={{ marginRight: 4, color: '#52c41a' }} />{t('docsAdmin.editPage.autoSaved')}</>
                )}
                {autoSaveStatus === 'idle' && lastSavedAt && (
                  <><CloudOutlined style={{ marginRight: 4 }} />{t('docsAdmin.editPage.lastSaved', { time: lastSavedAt.toLocaleTimeString('zh-CN') })}</>
                )}
                {hasUnsavedChanges && autoSaveStatus === 'idle' && (
                  <span style={{ color: '#faad14', marginLeft: 8 }}>● {t('docsAdmin.editPage.unsaved')}</span>
                )}
              </span>
            )}
            <Button
              icon={<SaveOutlined />}
              onClick={handleSave}
              loading={saving}
            >
              {t('docsAdmin.editPage.save')}（Ctrl+S）
            </Button>
            {!isPublished && (
              <Button
                type="primary"
                icon={<SendOutlined />}
                onClick={handlePublish}
                loading={saving}
              >
                {t('docsAdmin.editPage.publish')}
              </Button>
            )}
          </Space>
        </div>

        {/* 元信息表单 */}
        <Form
          form={form}
          layout="vertical"
          initialValues={{ is_published: false, sort_order: 0 }}
          onValuesChange={() => setHasUnsavedChanges(true)}
        >
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16, marginBottom: 16 }}>
            {!isEdit && (
              <Form.Item
                name="slug"
                label={t('docsAdmin.editPage.slugLabel')}
                rules={[
                  { required: true, message: t('docsAdmin.editPage.slugRequired') },
                  { pattern: /^[a-z0-9-]+$/, message: t('docsAdmin.editPage.slugPattern') },
                ]}
                style={{ marginBottom: 0 }}
              >
                <Input placeholder={t('docsAdmin.editPage.slugPlaceholder')} />
              </Form.Item>
            )}

            <Form.Item
              name="title"
              label={t('docsAdmin.editPage.titleLabel')}
              rules={[{ required: true, message: t('docsAdmin.editPage.titleRequired') }]}
              style={{ marginBottom: 0 }}
            >
              <Input placeholder={t('docsAdmin.editPage.titlePlaceholder')} />
            </Form.Item>

            <Form.Item name="icon" label={t('docsAdmin.editPage.iconLabel')} style={{ marginBottom: 0 }}>
              <Input placeholder={t('docsAdmin.editPage.iconPlaceholder')} />
            </Form.Item>

            <Form.Item name="sort_order" label={t('docsAdmin.editPage.sortOrderLabel')} style={{ marginBottom: 0 }}>
              <InputNumber min={0} style={{ width: '100%' }} placeholder={t('docsAdmin.editPage.sortOrderPlaceholder')} />
            </Form.Item>

            <Form.Item name="is_published" label={t('docsAdmin.editPage.publishStatus')} valuePropName="checked" style={{ marginBottom: 0 }}>
              <Switch checkedChildren={t('docsAdmin.editPage.published')} unCheckedChildren={t('docsAdmin.editPage.unpublished')} />
            </Form.Item>
          </div>

          <Form.Item name="subtitle" label={t('docsAdmin.editPage.subtitleLabel')} style={{ marginBottom: 16 }}>
            <Input.TextArea rows={2} placeholder={t('docsAdmin.editPage.subtitlePlaceholder')} />
          </Form.Item>
        </Form>

        {/* Markdown 编辑器 */}
        <div
          ref={editorRef}
          style={{ position: 'relative' }}
          onPaste={handlePaste}
          onDragEnter={handleDragEnter}
          onDragLeave={handleDragLeave}
          onDragOver={handleDragOver}
          onDrop={handleDrop}
        >
          <MDEditor
            value={content}
            onChange={handleContentChange}
            height={600}
            preview="live"
            previewOptions={{
              rehypePlugins: [[rehypeSanitize]],
            }}
            textareaProps={{
              placeholder: t('docsAdmin.editPage.contentPlaceholder'),
            }}
          />

          {/* 拖拽上传遮罩 */}
          {isDragging && (
            <div style={{
              position: 'absolute',
              inset: 0,
              zIndex: 10,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: 'rgba(0, 217, 255, 0.08)',
              backdropFilter: 'blur(2px)',
              border: '2px dashed #00D9FF',
              borderRadius: 8,
            }}>
              <div style={{ textAlign: 'center', color: '#00D9FF' }}>
                <PictureOutlined style={{ fontSize: 32, marginBottom: 8, display: 'block' }} />
                <span style={{ fontWeight: 500 }}>{t('docsAdmin.editPage.dropImageHere')}</span>
              </div>
            </div>
          )}

          {/* 上传进度遮罩 */}
          {isUploading && (
            <div style={{
              position: 'absolute',
              inset: 0,
              zIndex: 20,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: 'rgba(0, 0, 0, 0.3)',
              backdropFilter: 'blur(2px)',
              borderRadius: 8,
            }}>
              <div style={{
                background: 'var(--ant-color-bg-container)',
                borderRadius: 16,
                padding: '24px 32px',
                textAlign: 'center',
                boxShadow: '0 8px 32px rgba(0,0,0,0.15)',
              }}>
                <LoadingOutlined style={{ fontSize: 28, color: '#00D9FF', marginBottom: 12, display: 'block' }} />
                <span style={{ fontSize: 14, color: 'var(--ant-color-text)' }}>
                  {t('docsAdmin.editPage.uploading')}
                </span>
              </div>
            </div>
          )}
        </div>

        {/* 底部信息栏 */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginTop: 8,
          padding: '8px 0',
          fontSize: 12,
          color: 'var(--ant-color-text-secondary)',
        }}>
          <span>
            Markdown · {content.length} {t('docsAdmin.editPage.chars')}
          </span>
          <Space size={12}>
            <Tooltip title={t('docsAdmin.editPage.insertImage')}>
              <Button
                type="text"
                size="small"
                icon={<PictureOutlined />}
                onClick={() => imageInputRef.current?.click()}
                style={{ fontSize: 12, color: 'var(--ant-color-text-secondary)' }}
              >
                {t('docsAdmin.editPage.insertImage')}
              </Button>
            </Tooltip>
            <span>{t('docsAdmin.editPage.shortcutHint')}</span>
          </Space>
        </div>

        {/* 隐藏的图片选择器 */}
        <input
          ref={imageInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileChange}
          style={{ display: 'none' }}
        />
      </Card>
    </div>
  );
};

export default DocsEditPage;
