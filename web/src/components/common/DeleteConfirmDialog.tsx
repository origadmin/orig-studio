import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogCancel,
} from '@/components/ui/alert-dialog';
import {Button} from '@/components/ui/button';

export interface DeleteConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  isDeleting: boolean;
  onConfirm: () => void;
  /** Custom description text. Defaults to delete-specific message. */
  description?: string;
  /** Custom confirm button label. Defaults to "确认删除". */
  confirmLabel?: string;
  /** Confirm button variant. Defaults to "destructive". */
  confirmVariant?: 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
  /** Loading state label. Defaults to "删除中...". */
  loadingLabel?: string;
}

export function DeleteConfirmDialog({
  open,
  onOpenChange,
  title,
  isDeleting,
  onConfirm,
  description,
  confirmLabel,
  confirmVariant,
  loadingLabel,
}: DeleteConfirmDialogProps) {
  const isCustomDialog = !!(description || confirmLabel);
  const variant = confirmVariant || (isCustomDialog ? 'default' : 'destructive');
  const confirmText = confirmLabel || '确认删除';
  const loadingText = loadingLabel || (isCustomDialog ? '处理中...' : '删除中...');
  const descText = description || `确定要删除 "${title}" 吗？此操作不可撤销。`;

  return (
    <AlertDialog open={open} onOpenChange={isDeleting ? undefined : onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{isCustomDialog ? title : '确认删除'}</AlertDialogTitle>
          <AlertDialogDescription>
            {descText}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>取消</AlertDialogCancel>
          <Button
            variant={variant}
            onClick={onConfirm}
            disabled={isDeleting}
          >
            {isDeleting ? loadingText : confirmText}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
